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

func TestEnsurePandocAvailableMissingBinary(t *testing.T) {
	_, err := EnsurePandocAvailable(filepath.Join(t.TempDir(), "missing-pandoc"))
	require.Error(t, err)
}

func TestEnsurePandocAvailableSuccess(t *testing.T) {
	tmp := t.TempDir()
	fake := filepath.Join(tmp, "fake-pandoc.sh")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'pandoc 3.1.11'; exit 0; fi\nexit 0\n"
	require.NoError(t, os.WriteFile(fake, []byte(script), 0o755))

	info, err := EnsurePandocAvailable(fake)
	require.NoError(t, err)
	require.Equal(t, fake, info.BinaryPath)
	require.Equal(t, "3.1.11", info.Version)
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
	require.Contains(t, gotArgs, "gfm+raw_attribute+hard_line_breaks")
	require.Contains(t, gotArgs, "--reference-doc="+filepath.Join(tmp, "ref.docx"))
}

func TestPandocConverterPreservesBlankLines(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()

	var gotSourcePath string
	var gotSourceContent string
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if len(args) > 0 {
			gotSourcePath = args[0]
			bs, err := os.ReadFile(gotSourcePath)
			require.NoError(t, err)
			gotSourceContent = string(bs)
		}
		return exec.CommandContext(ctx, "sh", "-c", "exit 0")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	dst := filepath.Join(tmp, "a.docx")
	require.NoError(t, os.WriteFile(src, []byte("line1\n\nline2\n"), 0o644))

	conv := NewPandocConverter("pandoc-x", filepath.Join(tmp, "ref.docx"), false)
	res := conv.Convert(context.Background(), job.Task{SourcePath: src, TargetPath: dst})
	require.NoError(t, res.Error)

	require.NotEqual(t, src, gotSourcePath)
	require.Contains(t, gotSourceContent, "```{=openxml}\n<w:p/>\n```")
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

func TestPandocConverterUsesEmbeddedDefaultReferenceDocx(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()

	var gotReferencePath string
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		for _, arg := range args {
			if strings.HasPrefix(arg, "--reference-doc=") {
				gotReferencePath = strings.TrimPrefix(arg, "--reference-doc=")
			}
		}
		return exec.CommandContext(ctx, "sh", "-c", "exit 0")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	dst := filepath.Join(tmp, "a.docx")
	require.NoError(t, os.WriteFile(src, []byte("# a"), 0o644))

	conv := NewPandocConverter("pandoc", "", false)
	res := conv.Convert(context.Background(), job.Task{SourcePath: src, TargetPath: dst})
	require.NoError(t, res.Error)
	require.NotEmpty(t, gotReferencePath)

	_, err := os.Stat(gotReferencePath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestPandocConverterBuildCommandIncludesLuaFilter(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()

	var gotArgs []string
	execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
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

	luaFilter := ""
	for _, a := range gotArgs {
		if strings.HasPrefix(a, "--lua-filter=") {
			luaFilter = strings.TrimPrefix(a, "--lua-filter=")
			break
		}
	}
	require.NotEmpty(t, luaFilter)
	_, err := os.Stat(luaFilter)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}
