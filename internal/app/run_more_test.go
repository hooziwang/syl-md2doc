package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunPandocUnavailable(t *testing.T) {
	tmp := t.TempDir()
	res, err := Run(Options{
		Inputs:     []string{"a.md"},
		CWD:        tmp,
		PandocPath: filepath.Join(tmp, "missing-pandoc"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pandoc")
	require.Equal(t, 0, res.SuccessCount)
}

func TestRunNoMarkdownWarn(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("x"), 0o644))

	res, err := Run(Options{
		Inputs:    []string{tmp},
		CWD:       tmp,
		Converter: &stubConverter{},
		Jobs:      0,
	})
	require.NoError(t, err)
	require.Equal(t, 0, res.SuccessCount)
	require.Equal(t, 0, res.FailureCount)
	require.NotEmpty(t, res.Warnings)
}
