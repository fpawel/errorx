package errorx

import (
	"errors"
	"fmt"
	"github.com/fpawel/errorx/traceutils"
	"strings"
)

// ErrorBuilder накапливает контекст (префикс и аргументы),
// который может быть добавлен к ошибке при вызове Wrap.
type ErrorBuilder struct {
	// Prefix добавляется перед ошибкой в итоговом сообщении.
	Prefix string
	// Args — пары ключ-значение, которые будут добавлены после префикса.
	// Каждый ключ должен иметь соответствующее значение.
	Args []any
}

// NewBuilder создает новый ErrorBuilder с указанным сообщением-префиксом.
func NewBuilder(msg string) ErrorBuilder {
	return ErrorBuilder{}.WithPrefix(msg)
}

// Errorf создает ErrorBuilder с отформатированным сообщением-префиксом.
func Errorf(format string, args ...any) ErrorBuilder {
	return ErrorBuilder{}.WithPrefix(fmt.Sprintf(format, args...))
}

// WithArgs создает ErrorBuilder с указанными парами "ключ-значение".
// При вызове с нечётным числом аргументов пропущенные значения заменяются на "<missing>".
// Ключи приводятся к строке с помощью fmt.Sprintf.
func WithArgs(args ...any) ErrorBuilder {
	return ErrorBuilder{}.WithArgs(args...)
}

// New создает ошибку с указанным сообщением и оборачивает её с добавленным контекстом.
//
// Аналогично Wrap(errors.New(s)).
func (b ErrorBuilder) New(s string) error {
	return b.Wrap(errors.New(s))
}

// Errorf создает ошибку с отформатированным сообщением и оборачивает её с добавленным контекстом.
//
// Аналогично Wrap(fmt.Errorf(...)).
func (b ErrorBuilder) Errorf(format string, a ...any) error {
	return b.Wrap(fmt.Errorf(format, a...))
}

// WithPrefix добавляет префикс к ErrorBuilder.
// Повторные вызовы добавляют новый текст перед уже существующим через ": ".
func (b ErrorBuilder) WithPrefix(s string) ErrorBuilder {
	if b.Prefix == "" {
		b.Prefix = s
	} else {
		b.Prefix = s + ": " + b.Prefix
	}
	return b
}

// WithArgs добавляет пары "ключ-значение" в ErrorBuilder.
// Ключи приводятся к строке через fmt.Sprintf.
// При нечётном количестве аргументов — последнее значение будет заменено на "<missing>".
func (b ErrorBuilder) WithArgs(args ...any) ErrorBuilder {
	for i := 0; i < len(args); i += 2 {
		k := fmt.Sprintf("%v", args[i])
		var v any = "<missing>"
		if i+1 < len(args) {
			v = args[i+1]
		}
		b.Args = append(b.Args, k, v)
	}
	return b
}

// ExtendPrefix adds a suffix to the current prefix, forming a longer context chain.
func (b ErrorBuilder) ExtendPrefix(suffix string) ErrorBuilder {
	if b.Prefix == "" {
		b.Prefix = suffix
	} else {
		b.Prefix += ": " + suffix
	}
	return b
}

// Wrap оборачивает переданную ошибку с добавленным префиксом и аргументами.
//
// Поведение:
//   - Если err == nil, метод немедленно возвращает nil и ничего не делает;
//   - Если задан Prefix, он добавляется перед ошибкой через ": ";
//   - Если заданы аргументы, они добавляются в фигурных скобках после префикса,
//     ключ=значение разделены запятыми;
//   - Финальный формат: "<Prefix>: {key1=val1, key2=val2}: оригинальная ошибка".
//
// Пример:
//
//	err := NewBuilder("context").WithArgs("user", 42).Wrap(someErr)
//	// context: {user=42}: someErr
func (b ErrorBuilder) Wrap(err error) error {
	if err == nil {
		return nil
	}

	var sb strings.Builder

	if b.Prefix != "" {
		sb.WriteString(b.Prefix)
	}

	if len(b.Args) > 0 {
		if sb.Len() > 0 {
			sb.WriteString(": ")
		}
		sb.WriteString("{")
		for i := 0; i < len(b.Args); i += 2 {
			if i > 0 {
				sb.WriteString(", ")
			}
			k := fmt.Sprintf("%v", b.Args[i])
			sb.WriteString(keyValueFormat(k, b.Args[i+1]))
		}
		sb.WriteString("}")
	}

	if sb.Len() > 0 {
		return fmt.Errorf("%s: %w", sb.String(), err)
	}
	return err
}

func (b ErrorBuilder) WithFileLine() ErrorBuilder {
	b.Prefix += ": " + traceutils.FileLine(1)
	return b
}

func (b ErrorBuilder) WithFunction() ErrorBuilder {
	b.Prefix += ": " + traceutils.Function(1)
	return b
}

// keyValueFormat форматирует пару ключ=значение как строку.
// Если значение — пустая строка, оно заменяется на ”.
func keyValueFormat(k string, v any) string {
	if s, ok := v.(string); ok && s == "" {
		v = "''"
	}
	return fmt.Sprintf("%s=%v", k, v)
}
