//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	kiropkg "github.com/Wei-Shaw/sub2api/internal/kiro"
	kirocontinuation "github.com/Wei-Shaw/sub2api/internal/kiro/continuation"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type memoryKiroContinuationStore struct {
	states map[string]kirocontinuation.State
}

func (s *memoryKiroContinuationStore) Get(_ context.Context, scope string) (*kirocontinuation.State, error) {
	state, ok := s.states[scope]
	if !ok {
		return nil, nil
	}
	return &state, nil
}

func (s *memoryKiroContinuationStore) Set(_ context.Context, scope string, state kirocontinuation.State) error {
	if s.states == nil {
		s.states = make(map[string]kirocontinuation.State)
	}
	s.states[scope] = state
	return nil
}

func (s *memoryKiroContinuationStore) Delete(_ context.Context, scope string) error {
	delete(s.states, scope)
	return nil
}

func kiroContinuationParsedRequest(t *testing.T, apiKeyID int64) *ParsedRequest {
	t.Helper()
	parsed, err := ParseGatewayRequest(NewRequestBodyRef([]byte(`{
		"model":"claude-opus-5",
		"metadata":{"user_id":"{\"session_id\":\"desktop-session-1\",\"device_id\":\"device-1\"}"},
		"messages":[{"role":"user","content":"continue the task"}]
	}`)), domain.PlatformAnthropic)
	require.NoError(t, err)
	groupID := int64(3)
	parsed.GroupID = &groupID
	parsed.SessionContext = &SessionContext{APIKeyID: apiKeyID, ClientIP: "10.0.0.1", UserAgent: "Claude/1.0"}
	return parsed
}

func TestKiroContinuationStoreDoesNotOverrideClientHistory(t *testing.T) {
	store := &memoryKiroContinuationStore{}
	svc := &GatewayService{kiroContinuationStore: store}
	account := &Account{ID: 12, Platform: PlatformKiro, Type: AccountTypeOAuth}
	parsed := kiroContinuationParsedRequest(t, 101)
	continuation := kiropkg.KiroContinuation{
		ConversationID:      "conversation-server-1",
		AgentContinuationID: "continuation-server-1",
	}

	svc.saveKiroContinuation(context.Background(), account, parsed, continuation)
	loaded := svc.loadKiroContinuation(context.Background(), account, parsed)
	require.Equal(t, &continuation, loaded)

	otherKeyParsed := kiroContinuationParsedRequest(t, 202)
	require.Nil(t, svc.loadKiroContinuation(context.Background(), account, otherKeyParsed))

	result, err := svc.buildKiroPayloadForAccount(
		context.Background(),
		account,
		parsed,
		parsed.Body.Bytes(),
		"claude-opus-5",
		"token",
		"claude-opus-5",
		nil,
	)
	require.NoError(t, err)
	require.NotEqual(t, continuation.ConversationID, gjson.GetBytes(result.Payload, "conversationState.conversationId").String())
	require.False(t, gjson.GetBytes(result.Payload, "conversationState.agentContinuationId").Exists())
	require.NotEmpty(t, gjson.GetBytes(result.Payload, "conversationState.history").Array())
}

func TestKiroContinuationScopeUsesStableClaudeDesktopSessionID(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{ID: 12, Platform: PlatformKiro, Type: AccountTypeOAuth}
	first := kiroContinuationParsedRequest(t, 101)
	second, err := ParseGatewayRequest(NewRequestBodyRef([]byte(`{
		"model":"claude-opus-5",
		"metadata":{"user_id":"{\"session_id\":\"desktop-session-1\",\"device_id\":\"device-1\"}"},
		"messages":[
			{"role":"user","content":"continue the task"},
			{"role":"assistant","content":"I inspected it."},
			{"role":"user","content":"now apply the change"}
		]
	}`)), domain.PlatformAnthropic)
	require.NoError(t, err)
	groupID := int64(3)
	second.GroupID = &groupID
	second.SessionContext = &SessionContext{APIKeyID: 101, ClientIP: "10.0.0.1", UserAgent: "Claude/1.0"}

	require.Equal(t, svc.kiroContinuationScope(account, first), svc.kiroContinuationScope(account, second))
}
