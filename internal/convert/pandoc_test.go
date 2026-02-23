package convert

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/job"
)

func TestEnsurePandocAvailableMissingBinary(t *testing.T) {
	err := EnsurePandocAvailable(filepath.Join(t.TempDir(), "missing-pandoc"))
	require.Error(t, err)
}

func TestPandocConverterBuildCommand(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()

	var gotName string
	var gotArgs []string
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string{}, args...)
		return exec.CommandContext(ctx, "sh", "-c", "exit 0")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	dst := filepath.Join(tmp, "a.docx")
	require.NoError(t, os.WriteFile(src, []byte("# a"), 0o644))

	conv := NewPandocConverter("pandoc-x", filepath.Join(tmp, "ref.docx"), false)
	res := conv.Convert(context.Background(), job.Task{SourcePath: src, TargetPath: dst})
	require.NoError(t, res.Error)
	require.Equal(t, "pandoc-x", gotName)
	require.Contains(t, gotArgs, "-f")
	require.Contains(t, gotArgs, "gfm")
	require.Contains(t, gotArgs, "--reference-doc="+filepath.Join(tmp, "ref.docx"))
}

func TestPandocConverterMissingImageAsWarningWhenOutputExists(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'Could not fetch resource x.png' 1>&2; exit 1")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	dst := filepath.Join(tmp, "a.docx")
	require.NoError(t, os.WriteFile(src, []byte("![x](x.png)"), 0o644))
	require.NoError(t, os.WriteFile(dst, []byte("dummy"), 0o644))

	conv := NewPandocConverter("pandoc", "", false)
	res := conv.Convert(context.Background(), job.Task{SourcePath: src, TargetPath: dst})
	require.NoError(t, res.Error)
	require.NotEmpty(t, res.Warnings)
}
