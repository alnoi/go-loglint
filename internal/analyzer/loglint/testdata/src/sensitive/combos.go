package sensitive

import (
	"log/slog"

	"go.uber.org/zap"
)

func ComboCases() {
	logger := zap.NewNop()

	password := "123"
	token := "abc"

	slog.Info("msg", password, token)                                                // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("password", password), zap.String("token", token)) // want "loglint\\(no-sensitive\\)"

	slog.Info("msg", password+"_suffix")                           // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("password", password+"_suffix")) // want "loglint\\(no-sensitive\\)"

	username := "john"
	slog.Info("msg", username, password)                                                   // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("username", username), zap.String("password", password)) // want "loglint\\(no-sensitive\\)"

	slog.Info("msg", username)
	logger.Info("msg", zap.String("username", username))
}
