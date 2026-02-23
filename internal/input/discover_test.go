package input

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverMixedFileAndDirRecursiveOnlyMD(t *testing.T) {
	tmp := t.TempDir()
	dirA := filepath.Join(tmp, "a")
	require.NoError(t, os.MkdirAll(filepath.Join(dirA, "sub"), 0o755))

	md1 := filepath.Join(dirA, "sub", "a.md")
	require.NoError(t, os.WriteFile(md1, []byte("# a"), 0o644))
	txt1 := filepath.Join(dirA, "sub", "a.txt")
	require.NoError(t, os.WriteFile(txt1, []byte("x"), 0o644))
	md2 := filepath.Join(tmp, "root.md")
	require.NoError(t, os.WriteFile(md2, []byte("# b"), 0o644))

	items, warns, fails, err := Discover([]string{dirA, md2}, tmp)
	require.NoError(t, err)
	require.Len(t, fails, 0)
	require.Len(t, items, 2)
	require.NotEmpty(t, warns)
}

func TestDiscoverMissingInputAsFailure(t *testing.T) {
	tmp := t.TempDir()
	items, _, fails, err := Discover([]string{"missing.md"}, tmp)
	require.NoError(t, err)
	require.Len(t, items, 0)
	require.Len(t, fails, 1)
	require.Equal(t, filepath.Join(tmp, "missing.md"), fails[0].Input)
}
