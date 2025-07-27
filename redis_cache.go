package gocache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

// RedisCache is a generic Redis-based implementation of the Cache interface.
type RedisCache struct {
	redisClient *redis.Client
	prefix      string
}

func NewRedisCache(redisClient *redis.Client, prefix string) *RedisCache {
	return &RedisCache{redisClient: redisClient, prefix: prefix}
}

func (r *RedisCache) GetAllKeys(ctx context.Context) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		scannedKeys, nextCursor, err := r.redisClient.Scan(ctx, cursor, fmt.Sprintf("%s:*", r.prefix), 100).Result()
		if err != nil {
			return nil, err
		}

		for _, scannedKey := range scannedKeys {
			id := strings.Split(scannedKey, ":")[1]
			keys = append(keys, id)
		}

		// If cursor is 0, the iteration is complete
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return keys, nil
}

func (r *RedisCache) Delete(ctx context.Context, id string) error {
	_, err := r.redisClient.Del(ctx, r._buildKey(id)).Result()
	return err
}

func (r *RedisCache) GetRaw(ctx context.Context, id string) (*string, error) {
	result, err := r.redisClient.Get(ctx, r._buildKey(id)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// Get retrieves an item from Redis by its ID.

func (r *RedisCache) _buildKey(id string) string {
	return fmt.Sprintf("%s:%s", r.prefix, id)
}

func (r *RedisCache) SetRaw(ctx context.Context, key string, item string, expiration time.Duration) error {
	return r.redisClient.Set(ctx, r._buildKey(key), item, expiration).Err()
}

func (r *RedisCache) GetManyRaw(ctx context.Context, ids []string) (map[string]*string, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided for GetMany")
	}

	// Prefix each ID
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = r._buildKey(id)
	}

	// Use MGET to retrieve multiple keys
	results, err := r.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	// Prepare the result map
	items := make(map[string]*string, len(ids))

	for i, res := range results {
		id := ids[i]
		if res == nil {
			// Key does not exist
			items[id] = nil
			continue
		}
		// res is of type interface{}, need to assert to string
		strRes, ok := res.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for key %s: %T", keys[i], res)
		}
		items[id] = &strRes
	}

	return items, nil
}

type TypedRedisCache[T any] struct {
	RedisCache
}

// getTypeName returns the lowercase name of the type T.

func NewTypedRedisCache[T any](client *redis.Client) TypedCache[T] {
	return &TypedRedisCache[T]{RedisCache{client, GetTypeName[T]()}}
}

func (r *TypedRedisCache[T]) Get(ctx context.Context, id string) (*T, error) {

	result, err := r.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	var item T
	if err := json.Unmarshal([]byte(*result), &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TypedRedisCache[T]) Set(ctx context.Context, key string, item T) error {
	encoded, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return r.SetRaw(ctx, key, string(encoded), 0)
}

func (r *TypedRedisCache[T]) GetMany(ctx context.Context, ids []string) (map[string]*T, error) {
	typedResults := make(map[string]*T, len(ids))
	result, err := r.GetManyRaw(ctx, ids)
	if err != nil {
		return nil, err
	}
	for id, item := range result {
		if item == nil {
			continue
		}
		var typed T
		if err := json.Unmarshal([]byte(*item), &item); err != nil {
			return nil, fmt.Errorf("error unmarshaling key %s: %v", id, err)
		}
		typedResults[id] = &typed
	}
	return typedResults, nil
}
