package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionText(t *testing.T) {
	oldVersion, oldCommit, oldBuildTime := Version, Commit, BuildTime
	defer func() {
		Version, Commit, BuildTime = oldVersion, oldCommit, oldBuildTime
	}()

	Version = "v1.2.3"
	Commit = "abc1234"
	BuildTime = "2026-02-23T10:00:00Z"

	require.Equal(t, "syl-md2doc 版本：v1.2.3（commit: abc1234，构建时间: 2026-02-23T10:00:00Z）", versionText())
}

func TestPrintVersion(t *testing.T) {
	oldVersion, oldCommit, oldBuildTime := Version, Commit, BuildTime
	defer func() {
		Version, Commit, BuildTime = oldVersion, oldCommit, oldBuildTime
	}()

	Version = "v9.9.9"
	Commit = "def5678"
	BuildTime = "2026-02-23T11:00:00Z"

	banner := strings.TrimSpace(loveBanner(io.Discard))
	require.NotEmpty(t, banner)

	buf := bytes.NewBuffer(nil)
	printVersion(buf)
	out := buf.String()
	require.Contains(t, out, "syl-md2doc 版本：v9.9.9（commit: def5678，构建时间: 2026-02-23T11:00:00Z）")
	require.Contains(t, out, banner)
}
