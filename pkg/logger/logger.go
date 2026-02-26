package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type LoggerKey struct{}
type Logger interface {
	Debug(context context.Context, msg string, fields ...zap.Field)
	Info(context context.Context, msg string, fields ...zap.Field)
	Error(context context.Context, msg string, fields ...zap.Field)
}

type L struct {
	z *zap.Logger
}

func New(env string) Logger {
	var cfg zap.Config
	switch env {
	case "dev":
		cfg = zap.NewDevelopmentConfig()
	case "prod":
		cfg = zap.NewProductionConfig()
	default:
		fmt.Println("No logger env sent; default dev used")
		cfg = zap.NewDevelopmentConfig()
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	l := L{
		z: logger,
	}
	return l
}

func (l *L) Debug(context context.Context, msg string, fields ...zap.Field) {
	l.z.Debug(msg, fields...)
}

func (l *L) Info(context context.Context, msg string, fields ...zap.Field) {
	l.z.Info(msg, fields...)
}

func (l *L) Error(context context.Context, msg string, fields ...zap.Field) {
	l.z.Error(msg, fields...)
}

func WithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, LoggerKey{}, l)
}

func FromContext(ctx context.Context) Logger {
	return ctx.Value(LoggerKey{}).(Logger)
}
