package errorx

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder("prefix")
	if b.Prefix != "prefix" {
		t.Errorf("expected prefix 'prefix', got %q", b.Prefix)
	}
}

func TestErrorf(t *testing.T) {
	b := Errorf("formatted %s", "prefix")
	if b.Prefix != "formatted prefix" {
		t.Errorf("expected prefix 'formatted prefix', got %q", b.Prefix)
	}
}

func TestWithArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected []any
	}{
		{
			name:     "even args",
			args:     []any{"key1", "val1", "key2", 42},
			expected: []any{"key1", "val1", "key2", 42},
		},
		{
			name:     "odd args",
			args:     []any{"key1", "val1", "key2"},
			expected: []any{"key1", "val1", "key2", "<missing>"},
		},
		{
			name:     "non-string keys",
			args:     []any{1, "val1", nil, 42},
			expected: []any{"1", "val1", "<nil>", 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := WithArgs(tt.args...)
			if !reflect.DeepEqual(b.Args, tt.expected) {
				t.Errorf("expected args %+v, got %+v", tt.expected, b.Args)
			}
		})
	}
}

func TestErrorBuilder_New(t *testing.T) {
	err := NewBuilder("ctx").New("simple error")
	expected := "ctx: simple error"

	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorBuilder_Errorf(t *testing.T) {
	err := NewBuilder("ctx").Errorf("formatted %s", "error")
	expected := "ctx: formatted error"

	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorBuilder_WithPrefix(t *testing.T) {
	b := NewBuilder("old").WithPrefix("new")
	if b.Prefix != "new: old" {
		t.Errorf("expected prefix 'new: old', got %q", b.Prefix)
	}
}

func TestErrorBuilder_WithArgs(t *testing.T) {
	b := NewBuilder("").WithArgs("k1", "v1", "k2", 2)
	expected := []any{"k1", "v1", "k2", 2}
	if !reflect.DeepEqual(b.Args, expected) {
		t.Errorf("expected args %+v, got %+v", expected, b.Args)
	}
}

func TestWrap_NoPrefixNoArgs(t *testing.T) {
	err := errors.New("original")
	wrapped := ErrorBuilder{}.Wrap(err)
	if wrapped != err {
		t.Errorf("expected original error to be returned as-is")
	}
}

func TestWrap_OnlyPrefix(t *testing.T) {
	b := NewBuilder("context")
	wrapped := b.Wrap(errors.New("original"))
	expected := "context: original"

	if wrapped.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, wrapped.Error())
	}
}

func TestWrap_OnlyArgs(t *testing.T) {
	b := WithArgs("key", "value")
	wrapped := b.Wrap(errors.New("original"))
	expected := "{key=value}: original"

	if wrapped.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, wrapped.Error())
	}
}

func TestWrap_PrefixAndArgs(t *testing.T) {
	b := NewBuilder("context").WithArgs("user", 42)
	wrapped := b.Wrap(errors.New("original"))
	expected := "context: {user=42}: original"

	if wrapped.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, wrapped.Error())
	}
}

func TestWrap_NilError(t *testing.T) {
	var b ErrorBuilder
	if b.Wrap(nil) != nil {
		t.Errorf("expected nil when wrapping nil error")
	}
}

func TestKeyValueFormat_EmptyString(t *testing.T) {
	res := keyValueFormat("key", "")
	expected := "key=''"
	if res != expected {
		t.Errorf("expected %q, got %q", expected, res)
	}
}

func TestKeyValueFormat_NonEmpty(t *testing.T) {
	res := keyValueFormat("key", "value")
	expected := "key=value"
	if res != expected {
		t.Errorf("expected %q, got %q", expected, res)
	}
}
