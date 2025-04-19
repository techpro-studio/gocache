package gocache

import (
	"context"
	"time"
)

type TypedCache[T any] interface {
	Cache
	Get(ctx context.Context, id string) (*T, error)
	Set(ctx context.Context, key string, item T) error
	GetMany(ctx context.Context, ids []string) (map[string]*T, error)
}

type Cache interface {
	GetAllKeys(ctx context.Context) ([]string, error)
	GetRaw(ctx context.Context, id string) (*string, error)
	Delete(ctx context.Context, id string) error
	SetRaw(ctx context.Context, key string, item string, expiration time.Duration) error
}
