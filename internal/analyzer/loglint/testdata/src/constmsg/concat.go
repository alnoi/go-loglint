package constmsg

import (
	"log/slog"

	"go.uber.org/zap"
)

func ConcatMessages() {
	logger := zap.NewNop()

	slog.Info("Hel" + "lo")   // want "loglint\\(lowercase\\)"
	logger.Info("Hel" + "lo") // want "loglint\\(lowercase\\)"

	slog.Info("ok" + "...")   // want "loglint\\(no-emoji\\)"
	logger.Info("ok" + "...") // want "loglint\\(no-emoji\\)"

	slog.Info("ok" + "…")   // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	logger.Info("ok" + "…") // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"

	slog.Info("пр" + "ивет")   // want "loglint\\(english-only\\)"
	logger.Info("пр" + "ивет") // want "loglint\\(english-only\\)"

	slog.Info("he" + "llo")
	logger.Info("he" + "llo")
}
