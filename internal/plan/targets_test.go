package plan

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/input"
)

func TestBuildTargetsDirInputPreserveRelativePath(t *testing.T) {
	tmp := t.TempDir()
	oldGen := codeGenerator
	codeGenerator = func(n int) string { return "AbC123" }
	defer func() { codeGenerator = oldGen }()

	sources := []input.SourceItem{
		{SourcePath: filepath.Join(tmp, "src", "a", "x.md"), FromDir: true, RelPath: filepath.Join("a", "x.md")},
	}
	tasks, warns, err := BuildTargets(sources, Options{CWD: tmp, OutputArg: filepath.Join(tmp, "out")})
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, filepath.Join(tmp, "out", "a", "x_AbC123.docx"), tasks[0].TargetPath)
}

func TestBuildTargetsConflictRegenerateCode(t *testing.T) {
	tmp := t.TempDir()
	oldGen := codeGenerator
	codes := []string{"AAAAAA", "BBBBBB"}
	idx := 0
	codeGenerator = func(n int) string {
		if idx >= len(codes) {
			return "CCCCCC"
		}
		c := codes[idx]
		idx++
		return c
	}
	defer func() { codeGenerator = oldGen }()

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a_AAAAAA.docx"), []byte("x"), 0o644))
	sources := []input.SourceItem{{SourcePath: filepath.Join(tmp, "a.md")}}
	tasks, _, err := BuildTargets(sources, Options{CWD: tmp})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmp, "a_BBBBBB.docx"), tasks[0].TargetPath)
}

func TestBuildTargetsMultiInputOutputFileFallbackToDirWarn(t *testing.T) {
	tmp := t.TempDir()
	oldGen := codeGenerator
	codes := []string{"AA11BB", "CC22DD"}
	idx := 0
	codeGenerator = func(n int) string {
		c := codes[idx]
		idx++
		return c
	}
	defer func() { codeGenerator = oldGen }()

	sources := []input.SourceItem{{SourcePath: filepath.Join(tmp, "a.md")}, {SourcePath: filepath.Join(tmp, "b.md")}}
	tasks, warns, err := BuildTargets(sources, Options{CWD: tmp, OutputArg: filepath.Join(tmp, "x.docx")})
	require.NoError(t, err)
	require.NotEmpty(t, warns)
	require.Equal(t, filepath.Join(tmp, "a_AA11BB.docx"), tasks[0].TargetPath)
	require.Equal(t, filepath.Join(tmp, "b_CC22DD.docx"), tasks[1].TargetPath)
}

func TestRandomCodePattern(t *testing.T) {
	code := randomCode(6)
	require.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9]{6}$`), code)
}
