package sensitive

import (
	"log"
	"log/slog"
)

func AllowlistCases() {
	tokenbucket := "123"
	slog.Info("msg", tokenbucket)
	log.Print("msg", tokenbucket)

	tokenized := "abc"
	slog.Info("msg", tokenized)
	log.Print("msg", tokenized)

	secretariat := "data"
	slog.Info("msg", secretariat)
	log.Print("msg", secretariat)

	password := "123"
	slog.Info("msg", password) // want "loglint\\(no-sensitive\\)"
	log.Print("msg", password) // want "loglint\\(no-sensitive\\)"
}
