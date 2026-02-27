package constmsg

import (
	"fmt"
	"log"
	"log/slog"
)

func SprintfMessages() {
	slog.Info(fmt.Sprintf("Hello")) // want "loglint\\(lowercase\\)"
	log.Print(fmt.Sprintf("Hello")) // want "loglint\\(lowercase\\)"

	slog.Info(fmt.Sprintf("%s", "Hello")) // want "loglint\\(lowercase\\)"
	log.Print(fmt.Sprintf("%s", "Hello")) // want "loglint\\(lowercase\\)"

	slog.Info(fmt.Sprintf("ok...")) // want "loglint\\(no-emoji\\)"
	log.Print(fmt.Sprintf("ok...")) // want "loglint\\(no-emoji\\)"

	slog.Info(fmt.Sprintf("ok…")) // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	log.Print(fmt.Sprintf("ok…")) // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"

	slog.Info(fmt.Sprintf("привет")) // want "loglint\\(english-only\\)"
	log.Print(fmt.Sprintf("привет")) // want "loglint\\(english-only\\)"

	name := "Hello"
	slog.Info(fmt.Sprintf("%s", name))
	log.Print(fmt.Sprintf("%s", name))
}
