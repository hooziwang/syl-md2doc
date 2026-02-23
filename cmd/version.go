package cmd

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/hooziwang/daddylovesyl"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

func versionText() string {
	return fmt.Sprintf("syl-md2doc 版本：%s（commit: %s，构建时间: %s）", Version, Commit, BuildTime)
}

func loveBanner(w io.Writer) string {
	return daddylovesyl.Render(w)
}

func printVersion(w io.Writer) {
	emitNDJSON(w, "info", "version_info", "版本信息", map[string]any{
		"tool":       "syl-md2doc",
		"version":    Version,
		"commit":     Commit,
		"build_time": BuildTime,
		"banner":     loveBanner(io.Discard),
		"executable": filepath.Base("syl-md2doc"),
		"text":       versionText(),
	}, "")
}
