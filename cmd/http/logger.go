package main

import (
	"io"
	"log/slog"
	"time"
)

func replaceTimeFormat(group []string, attr slog.Attr) slog.Attr {
	if attr.Key == "time" {
		// format date: yyyy-mm-ddTHH:MM:SS
		value := time.Now().Format("2006-01-02T15:04:05")
		return slog.Attr{Key: attr.Key, Value: slog.StringValue(value)}
	}
	return slog.Attr{Key: attr.Key, Value: attr.Value}
}

func newLogger(out io.Writer, minLevel slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(
		out,
		&slog.HandlerOptions{
			AddSource:   true,
			Level:       minLevel,
			ReplaceAttr: replaceTimeFormat,
		}))
}
