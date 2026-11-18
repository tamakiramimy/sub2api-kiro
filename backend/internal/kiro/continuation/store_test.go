package continuation

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestStoreRoundTripExpiresState(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	store := NewStore(client)
	scope := "account=12|group=3|api_key=101|session=desktop-session-1"
	state := State{
		ConversationID:      "conversation-1",
		AgentContinuationID: "continuation-1",
	}

	require.NoError(t, store.Set(context.Background(), scope, state))
	require.True(t, server.Exists(RedisKey(scope)))
	require.LessOrEqual(t, server.TTL(RedisKey(scope)), stateTTL)
	require.Greater(t, server.TTL(RedisKey(scope)), stateTTL-time.Second)

	actual, err := store.Get(context.Background(), scope)
	require.NoError(t, err)
	require.Equal(t, &state, actual)

	server.FastForward(stateTTL)
	actual, err = store.Get(context.Background(), scope)
	require.NoError(t, err)
	require.Nil(t, actual)
}

func TestRedisKeyHashesContinuationScope(t *testing.T) {
	scope := "account=12|group=3|api_key=101|session=desktop-session-1"
	key := RedisKey(scope)

	require.Equal(t, RedisKey(scope), key)
	require.NotEqual(t, RedisKey(scope+"-other"), key)
	require.NotContains(t, key, "desktop-session-1")
	require.NotContains(t, key, "api_key=101")
}
