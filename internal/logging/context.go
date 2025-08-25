package logging

import (
	"context"
	"log/slog"
)

// keyer is a custom type for context keys to prevent key collisions
type keyer string

const (
	// key is the context key used to store and retrieve the logger
	key keyer = "logging"
)

// With returns a new context with the provided logger stored in it.
//
// The logger can be later retrieved using the Logger function.
func With(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

// Get retrieves the slog.Logger instance stored in the context.
//
// It panics if no logger was stored in the context.
func Get(ctx context.Context) *slog.Logger {
	return ctx.Value(key).(*slog.Logger)
}
