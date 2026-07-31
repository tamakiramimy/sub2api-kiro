package continuation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const stateTTL = 24 * time.Hour

var ErrStoreUnavailable = errors.New("kiro continuation store unavailable")

type State struct {
	ConversationID      string `json:"conversation_id"`
	AgentContinuationID string `json:"agent_continuation_id"`
}

func (s State) Valid() bool {
	return strings.TrimSpace(s.ConversationID) != "" && strings.TrimSpace(s.AgentContinuationID) != ""
}

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

func RedisKey(scope string) string {
	digest := sha256.Sum256([]byte(scope))
	return "kiro:continuation:" + hex.EncodeToString(digest[:])
}

func (s *Store) Get(ctx context.Context, scope string) (*State, error) {
	if s == nil || s.client == nil {
		return nil, ErrStoreUnavailable
	}
	if strings.TrimSpace(scope) == "" {
		return nil, nil
	}
	value, err := s.client.Get(ctx, RedisKey(scope)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get kiro continuation: %w", err)
	}
	var state State
	if err := json.Unmarshal([]byte(value), &state); err != nil {
		return nil, fmt.Errorf("decode kiro continuation: %w", err)
	}
	if !state.Valid() {
		return nil, nil
	}
	return &state, nil
}

func (s *Store) Set(ctx context.Context, scope string, state State) error {
	if s == nil || s.client == nil {
		return ErrStoreUnavailable
	}
	if strings.TrimSpace(scope) == "" || !state.Valid() {
		return nil
	}
	value, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode kiro continuation: %w", err)
	}
	if err := s.client.Set(ctx, RedisKey(scope), value, stateTTL).Err(); err != nil {
		return fmt.Errorf("set kiro continuation: %w", err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, scope string) error {
	if s == nil || s.client == nil {
		return ErrStoreUnavailable
	}
	if strings.TrimSpace(scope) == "" {
		return nil
	}
	if err := s.client.Del(ctx, RedisKey(scope)).Err(); err != nil {
		return fmt.Errorf("delete kiro continuation: %w", err)
	}
	return nil
}
