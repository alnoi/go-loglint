package apis

import (
	"log/slog"
)

func PkgCalls() {
	slog.Info("Hello") // want "loglint\\(lowercase\\)"

	slog.Info("ok...") // want "loglint\\(no-emoji\\)"

	slog.Info("ok…") // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	slog.Info("пр")  // want "loglint\\(english-only\\)"

	slog.Info("hello")
}
