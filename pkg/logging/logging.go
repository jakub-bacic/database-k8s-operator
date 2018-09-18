package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerKeyType int

const loggerKey loggerKeyType = iota

// Logger type, alias to underlying Logger type
type Logger = logrus.Entry

// Fields type, used to define extra fields to logger
type Fields = logrus.Fields

var logger *Logger

func init() {
	logger = logrus.NewEntry(logrus.New())
}

// GetLogger returns new logger from the given context
func GetLogger(ctx context.Context) *Logger {
	if ctx == nil {
		return logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return ctxLogger
	}

	return logger
}

// NewContext returns context containing logger with the extra fields
func NewContext(ctx context.Context, fields Fields) context.Context {
	return context.WithValue(ctx, loggerKey, GetLogger(ctx).WithFields(fields))
}
