package convert

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"syl-md2doc/internal/job"
)

var execCommandContext = exec.CommandContext
var execLookPath = exec.LookPath

//go:embed templates/default-reference.docx
var defaultReferenceDocx []byte

type PandocInfo struct {
	BinaryPath string
	Version    string
}

type PandocConverter struct {
	PandocPath    string
	ReferenceDocx string
	Verbose       bool
}

func NewPandocConverter(pandocPath, referenceDocx string, verbose bool) *PandocConverter {
	return &PandocConverter{PandocPath: pandocPath, ReferenceDocx: referenceDocx, Verbose: verbose}
}

func EnsurePandocAvailable(pandocPath string) (PandocInfo, error) {
	bin := strings.TrimSpace(pandocPath)
	if bin == "" {
		bin = "pandoc"
	}
	resolved, err := execLookPath(bin)
	if err != nil {
		return PandocInfo{}, fmt.Errorf("未找到 pandoc（%s）。%s；也可使用 --pandoc-path 指定路径", bin, installHint(runtime.GOOS))
	}

	version, err := detectPandocVersion(resolved)
	if err != nil {
		// 不把版本解析失败当成阻断错误，保证跨平台兼容性。
		version = ""
	}
	return PandocInfo{BinaryPath: resolved, Version: version}, nil
}

func (p *PandocConverter) Convert(ctx context.Context, task job.Task) job.Result {
	res := job.Result{Task: task, Warnings: make([]string, 0)}
	if err := os.MkdirAll(filepath.Dir(task.TargetPath), 0o755); err != nil {
		res.Error = fmt.Errorf("创建输出目录失败：%w", err)
		return res
	}

	bin := strings.TrimSpace(p.PandocPath)
	if bin == "" {
		bin = "pandoc"
	}
	refPath := strings.TrimSpace(p.ReferenceDocx)
	if refPath == "" {
		tmpRef, err := materializeDefaultReferenceDocx()
		if err != nil {
			res.Error = fmt.Errorf("准备内置 reference-docx 失败：%w", err)
			return res
		}
		refPath = tmpRef
		defer func() {
			_ = os.Remove(tmpRef)
		}()
	}
	args := []string{task.SourcePath, "-f", "gfm", "-t", "docx", "-o", task.TargetPath}
	args = append(args, "--reference-doc="+refPath)

	cmd := execCommandContext(ctx, bin, args...)
	stderr := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	if p.Verbose {
		cmd.Stdout = os.Stdout
	}

	err := cmd.Run()
	stderrText := strings.TrimSpace(stderr.String())
	res.Warnings = append(res.Warnings, collectWarnings(stderrText)...)

	if err != nil {
		if isMissingAssetOnly(stderrText) {
			if _, stErr := os.Stat(task.TargetPath); stErr == nil {
				res.Warnings = append(res.Warnings, "检测到缺失资源，已忽略并继续")
				return res
			}
		}
		reason := stderrText
		if reason == "" {
			reason = err.Error()
		}
		res.Error = fmt.Errorf("pandoc 转换失败：%s", reason)
	}
	return res
}

func collectWarnings(stderrText string) []string {
	if strings.TrimSpace(stderrText) == "" {
		return nil
	}
	lines := strings.Split(stderrText, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "warning") || looksLikeMissingAsset(line) {
			out = append(out, line)
		}
	}
	return out
}

func isMissingAssetOnly(stderrText string) bool {
	if strings.TrimSpace(stderrText) == "" {
		return false
	}
	return looksLikeMissingAsset(stderrText)
}

func looksLikeMissingAsset(text string) bool {
	lower := strings.ToLower(text)
	patterns := []string{
		"could not fetch resource",
		"resource not found",
		"could not find image",
		"cannot find image",
		"image not found",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func detectPandocVersion(binPath string) (string, error) {
	cmd := execCommandContext(context.Background(), binPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行 pandoc --version 失败：%w", err)
	}
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	if line == "" {
		return "", fmt.Errorf("读取 pandoc 版本失败：输出为空")
	}
	ver, ok := extractVersionToken(line)
	if !ok {
		return "", fmt.Errorf("无法识别 pandoc 版本：%s", line)
	}
	return ver, nil
}

func extractVersionToken(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", false
	}
	raw := strings.TrimPrefix(fields[1], "v")
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return "", false
	}
	if !isDigits(parts[0]) || !isDigits(parts[1]) {
		return "", false
	}
	if len(parts) == 2 {
		return parts[0] + "." + parts[1] + ".0", true
	}
	patch := leadingDigits(parts[2])
	if patch == "" {
		return "", false
	}
	return parts[0] + "." + parts[1] + "." + patch, true
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func leadingDigits(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func installHint(goos string) string {
	switch goos {
	case "darwin":
		return "可执行：brew install pandoc"
	case "windows":
		return "可执行：scoop install pandoc（或 choco install pandoc）"
	default:
		return "可执行：sudo apt-get install pandoc（或使用系统包管理器安装）"
	}
}

func materializeDefaultReferenceDocx() (string, error) {
	if len(defaultReferenceDocx) == 0 {
		return "", fmt.Errorf("内置 reference-docx 为空")
	}
	f, err := os.CreateTemp("", "syl-md2doc-reference-*.docx")
	if err != nil {
		return "", fmt.Errorf("创建临时 reference-docx 失败：%w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := f.Write(defaultReferenceDocx); err != nil {
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("写入临时 reference-docx 失败：%w", err)
	}
	return f.Name(), nil
}
