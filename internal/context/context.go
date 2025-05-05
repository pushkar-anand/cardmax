package context

import (
	"context"
	"errors"
	"fmt"
)

type Key string

const (
	KeyUser   Key = "user"
	KeyUserID Key = "user_id"
)

var (
	//ErrValueNotPresent is returned when the context is missing the requested value
	ErrValueNotPresent = errors.New("context doesn't have the requested value")
)

// Add a new key-value pair to the context and return the new context
func Add(ctx context.Context, key Key, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

// Get returns the value for the given key.
//
// ctx: context to read from
// key: key to read
func Get[T any](ctx context.Context, key Key) (*T, error) {
	// Get the value from the context
	value := ctx.Value(key)
	if value == nil {
		return nil, ErrValueNotPresent
	}

	// Try to typecast the value to the given type
	v, ok := value.(T)
	if !ok {
		return nil, fmt.Errorf("failed typecasting %T to %T", value, v)
	}

	// Return the value
	return &v, nil
}

// MustGet returns the value for the given key.
//
// it panics if the key is not present in the context or if the typecast fails
func MustGet[T any](ctx context.Context, key Key) *T {
	data, err := Get[T](ctx, key)
	if err != nil {
		panic(fmt.Errorf("failed to get value for key %s: %w", key, err))
	}

	return data
}
