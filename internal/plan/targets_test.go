package plan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/input"
)

func TestBuildTargetsDirInputPreserveRelativePath(t *testing.T) {
	tmp := t.TempDir()
	sources := []input.SourceItem{
		{SourcePath: filepath.Join(tmp, "src", "a", "x.md"), FromDir: true, RelPath: filepath.Join("a", "x.md")},
	}
	tasks, warns, err := BuildTargets(sources, Options{CWD: tmp, OutputArg: filepath.Join(tmp, "out")})
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, filepath.Join(tmp, "out", "a", "x.docx"), tasks[0].TargetPath)
}

func TestBuildTargetsConflictAppendSuffix(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.docx"), []byte("x"), 0o644))
	sources := []input.SourceItem{{SourcePath: filepath.Join(tmp, "a.md")}}
	tasks, _, err := BuildTargets(sources, Options{CWD: tmp})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmp, "a_1.docx"), tasks[0].TargetPath)
}

func TestBuildTargetsMultiInputOutputFileFallbackToDirWarn(t *testing.T) {
	tmp := t.TempDir()
	sources := []input.SourceItem{{SourcePath: filepath.Join(tmp, "a.md")}, {SourcePath: filepath.Join(tmp, "b.md")}}
	tasks, warns, err := BuildTargets(sources, Options{CWD: tmp, OutputArg: filepath.Join(tmp, "x.docx")})
	require.NoError(t, err)
	require.NotEmpty(t, warns)
	require.Equal(t, filepath.Join(tmp, "a.docx"), tasks[0].TargetPath)
	require.Equal(t, filepath.Join(tmp, "b.docx"), tasks[1].TargetPath)
}
