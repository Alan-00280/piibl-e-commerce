package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func parseLevel(val string) slog.Level {
	switch strings.ToLower(val) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewLogger() *slog.Logger {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		panic("gagal")
	}

	rotator := &lumberjack.Logger{
		Filename: filepath.Join("logs", "app.log"),
		MaxSize: 10,
		MaxBackups: 5,
		MaxAge: 14,
		Compress: true,
	}

	writer := io.MultiWriter(os.Stdout, rotator)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(GetEnv("LOG_LEVEL", "info")),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.String("time", time.Now().Format(time.UnixDate))
			}
			return a
		},
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
