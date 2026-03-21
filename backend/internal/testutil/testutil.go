// Package testutil provides shared helpers for tests.
package testutil

import (
	"context"

	"go.uber.org/zap"

	"meetings-editor/pkg/logger"
)

// Ctx returns a context with a no-op zap logger injected, satisfying
// logger.FromContext() calls in services and repositories.
func Ctx() context.Context {
	nop := &nopLogger{zap.NewNop()}
	return logger.WithLogger(context.Background(), nop)
}

type nopLogger struct{ z *zap.Logger }

func (n *nopLogger) Debug(_ context.Context, _ string, _ ...zap.Field) {}
func (n *nopLogger) Info(_ context.Context, _ string, _ ...zap.Field)  {}
func (n *nopLogger) Error(_ context.Context, _ string, _ ...zap.Field) {}
