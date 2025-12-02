package gocache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoCache_GetRaw(t *testing.T) {
	cache := No()
	ctx := context.Background()

	result, err := cache.GetRaw(ctx, "any-key")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestNoCache_SetRaw(t *testing.T) {
	cache := No()
	ctx := context.Background()

	err := cache.SetRaw(ctx, "any-key", "any-value", time.Hour)
	require.NoError(t, err)

	result, err := cache.GetRaw(ctx, "any-key")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestNoCache_Delete(t *testing.T) {
	cache := No()
	ctx := context.Background()

	err := cache.Delete(ctx, "any-key")
	require.NoError(t, err)
}

func TestNoCache_GetManyRaw(t *testing.T) {
	cache := No()
	ctx := context.Background()

	results, err := cache.GetManyRaw(ctx, []string{"key1", "key2", "key3"})
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestNoCache_GetAllKeys(t *testing.T) {
	cache := No()
	ctx := context.Background()

	keys, err := cache.GetAllKeys(ctx)
	require.NoError(t, err)
	assert.Nil(t, keys)
}

func TestNoTypedCache_Get(t *testing.T) {
	type TestData struct {
		ID   string
		Name string
	}

	cache := NoTyped[TestData]()
	ctx := context.Background()

	result, err := cache.Get(ctx, "any-key")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestNoTypedCache_Set(t *testing.T) {
	type TestData struct {
		ID   string
		Name string
	}

	cache := NoTyped[TestData]()
	ctx := context.Background()

	data := TestData{ID: "1", Name: "Test"}
	err := cache.Set(ctx, "key1", data)
	require.NoError(t, err)

	result, err := cache.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestNoTypedCache_GetMany(t *testing.T) {
	type TestData struct {
		ID   string
		Name string
	}

	cache := NoTyped[TestData]()
	ctx := context.Background()

	results, err := cache.GetMany(ctx, []string{"key1", "key2", "key3"})
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestNoCache_AsInterface(t *testing.T) {
	var cache Cache = No()
	ctx := context.Background()

	err := cache.SetRaw(ctx, "key", "value", 0)
	require.NoError(t, err)

	result, err := cache.GetRaw(ctx, "key")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestNoTypedCache_AsInterface(t *testing.T) {
	type TestData struct {
		Value int
	}

	var cache TypedCache[TestData] = NoTyped[TestData]()
	ctx := context.Background()

	err := cache.Set(ctx, "key", TestData{Value: 42})
	require.NoError(t, err)

	result, err := cache.Get(ctx, "key")
	require.NoError(t, err)
	assert.Nil(t, result)

	results, err := cache.GetMany(ctx, []string{"key1", "key2"})
	require.NoError(t, err)
	assert.Nil(t, results)
}
