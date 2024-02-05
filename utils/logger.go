package utils

import (
	"log/slog"
	"time"
)

func LogTime(l *slog.Logger, b time.Time, msg string, args ...any) {
	d := time.Now().Sub(b)
	args = append(args, "time", d)
	if d > time.Millisecond {
		l.Info(msg, args...)
	}
	l.Debug(msg, args...)
}
