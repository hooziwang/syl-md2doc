package convert

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"syl-md2doc/internal/job"
)

var execCommandContext = exec.CommandContext

type PandocConverter struct {
	PandocPath    string
	ReferenceDocx string
	Verbose       bool
}

func NewPandocConverter(pandocPath, referenceDocx string, verbose bool) *PandocConverter {
	return &PandocConverter{PandocPath: pandocPath, ReferenceDocx: referenceDocx, Verbose: verbose}
}

func EnsurePandocAvailable(pandocPath string) error {
	bin := strings.TrimSpace(pandocPath)
	if bin == "" {
		bin = "pandoc"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("未找到 pandoc，可使用 --pandoc-path 指定路径，或先安装 pandoc")
	}
	return nil
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
	args := []string{task.SourcePath, "-f", "gfm", "-t", "docx", "-o", task.TargetPath}
	if strings.TrimSpace(p.ReferenceDocx) != "" {
		args = append(args, "--reference-doc="+p.ReferenceDocx)
	}

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
