package format

import (
	"context"
	"template-go-cli/internal/types/output"
)

// keyer is a custom type for context keys to prevent key collisions.
type keyer string

const (
	// key is the context key used to store and retrieve the package's context value. See [With] and [Get] for additional details.
	key keyer = "format"
)

func With(ctx context.Context, v output.Type) context.Context {
	return context.WithValue(ctx, key, v)
}

func Get(ctx context.Context) output.Type {
	return ctx.Value(key).(output.Type)
}
