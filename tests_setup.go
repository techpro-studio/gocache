package gocache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"testing"
)

func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	redisContainer, err := rediscontainer.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	cleanup := func() {
		client.Close()
		redisContainer.Terminate(ctx)
	}

	return client, cleanup
}
