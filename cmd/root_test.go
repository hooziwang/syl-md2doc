package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeArgsInsertBuildForDirectRun(t *testing.T) {
	got := normalizeArgs([]string{"a.md", "-o", "out"})
	require.Equal(t, []string{"build", "a.md", "-o", "out"}, got)
}

func TestNormalizeArgsKeepSubcommand(t *testing.T) {
	got := normalizeArgs([]string{"version"})
	require.Equal(t, []string{"version"}, got)
}

func TestBuildRequiresInput(t *testing.T) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{"build"})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "至少提供一个输入")
}
