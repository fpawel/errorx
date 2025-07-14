package traceutils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// FileLine возвращает строку с именем файла и номером строки,
// откуда была вызвана функция, с учетом указанного смещения skip.
//
// Формат возвращаемой строки: "<имя_файла>:<номер_строки>".
//
// Аргументы:
//   - skip: количество фреймов вызова, которые нужно пропустить.
//     Например, 0 — получить информацию о непосредственном вызывающем.
//
// Возвращает:
//   - строку вида "main.go:42", либо "unknown", если информацию получить не удалось.
func FileLine(skip int) string {
	frame := Frame(skip + 1)
	if frame.File == "" {
		return "unknown"
	}

	// Используем только имя файла (без пути) и номер строки
	return fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
}

// Function возвращает полное имя функции (включая путь пакета),
// откуда была вызвана функция, с учетом указанного смещения skip.
//
// Аргументы:
//   - skip: количество фреймов вызова, которые нужно пропустить.
//     Например, 0 — получить информацию о непосредственном вызывающем.
//
// Возвращает:
//   - имя функции (например, "main.myFunc"), либо "unknown", если информацию получить не удалось.
func Function(skip int) string {
	frame := Frame(skip + 1)
	if frame.File == "" {
		return "unknown"
	}

	return frame.Function
}

func ShortFunction(skip int) string {
	frame := Frame(skip + 1)
	if frame.File == "" {
		return "unknown"
	}

	return filepath.Base(frame.Function)
}

// Frame возвращает структуру runtime.Frame с информацией о месте вызова,
// с учетом указанного смещения skip.
//
// Аргументы:
//   - skip: количество фреймов вызова, которые нужно пропустить
//     относительно кода, вызывающего эту функцию.
//
// Возвращает:
//   - runtime.Frame с данными о файле, номере строки и функции.
//     Если стек вызовов пуст, возвращается пустая структура.
func Frame(skip int) runtime.Frame {
	// callersToSkip — количество внутренних уровней, которые нужно пропустить,
	// чтобы добраться до вызывающего кода (саму функцию Frame и runtime.Callers)
	const callersToSkip = 2

	// pc — массив для хранения адресов вызовов
	pc := make([]uintptr, 1)

	// Получаем адреса вызовов начиная с указанного смещения
	n := runtime.Callers(skip+callersToSkip, pc)
	if n == 0 {
		return runtime.Frame{} // не удалось получить стек вызовов
	}

	// Преобразуем адреса в фреймы
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next() // Получаем первый (самый внешний) фрейм

	return frame
}
