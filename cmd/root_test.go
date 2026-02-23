package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeArgsInsertBuildForDirectRun(t *testing.T) {
	got := normalizeArgs([]string{"a.md", "-o", "out"})
	require.Equal(t, []string{"a.md", "-o", "out"}, got)
}

func TestNormalizeArgsKeepSubcommand(t *testing.T) {
	got := normalizeArgs([]string{"version"})
	require.Equal(t, []string{"--version"}, got)
}

func TestBuildRequiresInput(t *testing.T) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.ErrorIs(t, err, errBuildFailed)
	require.Contains(t, stderr.String(), "\"event\":\"invalid_input\"")
	require.Contains(t, stderr.String(), "至少一个 .md 文件或目录")
}
