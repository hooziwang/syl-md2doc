package convert

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/job"
)

func TestEnsurePandocAvailableVersionTooLowStillPass(t *testing.T) {
	tmp := t.TempDir()
	fake := filepath.Join(tmp, "fake-pandoc-low.sh")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'pandoc 2.17'; exit 0; fi\nexit 0\n"
	require.NoError(t, os.WriteFile(fake, []byte(script), 0o755))

	info, err := EnsurePandocAvailable(fake)
	require.NoError(t, err)
	require.Equal(t, fake, info.BinaryPath)
	require.Equal(t, "2.17.0", info.Version)
}

func TestPandocConverterNonMissingAssetError(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'fatal parser error' 1>&2; exit 1")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	dst := filepath.Join(tmp, "a.docx")
	require.NoError(t, os.WriteFile(src, []byte("# x"), 0o644))

	res := NewPandocConverter("pandoc", "", false).Convert(context.Background(), job.Task{SourcePath: src, TargetPath: dst})
	require.Error(t, res.Error)
}

func TestPandocConverterMissingAssetWithoutOutputStillError(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'Could not fetch resource lost.png' 1>&2; exit 1")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	dst := filepath.Join(tmp, "missing", "a.docx")
	require.NoError(t, os.WriteFile(src, []byte("![x](lost.png)"), 0o644))

	res := NewPandocConverter("pandoc", "", false).Convert(context.Background(), job.Task{SourcePath: src, TargetPath: dst})
	require.Error(t, res.Error)
}

func TestPandocConverterMkdirFail(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "exit 0")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	require.NoError(t, os.WriteFile(src, []byte("# x"), 0o644))

	res := NewPandocConverter("pandoc", "", false).Convert(context.Background(), job.Task{SourcePath: src, TargetPath: filepath.Join("/dev/null", "a.docx")})
	require.Error(t, res.Error)
}

func TestPreserveMarkdownBlankLines(t *testing.T) {
	out, changed := preserveMarkdownBlankLines("line1\n\nline2\n")
	require.True(t, changed)
	require.Contains(t, out, "```{=openxml}\n<w:p/>\n```")
	require.True(t, strings.HasSuffix(out, "\n"))
}

func TestPreserveMarkdownBlankLinesSkipFencedCode(t *testing.T) {
	in := "```go\n\nx := 1\n```\n"
	out, changed := preserveMarkdownBlankLines(in)
	require.False(t, changed)
	require.Equal(t, in, out)
}

func TestBuildHighlightLuaFilterIncludesStrongRule(t *testing.T) {
	script := buildHighlightLuaFilter()
	require.Contains(t, script, "function Strong(el)")
	require.Contains(t, script, "pandoc.Strong")
	require.Contains(t, script, "KeywordHighlight")
}
