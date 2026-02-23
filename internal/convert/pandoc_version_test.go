package convert

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractVersionToken(t *testing.T) {
	v, ok := extractVersionToken("pandoc 3.1.11")
	require.True(t, ok)
	require.Equal(t, "3.1.11", v)

	v, ok = extractVersionToken("pandoc 2.19")
	require.True(t, ok)
	require.Equal(t, "2.19.0", v)

	_, ok = extractVersionToken("pandoc unknown")
	require.False(t, ok)
}

func TestInstallHint(t *testing.T) {
	require.Contains(t, installHint("darwin"), "brew install pandoc")
	require.Contains(t, installHint("windows"), "scoop install pandoc")
	require.Contains(t, installHint("linux"), "apt-get install pandoc")
}
