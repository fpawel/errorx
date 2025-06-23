package errorx

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasis(t *testing.T) {
	err := NewBuilder("prepend").WithArgs("key", "value", "key2", "value2").New("error message")
	require.NotNil(t, err)
	require.Equal(t, "prepend: {key=value key2=value2}: error message", err.Error())
	err2 := NewBuilder("prepend2").Wrap(err)
	require.ErrorIs(t, err2, err)
}
