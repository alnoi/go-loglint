package constmsg

import (
	"log/slog"
)

func LocalConstMessages() {
	const hello = "Hello"
	slog.Info(hello) // want "loglint\\(lowercase\\)"

	const dots = "ok..."
	slog.Info(dots) // want "loglint\\(no-emoji\\)"

	const unicode = "ok…"
	slog.Info(unicode) // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"

	const nonASCII = "привет"
	slog.Info(nonASCII) // want "loglint\\(english-only\\)"

	const valid = "hello"
	slog.Info(valid)

	var notConst = "Hello"
	slog.Info(notConst)
}
