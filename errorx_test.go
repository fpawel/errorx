package errorx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

var (
	reSrc   = regexp.MustCompile(`errorx_test\.go:(\d+)\.`)
	errTest = errors.New("test error")
)

func jsonify(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func jsonifyErr(err error) string {
	return replaceSrc(jsonify(Get(err).Details()))
}

func replaceSrc(s string) string {
	return reSrc.ReplaceAllString(s, "src.go:0.")
}

func printErr(err error) {
	b, _ := json.MarshalIndent(Get(err).Details(), "", "\t")
	s := string(b)
	s = replaceSrc(s)
	fmt.Printf("%s\n", s)
}

func TestWrap1(t *testing.T) {
	err := func() error {
		return Args("key3", "val3").
			Prepend("prepend3").
			Append("append3").
			Wrap(func() error {
				return Args("key2", "val2").
					Prepend("prepend2").
					Append("append2").
					Wrap(func() error {
						return Args("key1", "val1").
							Prepend("prepend1").
							Append("append1").
							Wrap(errTest)
					}())
			}())
	}()
	assert.Equal(t, "prepend3: prepend2: prepend1: test error: append1: append2: append3", err.Error())
	printErr(err)
	assertErr(t, `[
	{
		"src.go:0.TestWrap1.func1.1.1": {
			"key1": "val1"
		}
	},
	{
		"src.go:0.TestWrap1.func1.1": {
			"key2": "val2"
		}
	},
	{
		"src.go:0.TestWrap1.func1": {
			"key3": "val3"
		}
	}
]`, err)
	assert.True(t, errors.Is(err, errTest))
	assert.True(t, errors.Is(Get(err), errTest))
}

func TestWrap2(t *testing.T) {
	assertErr(t, `["src.go:0.TestWrap2"]`, New("some shit"))
}

func TestWrap3(t *testing.T) {
	err := Wrap(func3test())
	assert.Equal(t, "prepend3: prepend2: prepend1: test error: append1: append2: append3", err.Error())
	assertErr(t, `[
	{
		"src.go:0.func1test": {
			"key1": "val1"
		}
	},
	{
		"src.go:0.func2test": {
			"key2": "val2"
		}
	},
	{
		"src.go:0.func3test": {
			"key3": "val3"
		}
	},
	"src.go:0.TestWrap3"
]`, err)
	assert.True(t, errors.Is(err, errTest))
}

func TestWrap4(t *testing.T) {
	s := errTest.Error()
	e1, e2, e3, e4, e5, e6 := New(s), Errorf(s), Wrap(errTest), Skip(0).New(s), Skip(0).Errorf(s), Skip(0).Wrap(errTest)
	ex1, ex2, ex3, ex4, ex5, ex6 := Get(e1), Get(e2), Get(e3), Get(e4), Get(e5), Get(e6)
	assert.EqualValues(t, ex1.Details(), ex2.Details())
	assert.EqualValues(t, ex1.Details(), ex3.Details())
	assert.EqualValues(t, ex1.Details(), ex4.Details())
	assert.EqualValues(t, ex1.Details(), ex5.Details())
	assert.EqualValues(t, ex1.Details(), ex6.Details())
}

func TestExternal(t *testing.T) {
	err := Get(Args("external", "some external text").Wrap(errTest))

	assertErr(t, `[
	{
		"src.go:0.TestExternal":{
				"external": "some external text"
		}
	}
]`, err)
	assert.True(t, errors.Is(err, errTest))
	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, "some external text", err.Value("external"))

	assert.Equal(t, "some external text2", Get(fmt.Errorf("%w", errors.Join(
		errors.New(""),
		Args("external", "some external text2").New("err2"),
		Args("external", "some external text1").Wrap(err),
		Args("external", "some external text2").New("err3"),
	))).Value("external"))

}

func TestCode(t *testing.T) {
	err := Get(Args("code", "test code").Wrap(errTest))
	assert.Equal(t, "test code", err.Value("code"))
}

func TestPrepend(t *testing.T) {
	err := func() error {
		return Args("key3", "val3").
			Prepend("prepend3").
			Append("append3").
			Wrap(func() error {
				return Args("key2", "val2").
					Prepend("prepend2").
					Append("append2").
					Wrap(func() error {
						return Args("key1", "val1").
							Prepend("prepend1").
							Append("append1").
							Wrap(errTest)
					}())
			}())
	}()

	assertErr(t, `[
	{
		"src.go:0.TestPrepend.func1.1.1": {
			"key1": "val1"
		}
	},
	{
		"src.go:0.TestPrepend.func1.1": {
			"key2": "val2"
		}
	},
	{
		"src.go:0.TestPrepend.func1": {
			"key3": "val3"
		}
	}
]`, err)
}

func assertErr(t *testing.T, expected string, actual error) {
	assert.JSONEq(t, expected, jsonifyErr(actual))
}

func func1test() error {
	return Args("key1", "val1").
		Prepend("prepend1").
		Append("append1").
		Wrap(errTest)
}

func func2test() error {
	return Args("key2", "val2").
		Prepend("prepend2").
		Append("append2").
		Wrap(func1test())
}

func func3test() error {
	return Args("key3", "val3").
		Prepend("prepend3").
		Append("append3").
		Wrap(func2test())
}
