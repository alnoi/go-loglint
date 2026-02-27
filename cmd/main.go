package main

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
)

func main() {
	logger := zap.NewNop()

	logger.Info("Hello World")
	logger.Info("hello World")
	slog.Info(fmt.Sprintf("%s", "Hello"))
	slog.Info("A" + "b")
	const f = "Fg"
	slog.Info(f)
	slog.Info("Hello")                             // uppercase -> lowercase violation
	slog.Info("ok...")                             // "..." violation
	slog.Info("ok…")                               // unicode ellipsis violation
	slog.Info("all good")                          // ok
	slog.Info("why?!")                             // '!' or '?' violation
	slog.Info("started 🚀")                         // emoji violation (если englishOnly выключишь)
	slog.Info("user logged in", "password", "123") // sensitive (через поля)
}
