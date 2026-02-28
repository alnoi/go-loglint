package sensitive

import (
	"log/slog"

	"go.uber.org/zap"
)

func AllowlistCases() {
	logger := zap.NewNop()

	tokenbucket := "123"
	slog.Info("msg", tokenbucket)
	logger.Info("msg", zap.String("tokenbucket", tokenbucket))

	tokenized := "abc"
	slog.Info("msg", tokenized)
	logger.Info("msg", zap.String("tokenized", tokenized))

	secretariat := "data"
	slog.Info("msg", secretariat)
	logger.Info("msg", zap.String("secretariat", secretariat))

	password := "123"
	slog.Info("msg", password)                           // want "loglint\\(no-sensitive\\)"
	logger.Info("msg", zap.String("password", password)) // want "loglint\\(no-sensitive\\)"
}
