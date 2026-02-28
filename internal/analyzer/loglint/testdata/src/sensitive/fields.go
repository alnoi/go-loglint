package sensitive

import (
	"log/slog"

	"go.uber.org/zap"
)

type User struct {
	Password string
	Token    string
	APIKey   string
	Email    string
}

func StructField(u User) {
	logger := zap.NewNop()

	slog.Info("msg", u.Password) // want "loglint\\(no-sensitive\\)"
	slog.Info("msg", u.Token)    // want "loglint\\(no-sensitive\\)"
	slog.Info("msg", u.APIKey)   // want "loglint\\(no-sensitive\\)"
	slog.Info("msg", u.Email)

	slog.Info("msg", slog.String("password", u.Password)) // want "loglint\\(no-sensitive\\)"
	slog.Info("msg", slog.String("token", u.Token))       // want "loglint\\(no-sensitive\\)"
	slog.Info("msg", slog.String("api_key", u.APIKey))    // want "loglint\\(no-sensitive\\)"
	slog.Info("msg", slog.String("email", u.Email))

	logger.Info("msg", zap.String("password", u.Password)) // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("token", u.Token))       // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("api_key", u.APIKey))    // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("email", u.Email))
}
