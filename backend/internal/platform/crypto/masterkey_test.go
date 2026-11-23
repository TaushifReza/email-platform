package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadMasterKey_SEC04(t *testing.T) {
	out, err := LoadMasterKey([]byte("  0123456789abcdef0123456789abcdef\n"))
	require.NoError(t, err)
	require.Equal(t, []byte("0123456789abcdef0123456789abcdef"), out)
	require.NotSame(t, &out[0], &[]byte("x")[0])

	_, err = LoadMasterKey([]byte("short"))
	require.Error(t, err)

	_, err = LoadMasterKey(nil)
	require.Error(t, err)
}
