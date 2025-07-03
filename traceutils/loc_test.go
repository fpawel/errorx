package traceutils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLineOfCode_CurrentFunction(t *testing.T) {
	require.Equal(t,
		"loc_test.go:11",
		FileLine(0))
	require.Equal(t,
		"github.com/fpawel/errorx/traceutils.TestLineOfCode_CurrentFunction",
		Function(0))
}

func TestLineOfCode_WrapperFunction(t *testing.T) {
	wrapper := func(skip int) (string, string) {
		return Function(skip), FileLine(skip)
	}
	fun, fileLine := wrapper(0)
	require.Equal(t, "github.com/fpawel/errorx/traceutils.TestLineOfCode_WrapperFunction.func1", fun)
	require.Equal(t, "loc_test.go:19", fileLine)
	fun, fileLine = wrapper(1)
	require.Equal(t, "github.com/fpawel/errorx/traceutils.TestLineOfCode_WrapperFunction", fun)
	require.Equal(t, "loc_test.go:24", fileLine)
}
