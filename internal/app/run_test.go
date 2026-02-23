package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/job"
)

type stubConverter struct{}

func (s *stubConverter) Convert(ctx context.Context, task job.Task) job.Result {
	if filepath.Base(task.SourcePath) == "bad.md" {
		return job.Result{Task: task, Error: fmt.Errorf("boom")}
	}
	return job.Result{Task: task, Warnings: []string{"ok"}}
}

func TestRunCollectSummaryAndFailures(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.md"), []byte("# a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "bad.md"), []byte("# b"), 0o644))

	res, err := Run(Options{
		Inputs:    []string{"a.md", "bad.md", "missing.md"},
		CWD:       tmp,
		Converter: &stubConverter{},
		Jobs:      2,
	})
	require.NoError(t, err)
	require.Equal(t, 1, res.SuccessCount)
	require.Equal(t, 2, res.FailureCount)
	require.NotEmpty(t, res.Warnings)
}

func TestRunRequiresInput(t *testing.T) {
	_, err := Run(Options{})
	require.Error(t, err)
}
