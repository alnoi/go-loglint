package constmsg

import (
	"log/slog"
)

func ConcatMessages() {
	slog.Info("Hel" + "lo") // want "loglint\\(lowercase\\)"

	slog.Info("ok" + "...") // want "loglint\\(no-emoji\\)"

	slog.Info("ok" + "…") // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"

	slog.Info("пр" + "ивет") // want "loglint\\(english-only\\)"

	slog.Info("he" + "llo")
}
