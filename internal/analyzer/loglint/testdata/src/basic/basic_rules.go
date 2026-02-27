package basic

import "log/slog"

func BasicRules() {
	slog.Info("Hello") // want "loglint\\(lowercase\\)"
	slog.Info("hello")

	slog.Info("привет") // want "loglint\\(english-only\\)"

	slog.Info("started service")

	slog.Info("ok...")     // want "loglint\\(no-emoji\\)"
	slog.Info("ok…")       // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	slog.Info("why?!")     // want "loglint\\(no-emoji\\)"
	slog.Info("started 🚀") // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
}
