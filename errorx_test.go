package errorx

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorBuilder_WithFunction(t *testing.T) {
	require.Equal(t,
		"init: github.com/fpawel/errorx.TestErrorBuilder_WithFunction: error",
		NewBuilder("init").WithFunction().New("error").Error())
}

func TestErrorBuilder_WithFileLine(t *testing.T) {
	require.Equal(t,
		"init: errorx_test.go:20: error",
		NewBuilder("init").WithFileLine().New("error").Error())
}

func TestErrorBuilder_WrapVariants(t *testing.T) {
	tests := []struct {
		name     string
		builder  ErrorBuilder
		inputErr error
		expected string
	}{
		{
			name:     "With prefix only",
			builder:  NewBuilder("ctx"),
			inputErr: errors.New("original"),
			expected: "ctx: original",
		},
		{
			name:     "With args only",
			builder:  WithArgs("key", "value"),
			inputErr: errors.New("original"),
			expected: "{key=value}: original",
		},
		{
			name:     "With prefix and args",
			builder:  NewBuilder("ctx").WithArgs("user", 42),
			inputErr: errors.New("original"),
			expected: "ctx: {user=42}: original",
		},
		{
			name:     "With ExtendPrefix",
			builder:  NewBuilder("base").ExtendPrefix("extra"),
			inputErr: errors.New("fail"),
			expected: "base: extra: fail",
		},
		{
			name:     "Wrap nil error",
			builder:  NewBuilder("ctx"),
			inputErr: nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder.Wrap(tt.inputErr)
			if tt.inputErr == nil {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tt.expected)
			}
		})
	}
}

func TestErrorBuilder_WithArgsVariants(t *testing.T) {
	assert.Equal(t,
		[]any{"key1", "val1", "key2", 42},
		WithArgs("key1", "val1", "key2", 42).Args,
	)

	assert.Equal(t,
		[]any{"key1", "val1", "key2", "<missing>"},
		WithArgs("key1", "val1", "key2").Args,
	)

	assert.Equal(t,
		[]any{"1", "val1", "<nil>", 42},
		WithArgs(1, "val1", nil, 42).Args,
	)
}

func TestErrorBuilder_NewAndErrorf(t *testing.T) {
	err := NewBuilder("ctx").New("something went wrong")
	assert.EqualError(t, err, "ctx: something went wrong")

	errf := NewBuilder("ctx").Errorf("error %d", 500)
	assert.EqualError(t, errf, "ctx: error 500")
}

func TestErrorBuilder_WithPrefix(t *testing.T) {
	b := NewBuilder("initial").WithPrefix("outer")
	assert.Equal(t, "outer: initial", b.Prefix)
}

func TestKeyValueFormat(t *testing.T) {
	assert.Equal(t, "key=value", keyValueFormat("key", "value"))
	assert.Equal(t, "key=''", keyValueFormat("key", ""))
}
