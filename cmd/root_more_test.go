package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildWithVersionFlag(t *testing.T) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "syl-md2doc 版本：")
}

func TestDirectRunNoInputReturnsError(t *testing.T) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.ErrorIs(t, err, errBuildFailed)
	require.Contains(t, stderr.String(), "\"event\":\"invalid_input\"")
}

func TestBuildSuccessWithFakePandoc(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	require.NoError(t, os.WriteFile(src, []byte("# hi"), 0o644))

	pandoc := filepath.Join(tmp, "fake-pandoc.sh")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'pandoc 3.1.11'; exit 0; fi\nout=\"\"\nwhile [ $# -gt 0 ]; do\n  if [ \"$1\" = \"-o\" ]; then out=\"$2\"; shift 2; continue; fi\n  shift\ndone\nmkdir -p \"$(dirname \"$out\")\"\nprintf 'ok' > \"$out\"\nexit 0\n"
	require.NoError(t, os.WriteFile(pandoc, []byte(script), 0o755))

	outDir := filepath.Join(tmp, "out")
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{src, "--pandoc-path", pandoc, "--output", outDir})

	err := cmd.Execute()
	require.NoError(t, err)
	require.NotContains(t, stdout.String(), "\"event\":\"build_start\"")
	require.Contains(t, stdout.String(), "\"event\":\"summary\"")
	require.Contains(t, stdout.String(), "\"success_count\":1")
	require.Contains(t, stdout.String(), "\"failure_count\":0")
	require.NotContains(t, stdout.String(), "\"pandoc_path\"")
	matches, gErr := filepath.Glob(filepath.Join(outDir, "a_*.docx"))
	require.NoError(t, gErr)
	require.Len(t, matches, 1)
}

func TestBuildVerbosePrintPandocInfo(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	require.NoError(t, os.WriteFile(src, []byte("# hi"), 0o644))

	pandoc := filepath.Join(tmp, "fake-pandoc-version.sh")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'pandoc 3.1.11'; exit 0; fi\nout=\"\"\nwhile [ $# -gt 0 ]; do\n  if [ \"$1\" = \"-o\" ]; then out=\"$2\"; shift 2; continue; fi\n  shift\ndone\nmkdir -p \"$(dirname \"$out\")\"\nprintf 'ok' > \"$out\"\nexit 0\n"
	require.NoError(t, os.WriteFile(pandoc, []byte(script), 0o755))

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{src, "--pandoc-path", pandoc, "--verbose"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "\"event\":\"pandoc_environment\"")
	require.Contains(t, stdout.String(), "\"pandoc_version\":\"3.1.11\"")
}

func TestNormalizeArgsFlagsOnlyNoBuildInjected(t *testing.T) {
	got := normalizeArgs([]string{"--jobs", "4", "--verbose"})
	require.Equal(t, []string{"--jobs", "4", "--verbose"}, got)
}

func TestNormalizeArgsVersionAlias(t *testing.T) {
	got := normalizeArgs([]string{"version"})
	require.Equal(t, []string{"--version"}, got)
}

func TestBuildWithWarningAndFailureSummary(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	require.NoError(t, os.WriteFile(src, []byte("# hi"), 0o644))

	pandoc := filepath.Join(tmp, "fake-pandoc-fail.sh")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'pandoc 3.1.11'; exit 0; fi\necho 'fatal: failed' 1>&2\nexit 1\n"
	require.NoError(t, os.WriteFile(pandoc, []byte(script), 0o755))

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{src, "--pandoc-path", pandoc})

	err := cmd.Execute()
	require.ErrorIs(t, err, errBuildFailed)
	require.Contains(t, stdout.String(), "\"event\":\"summary\"")
	require.Contains(t, stdout.String(), "\"failure_count\":1")
	require.Contains(t, stdout.String(), "\"suggestion\":\"修复失败项后重试")
	require.Contains(t, stdout.String(), "\"pandoc_path\"")
	require.True(t, strings.Contains(stderr.String(), "\"event\":\"file_failed\""))
}

func TestBuildWithHighlightWordsFlag(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.md")
	require.NoError(t, os.WriteFile(src, []byte("# hi"), 0o644))

	pandoc := filepath.Join(tmp, "fake-pandoc-highlight.sh")
	script := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'pandoc 3.1.11'; exit 0; fi\nout=\"\"\nhas_filter=0\nwhile [ $# -gt 0 ]; do\n  case \"$1\" in\n    -o) out=\"$2\"; shift 2; continue;;\n    --lua-filter=*) has_filter=1; shift; continue;;\n  esac\n  shift\ndone\nif [ \"$has_filter\" -ne 1 ]; then echo 'missing lua filter' 1>&2; exit 2; fi\nmkdir -p \"$(dirname \"$out\")\"\nprintf 'ok' > \"$out\"\nexit 0\n"
	require.NoError(t, os.WriteFile(pandoc, []byte(script), 0o755))

	outDir := filepath.Join(tmp, "out")
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{src, "--pandoc-path", pandoc, "--output", outDir, "-w", "paper,lanterns"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "\"success_count\":1")
}

func TestHelpContainsDetailedGuidance(t *testing.T) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)
	out := stdout.String()
	require.Contains(t, out, "输入规则：")
	require.Contains(t, out, "输出规则：")
	require.Contains(t, out, "Examples:")
	require.Contains(t, out, "syl-md2doc /abs/docs/a.md")
	require.Contains(t, out, "syl-md2doc --version")
}
