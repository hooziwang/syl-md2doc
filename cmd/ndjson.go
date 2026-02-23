package cmd

import (
	"encoding/json"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ndjsonEvent struct {
	Timestamp  string         `json:"timestamp"`
	Level      string         `json:"level"`
	Event      string         `json:"event"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	Suggestion string         `json:"suggestion,omitempty"`
}

func emitNDJSON(w io.Writer, level, event, message string, details map[string]any, suggestion string) {
	if w == nil {
		return
	}
	e := ndjsonEvent{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Level:      level,
		Event:      event,
		Message:    message,
		Details:    details,
		Suggestion: suggestion,
	}
	buf, err := json.Marshal(e)
	if err != nil {
		fallback, _ := json.Marshal(ndjsonEvent{
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Level:      "error",
			Event:      "logger_error",
			Message:    "NDJSON 序列化失败",
			Details:    map[string]any{"reason": err.Error()},
			Suggestion: "检查日志字段是否包含无法序列化的数据结构",
		})
		_, _ = w.Write(append(fallback, '\n'))
		return
	}
	_, _ = w.Write(append(buf, '\n'))
}

func absPath(cwd, p string) string {
	if strings.TrimSpace(p) == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	if strings.TrimSpace(cwd) == "" {
		if wd, err := filepath.Abs(p); err == nil {
			return filepath.Clean(wd)
		}
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(cwd, p))
}

func absPaths(cwd string, paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, absPath(cwd, p))
	}
	return out
}

func suggestionForFailure(reason string) string {
	lower := strings.ToLower(reason)
	switch {
	case strings.Contains(reason, "输入不存在"):
		return "检查输入路径是否存在且可读；建议使用绝对路径重新执行"
	case strings.Contains(reason, "创建输出目录失败"):
		return "检查输出目录权限，或切换到有写权限的目录后重试"
	case strings.Contains(reason, "pandoc 转换失败"):
		if strings.Contains(lower, "could not fetch resource") || strings.Contains(lower, "image not found") {
			return "补齐 Markdown 引用的本地资源文件，或改为可访问路径；然后重试"
		}
		return "建议先手工执行 pandoc 命令定位具体语法/资源问题，再修复 Markdown 后重试"
	default:
		return "检查错误详情与输入文件内容；确认路径、权限和依赖环境后重试"
	}
}

func suggestionForTopError(errText string) string {
	lower := strings.ToLower(errText)
	switch {
	case strings.Contains(errText, "至少提供一个输入"):
		return "至少传入一个 .md 文件或目录，例如：syl-md2doc /abs/path/a.md /abs/path/docs"
	case strings.Contains(errText, "未找到 pandoc"):
		switch runtime.GOOS {
		case "darwin":
			return "先执行 brew install pandoc；若已安装但不在 PATH，使用 --pandoc-path 指定绝对路径"
		case "windows":
			return "先执行 scoop install pandoc（或 choco install pandoc）；若 PATH 未生效，使用 --pandoc-path"
		default:
			return "先执行 sudo apt-get install pandoc（或系统包管理器安装）；也可使用 --pandoc-path 指定"
		}
	case strings.Contains(errText, "版本过低"):
		return "升级 pandoc 到 >= 2.19.0 后重试；可用 pandoc --version 确认版本"
	case strings.Contains(lower, "permission denied"):
		return "检查文件读写权限，确保输入可读、输出目录可写"
	default:
		return "根据 details 中的错误信息逐项排查；优先检查路径、依赖和权限"
	}
}

func EmitUnhandledError(w io.Writer, err error) {
	if err == nil {
		return
	}
	emitNDJSON(w, "error", "fatal_error", "程序执行失败", map[string]any{
		"error": err.Error(),
	}, suggestionForTopError(err.Error()))
}
