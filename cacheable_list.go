package gocache

import "context"

type ListRepository[T any] interface {
	GetByIdList(ctx context.Context, ids []string) (map[string]T, error)
}

func GetCacheableList[T any](ctx context.Context, ids []string, repository ListRepository[T], cache TypedCache[T]) (map[string]*T, error) {
	cacheMap, err := cache.GetMany(ctx, ids)
	var remaining []string
	if err != nil {
		remaining = ids
	} else {
		for id, value := range cacheMap {
			if value == nil {
				remaining = append(remaining, id)
			}
		}
	}
	if len(remaining) > 0 {
		objects, err := repository.GetByIdList(ctx, remaining)
		if err != nil && len(remaining) == len(ids) {
			return nil, err
		}
		for key, value := range objects {
			cacheMap[key] = &value
			_ = cache.Set(ctx, key, value)
		}
	}
	return cacheMap, nil
}
