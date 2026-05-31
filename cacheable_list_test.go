package gocache

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	data      map[string]testEntity
	callCount int
	err       error
}

type testEntity struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (m *mockRepository) GetByIdList(ctx context.Context, ids []string) (map[string]testEntity, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}

	result := make(map[string]testEntity)
	for _, id := range ids {
		if entity, ok := m.data[id]; ok {
			result[id] = entity
		}
	}
	return result, nil
}

func TestGetCacheableList_AllInCache(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	entity1 := testEntity{ID: "1", Name: "Entity 1", Value: 100}
	entity2 := testEntity{ID: "2", Name: "Entity 2", Value: 200}

	err := cache.Set(ctx, "1", entity1)
	require.NoError(t, err)
	err = cache.Set(ctx, "2", entity2)
	require.NoError(t, err)

	repo := &mockRepository{
		data: map[string]testEntity{},
	}

	results, err := GetCacheableList(ctx, []string{"1", "2"}, repo, cache)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 0, repo.callCount, "repository should not be called when all items are in cache")

	assert.Equal(t, entity1.Name, results["1"].Name)
	assert.Equal(t, entity2.Name, results["2"].Name)
}

func TestGetCacheableList_NoneInCache(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	entity1 := testEntity{ID: "1", Name: "Entity 1", Value: 100}
	entity2 := testEntity{ID: "2", Name: "Entity 2", Value: 200}

	repo := &mockRepository{
		data: map[string]testEntity{
			"1": entity1,
			"2": entity2,
		},
	}

	results, err := GetCacheableList(ctx, []string{"1", "2"}, repo, cache)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, repo.callCount, "repository should be called once")

	assert.Equal(t, entity1.Name, results["1"].Name)
	assert.Equal(t, entity2.Name, results["2"].Name)

	cached1, err := cache.Get(ctx, "1")
	require.NoError(t, err)
	require.NotNil(t, cached1)
	assert.Equal(t, entity1.Name, cached1.Name)

	cached2, err := cache.Get(ctx, "2")
	require.NoError(t, err)
	require.NotNil(t, cached2)
	assert.Equal(t, entity2.Name, cached2.Name)
}

func TestGetCacheableList_PartialCache(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	entity1 := testEntity{ID: "1", Name: "Cached Entity", Value: 100}
	entity2 := testEntity{ID: "2", Name: "Repo Entity", Value: 200}

	err := cache.Set(ctx, "1", entity1)
	require.NoError(t, err)

	repo := &mockRepository{
		data: map[string]testEntity{
			"2": entity2,
		},
	}

	results, err := GetCacheableList(ctx, []string{"1", "2"}, repo, cache)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, repo.callCount, "repository should be called once for missing items")

	assert.Equal(t, entity1.Name, results["1"].Name)
	assert.Equal(t, entity2.Name, results["2"].Name)

	cached2, err := cache.Get(ctx, "2")
	require.NoError(t, err)
	require.NotNil(t, cached2)
	assert.Equal(t, entity2.Name, cached2.Name)
}

func TestGetCacheableList_RepositoryError(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	repo := &mockRepository{
		data: map[string]testEntity{},
		err:  errors.New("repository error"),
	}

	results, err := GetCacheableList(ctx, []string{"1", "2"}, repo, cache)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestGetCacheableList_PartialRepositoryError(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	entity1 := testEntity{ID: "1", Name: "Cached Entity", Value: 100}
	err := cache.Set(ctx, "1", entity1)
	require.NoError(t, err)

	repo := &mockRepository{
		data: map[string]testEntity{},
		err:  errors.New("repository error"),
	}

	results, err := GetCacheableList(ctx, []string{"1", "2"}, repo, cache)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, entity1.Name, results["1"].Name)
}

func TestGetCacheableList_EmptyIDs(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	repo := &mockRepository{
		data: map[string]testEntity{},
	}

	results, err := GetCacheableList(ctx, []string{}, repo, cache)
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 0, repo.callCount)
}

func TestGetCacheableList_SomeNotFoundInRepository(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()
	cache := NewTypedRedisCache[testEntity](client, "")

	entity1 := testEntity{ID: "1", Name: "Entity 1", Value: 100}

	repo := &mockRepository{
		data: map[string]testEntity{
			"1": entity1,
		},
	}

	results, err := GetCacheableList(ctx, []string{"1", "2", "3"}, repo, cache)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, entity1.Name, results["1"].Name)

	_, exists2 := results["2"]
	assert.False(t, exists2)
	_, exists3 := results["3"]
	assert.False(t, exists3)
}
