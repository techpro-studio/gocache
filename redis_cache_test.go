package gocache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_SetRaw_GetRaw(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	t.Run("set and get existing key", func(t *testing.T) {
		key := "key1"
		value := "value1"

		err := cache.SetRaw(ctx, key, value, 0)
		require.NoError(t, err)

		result, err := cache.GetRaw(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, value, *result)
	})

	t.Run("get non-existing key", func(t *testing.T) {
		result, err := cache.GetRaw(ctx, "non-existing-key")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("set with expiration", func(t *testing.T) {
		key := "expiring-key"
		value := "expiring-value"

		err := cache.SetRaw(ctx, key, value, 500*time.Millisecond)
		require.NoError(t, err)

		result, err := cache.GetRaw(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, value, *result)

		time.Sleep(600 * time.Millisecond)

		result, err = cache.GetRaw(ctx, key)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestRedisCache_Delete(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	key := "key-to-delete"
	value := "value-to-delete"

	err := cache.SetRaw(ctx, key, value, 0)
	require.NoError(t, err)

	result, err := cache.GetRaw(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, result)

	err = cache.Delete(ctx, key)
	require.NoError(t, err)

	result, err = cache.GetRaw(ctx, key)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedisCache_GetManyRaw(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache := NewRedisCache(client, "test")
	ctx := context.Background()

	t.Run("get many with all existing keys", func(t *testing.T) {
		keys := []string{"many1", "many2", "many3"}
		values := map[string]string{
			"many1": "value1",
			"many2": "value2",
			"many3": "value3",
		}

		for k, v := range values {
			err := cache.SetRaw(ctx, k, v, 0)
			require.NoError(t, err)
		}

		results, err := cache.GetManyRaw(ctx, keys)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		for _, key := range keys {
			result, ok := results[key]
			assert.True(t, ok)
			require.NotNil(t, result)
			assert.Equal(t, values[key], *result)
		}
	})

	t.Run("get many with some non-existing keys", func(t *testing.T) {
		err := cache.SetRaw(ctx, "exists1", "val1", 0)
		require.NoError(t, err)

		keys := []string{"exists1", "not-exists1", "not-exists2"}
		results, err := cache.GetManyRaw(ctx, keys)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		require.NotNil(t, results["exists1"])
		assert.Equal(t, "val1", *results["exists1"])

		assert.Nil(t, results["not-exists1"])
		assert.Nil(t, results["not-exists2"])
	})

	t.Run("get many with empty slice", func(t *testing.T) {
		results, err := cache.GetManyRaw(ctx, []string{})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestRedisCache_GetAllKeys(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache := NewRedisCache(client, "test-prefix")
	ctx := context.Background()

	t.Run("get all keys when empty", func(t *testing.T) {
		keys, err := cache.GetAllKeys(ctx)
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("get all keys with data", func(t *testing.T) {
		testKeys := []string{"key1", "key2", "key3"}
		for _, key := range testKeys {
			err := cache.SetRaw(ctx, key, "value", 0)
			require.NoError(t, err)
		}

		keys, err := cache.GetAllKeys(ctx)
		require.NoError(t, err)
		assert.Len(t, keys, 3)
		assert.ElementsMatch(t, testKeys, keys)
	})

	t.Run("only gets keys with correct prefix", func(t *testing.T) {
		otherCache := NewRedisCache(client, "other-prefix")
		err := otherCache.SetRaw(ctx, "other-key", "value", 0)
		require.NoError(t, err)

		keys, err := cache.GetAllKeys(ctx)
		require.NoError(t, err)
		assert.NotContains(t, keys, "other-key")
	})
}

func TestTypedRedisCache_Get_Set(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	type TestStruct struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	cache := NewTypedRedisCache[TestStruct](client, "")
	ctx := context.Background()

	t.Run("set and get typed value", func(t *testing.T) {
		key := "user1"
		value := TestStruct{
			ID:   "1",
			Name: "John Doe",
			Age:  30,
		}

		err := cache.Set(ctx, key, value)
		require.NoError(t, err)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, value.ID, result.ID)
		assert.Equal(t, value.Name, result.Name)
		assert.Equal(t, value.Age, result.Age)
	})

	t.Run("get non-existing typed value", func(t *testing.T) {
		result, err := cache.Get(ctx, "non-existing")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestTypedRedisCache_GetMany(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	type Product struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	cache := NewTypedRedisCache[Product](client, "")
	ctx := context.Background()

	t.Run("get many typed values", func(t *testing.T) {
		products := map[string]Product{
			"prod1": {ID: "1", Name: "Laptop", Price: 999.99},
			"prod2": {ID: "2", Name: "Mouse", Price: 29.99},
			"prod3": {ID: "3", Name: "Keyboard", Price: 79.99},
		}

		for k, v := range products {
			err := cache.Set(ctx, k, v)
			require.NoError(t, err)
		}

		keys := []string{"prod1", "prod2", "prod3"}
		results, err := cache.GetMany(ctx, keys)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		for key, expected := range products {
			result, ok := results[key]
			assert.True(t, ok)
			require.NotNil(t, result)
			assert.Equal(t, expected.ID, result.ID)
			assert.Equal(t, expected.Name, result.Name)
			assert.Equal(t, expected.Price, result.Price)
		}
	})

	t.Run("get many with some missing values", func(t *testing.T) {
		err := cache.Set(ctx, "existing", Product{ID: "1", Name: "Test", Price: 10})
		require.NoError(t, err)

		keys := []string{"existing", "missing1", "missing2"}
		results, err := cache.GetMany(ctx, keys)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		_, existsOk := results["existing"]
		assert.True(t, existsOk)

		_, missing1Ok := results["missing1"]
		assert.False(t, missing1Ok)

		_, missing2Ok := results["missing2"]
		assert.False(t, missing2Ok)
	})
}

func TestTypedRedisCache_Prefix(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	type User struct {
		Name string `json:"name"`
	}

	cache := NewTypedRedisCache[User](client, "")
	ctx := context.Background()

	err := cache.Set(ctx, "test-user", User{Name: "Alice"})
	require.NoError(t, err)

	keys, err := client.Keys(ctx, "*").Result()
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "user:test-user", keys[0])
}
