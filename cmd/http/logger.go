package main

import (
	"io"
	"log/slog"
	"time"
)

func replaceTimeFormat(group []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		value := time.Now().Format("2006-01-02T15:04:05")
		return slog.Attr{Key: a.Key, Value: slog.StringValue(value)}
	}
	return slog.Attr{Key: a.Key, Value: a.Value}
}

func newLogger(out io.Writer, minLevel slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level:       minLevel,
		AddSource:   true,
		ReplaceAttr: replaceTimeFormat,
	}))
}
