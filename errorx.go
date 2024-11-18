package errorx

import (
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"log/slog"
	"path/filepath"
	"runtime"
)

type (
	// Error standard error wrapper with additional information
	Error struct {
		Frames []Frame
		Err    error
	}

	Const string

	// Frame additional information related to error
	Frame struct {
		Loc  string `json:"loc"`                          // Where the error was created in the code
		Args M      `json:",omitempty" yaml:",omitempty"` // Structural arguments
	}

	// M alias for abbreviation
	M = map[string]any

	// Builder to simplify error constructors
	Builder struct {
		skip    int    // How many stack frames to skip when calculating where the error was created
		prepend string // Text appended to the beginning of the error message, separated by ":"
		append  string // Text appended to the end of the error message, separated by ":"
		args    M      // Structural error arguments
	}
)

func (s Const) Error() string {
	return string(s)
}

// Attr attribute for error logging
func Attr(err error) slog.Attr {
	if e := Get(err); len(e.Frames) != 0 {
		return slog.Group(err.Error(), e.AttrsAny()...)
	}
	return slog.String("error", err.Error())
}

func (e Error) Value(arg string) any {
	for _, f := range e.Frames {
		v, ok := f.Args[arg]
		if ok {
			return v
		}
	}

	var wrapped Error
	if errors.As(e.Err, &wrapped) {
		return wrapped.Value(arg)
	}
	return nil
}

func (e Error) Error() string {
	return e.Err.Error()
}

// Details list of parent errors with arguments to log
func (e Error) Details() []any {
	xs := make([]any, 0, len(e.Frames))
	for _, x := range e.Frames {
		if len(x.Args) == 0 {
			xs = append(xs, x.Loc)
		} else {
			xs = append(xs, M{x.Loc: x.Args})
		}
	}
	return xs
}

func (e Error) AttrsAny() []any {
	return pie.Map(e.Attrs(), func(a slog.Attr) any {
		return a
	})
}

func (e Error) Attrs() []slog.Attr {
	xs := make([]slog.Attr, 0, len(e.Frames))
	for _, x := range e.Frames {
		if len(x.Args) == 0 {
			xs = append(xs, slog.String(x.Loc, ""))
		} else {
			xs = append(xs, slog.Any(x.Loc, x.Args))
		}
	}
	return xs
}

// Unwrap original error. Necessary for errors.Is and errors.As to work correctly
func (e Error) Unwrap() error {
	if v, ok := e.Err.(interface{ Unwrap() error }); ok {
		return v.Unwrap()
	}
	return e.Err
}

// Get finds an object of type Error in the error tree of the original Err object and returns it (see errors.As).
// Otherwise, Error is returned with default fields except Error.Err = Err,
// that is, Error with the original error wrapped without any additional information.
func Get(err error) Error {
	var wrapped Error
	if !errors.As(err, &wrapped) {
		wrapped.Err = err
	}
	return wrapped
}

// Wrap wraps the original error in Error according to the additional error context passed in the opts constructor
// The loc of the Wrap call is added to the additional error context
func Wrap(err error) error {
	return Skip(1).Wrap(err)
}

// New create an Error object with the text message
// The loc of the New call is added to the additional error context
func New(message string) error {
	return Skip(1).New(message)
}

// Errorf create an Error object with printf-like formatted text
// The loc of the Errorf call is added to the additional error context
func Errorf(format string, args ...any) error {
	return Skip(1).Errorf(format, args...)
}

// Skip the Error constructor, skipping stack frames loc
func Skip(skip int) Builder {
	return Builder{}.Skip(skip)
}

// Args Error constructor with structure arguments
func Args(args ...any) Builder {
	return Builder{}.Args(args...)
}

// Prepend the Error constructor with a prefix in the text
func Prepend(s string) Builder {
	return Builder{}.Prepend(s)
}

// Prependf the Error constructor with a prefix in printf-like formatted text
func Prependf(format string, args ...any) Builder {
	return Builder{}.Prependf(format, args...)
}

// Append the Error constructor with a suffix in the text
func Append(s string) Builder {
	return Builder{}.Append(s)
}

// Appendf the Error constructor with a prefix in printf-like formatted text
func Appendf(format string, args ...any) Builder {
	return Builder{}.Appendf(format, args...)
}

// New create an Error with the specified message
func (o Builder) New(msg string) error {
	return o.wrap(errors.New(msg))
}

// Errorf version of New with printf formatting
func (o Builder) Errorf(format string, args ...any) error {
	return o.wrap(fmt.Errorf(format, args...))
}

// Wrap constructor for wrapping the original error in Error.
func (o Builder) Wrap(err error) error {
	if err == nil {
		return nil
	}
	return o.wrap(err)
}

// Skip adds skipping stack frames loc to the original error constructor
func (o Builder) Skip(skip int) Builder {
	o.skip = skip
	return o
}

// Prepend adds a prefix to the error text
func (o Builder) Prepend(prepend string) Builder {
	if o.prepend == "" {
		o.prepend = prepend
	} else {
		o.prepend += ": " + prepend
	}
	return o
}

// Prependf version of Prepend with printf formatting
func (o Builder) Prependf(format string, args ...any) Builder {
	return o.Prepend(fmt.Sprintf(format, args...))
}

// Append adds a suffix to the error text
func (o Builder) Append(append string) Builder {
	if o.append == "" {
		o.append = append
	} else {
		o.append += ": " + append
	}
	return o
}

// Appendf version of Append with printf formatting
func (o Builder) Appendf(format string, args ...any) Builder {
	return o.Append(fmt.Sprintf(format, args...))
}

// Args structure arguments
func (o Builder) Args(args ...any) Builder {
	if o.args == nil {
		o.args = M{}
	}
	for i := 0; i < len(args); i += 2 {
		k := args[i]
		var v any = "?"
		if i+1 < len(args) {
			v = args[i+1]
		}
		o.args[k.(string)] = v
	}
	return o
}

func (o Builder) wrap(err error) Error {
	d := Frame{
		Loc:  loc(2 + o.skip),
		Args: o.args,
	}

	var wrapped Error
	errors.As(err, &wrapped)

	if o.prepend != "" {
		if err.Error() == "" {
			err = fmt.Errorf("%s%w", o.prepend, err)
		} else {
			err = fmt.Errorf("%s: %w", o.prepend, err)
		}
	}

	if o.append != "" {
		if err.Error() == "" {
			err = fmt.Errorf("%w%s", err, o.append)
		} else {
			err = fmt.Errorf("%w: %s", err, o.append)
		}
	}

	wrapped.Frames = append(wrapped.Frames, d)
	wrapped.Err = err
	return wrapped
}

func loc(skip int) string {
	return formatFrame(skip+3, func(frame runtime.Frame) string {
		function := filepath.Base(frame.Function)
		for i, ch := range function {
			if string(ch) == "." {
				function = function[i:]
				break
			}
		}
		return fmt.Sprintf("%s:%d%s", filepath.Base(frame.File), frame.Line, function)
	})
}

func formatFrame(skip int, f func(frame runtime.Frame) string) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return f(frame)
}
