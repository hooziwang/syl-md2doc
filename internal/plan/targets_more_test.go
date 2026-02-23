package plan

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/input"
)

func TestBuildTargetsSingleInputOutputFile(t *testing.T) {
	tmp := t.TempDir()
	sources := []input.SourceItem{{SourcePath: filepath.Join(tmp, "a.md")}}
	fixed := filepath.Join(tmp, "x.docx")
	tasks, warns, err := BuildTargets(sources, Options{CWD: tmp, OutputArg: fixed})
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, fixed, tasks[0].TargetPath)
}

func TestReplaceExtNoExt(t *testing.T) {
	require.Equal(t, "a.docx", replaceExt("a", ".docx"))
}
