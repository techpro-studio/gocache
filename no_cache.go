package gocache

import (
	"context"
	"time"
)

type NoCache[T any] struct{}

func (n NoCache[T]) GetAllKeys(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (n NoCache[T]) GetRaw(ctx context.Context, id string) (*string, error) {
	return nil, nil
}

func (n NoCache[T]) Delete(ctx context.Context, id string) error {
	return nil
}

func (n NoCache[T]) SetRaw(ctx context.Context, key string, item string, expiration time.Duration) error {
	return nil
}

func (n NoCache[T]) Get(ctx context.Context, id string) (*T, error) {
	return nil, nil
}

func (n NoCache[T]) Set(ctx context.Context, key string, item T) error {
	return nil
}

func (n NoCache[T]) GetMany(ctx context.Context, ids []string) (map[string]*T, error) {
	return nil, nil
}
