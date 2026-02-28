package apis

import (
	"log/slog"

	"go.uber.org/zap"
)

func ReceiverCalls() {
	l := slog.Default()

	l.Info("Hello")  // want "loglint\\(lowercase\\)"
	l.Info("ok...")  // want "loglint\\(no-emoji\\)"
	l.Info("ok…")    // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	l.Info("привет") // want "loglint\\(english-only\\)"
	l.Info("why?!")  // want "loglint\\(no-emoji\\)"

	l.Info("hello")

	logger := zap.NewNop()

	logger.Info("Hello") // want "loglint\\(lowercase\\)"
	logger.Info("ok...") // want "loglint\\(no-emoji\\)"
	logger.Info("ok…")   // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	logger.Info("пр")    // want "loglint\\(english-only\\)"
	logger.Info("why?!") // want "loglint\\(no-emoji\\)"

	logger.Info("hello")
}
