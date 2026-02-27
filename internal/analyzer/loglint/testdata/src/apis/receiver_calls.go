package apis

import (
	"log"
	"log/slog"
)

func ReceiverCalls() {
	l := slog.Default()

	l.Info("Hello")  // want "loglint\\(lowercase\\)"
	l.Info("ok...")  // want "loglint\\(no-emoji\\)"
	l.Info("ok…")    // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	l.Info("привет") // want "loglint\\(english-only\\)"
	l.Info("why?!")  // want "loglint\\(no-emoji\\)"

	l.Info("hello")

	std := log.Default()

	std.Print("Hello")  // want "loglint\\(lowercase\\)"
	std.Print("ok...")  // want "loglint\\(no-emoji\\)"
	std.Print("ok…")    // want "loglint\\(no-emoji\\)" "loglint\\(english-only\\)"
	std.Print("привет") // want "loglint\\(english-only\\)"
	std.Print("why?!")  // want "loglint\\(no-emoji\\)"

	std.Print("hello")
}
