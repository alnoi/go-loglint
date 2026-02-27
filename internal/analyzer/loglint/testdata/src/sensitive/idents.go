package sensitive

import (
	"log/slog"

	"go.uber.org/zap"
)

func IdentCases() {
	logger := zap.NewNop()

	password := "123"
	token := "abc"
	apiKey := "key"
	secret := "sec"

	slog.Info("msg", password)                           // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("password", password)) // want "loglint\\(no-sensitive\\)"

	slog.Info("msg", token)                        // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("token", token)) // want "loglint\\(no-sensitive\\)"

	slog.Info("msg", apiKey)                          // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("api_key", apiKey)) // want "loglint\\(no-sensitive\\)"

	slog.Info("msg", secret)                         // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("secret", secret)) // want "loglint\\(no-sensitive\\)"

	username := "john"
	slog.Info("msg", username)
	logger.Info("msg", zap.String("username", username))
}
