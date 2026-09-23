//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestBuildKiroPayloadForAccountUsesStableConversationIDs(t *testing.T) {
	t.Setenv("SUB2API_KIRO_CONVERSATION_ID_MODE", "")
	svc := &GatewayService{}
	account := &Account{ID: 40, Credentials: map[string]any{"profile_arn": "profile-a"}}
	body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello","additional_kwargs":{"conversationId":"client-conv","continuationId":"client-cont"}}]}`)

	first, err := svc.buildKiroPayloadForAccount(context.Background(), account, nil, body, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)
	second, err := svc.buildKiroPayloadForAccount(context.Background(), account, nil, body, "claude-sonnet-4.5", "rotated-token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)

	firstConversationID := gjson.GetBytes(first.Payload, "conversationState.conversationId").String()
	secondConversationID := gjson.GetBytes(second.Payload, "conversationState.conversationId").String()
	require.NotEmpty(t, firstConversationID)
	require.Equal(t, firstConversationID, secondConversationID)
	require.NotEqual(t, "client-conv", firstConversationID)
	require.False(t, gjson.GetBytes(first.Payload, "conversationState.agentContinuationId").Exists())
	require.False(t, gjson.GetBytes(second.Payload, "conversationState.agentContinuationId").Exists())

	t.Setenv("SUB2API_KIRO_CONVERSATION_ID_MODE", "random")
	randomized, err := svc.buildKiroPayloadForAccount(context.Background(), account, nil, body, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)
	require.NotEqual(t, firstConversationID, gjson.GetBytes(randomized.Payload, "conversationState.conversationId").String())
}

func TestBuildKiroPayloadForAccountKeepsConversationIDAcrossTurns(t *testing.T) {
	t.Setenv("SUB2API_KIRO_CONVERSATION_ID_MODE", "")
	svc := &GatewayService{}
	account := &Account{ID: 40, Credentials: map[string]any{"profile_arn": "profile-a"}}
	groupID := int64(3)
	sessionID := "0b4445e1-f5be-49e1-87ce-62bbc28ad705"
	firstBody := []byte(`{
		"model":"claude-sonnet-4-5",
		"metadata":{"user_id":"{\"session_id\":\"0b4445e1-f5be-49e1-87ce-62bbc28ad705\",\"device_id\":\"device-a\"}"},
		"messages":[{"role":"user","content":"first"}]
	}`)
	secondBody := []byte(`{
		"model":"claude-sonnet-4-5",
		"metadata":{"user_id":"{\"session_id\":\"0b4445e1-f5be-49e1-87ce-62bbc28ad705\",\"device_id\":\"device-a\"}"},
		"messages":[
			{"role":"user","content":"first"},
			{"role":"assistant","content":"answer"},
			{"role":"user","content":"second"}
		]
	}`)

	firstParsed, err := ParseGatewayRequest(NewRequestBodyRef(firstBody), "anthropic")
	require.NoError(t, err)
	firstParsed.GroupID = &groupID
	firstParsed.SessionContext = &SessionContext{APIKeyID: 9}
	secondParsed, err := ParseGatewayRequest(NewRequestBodyRef(secondBody), "anthropic")
	require.NoError(t, err)
	secondParsed.GroupID = &groupID
	secondParsed.SessionContext = &SessionContext{APIKeyID: 9}

	first, err := svc.buildKiroPayloadForAccount(context.Background(), account, firstParsed, firstBody, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)
	second, err := svc.buildKiroPayloadForAccount(context.Background(), account, secondParsed, secondBody, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)

	require.Equal(t, sessionID, gjson.GetBytes(first.Payload, "conversationState.conversationId").String())
	require.Equal(t, sessionID, gjson.GetBytes(second.Payload, "conversationState.conversationId").String())

	otherBody := []byte(`{
		"model":"claude-sonnet-4-5",
		"metadata":{"user_id":"{\"session_id\":\"cc62cfc3-6511-47e3-a2b5-20d7f5768245\",\"device_id\":\"device-a\"}"},
		"messages":[{"role":"user","content":"first"}]
	}`)
	otherParsed, err := ParseGatewayRequest(NewRequestBodyRef(otherBody), "anthropic")
	require.NoError(t, err)
	otherParsed.GroupID = &groupID
	otherParsed.SessionContext = &SessionContext{APIKeyID: 9}
	other, err := svc.buildKiroPayloadForAccount(context.Background(), account, otherParsed, otherBody, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)
	require.NotEqual(t,
		gjson.GetBytes(first.Payload, "conversationState.conversationId").String(),
		gjson.GetBytes(other.Payload, "conversationState.conversationId").String(),
	)
}

func TestExtractKiroConversationID(t *testing.T) {
	const sessionID = "0b4445e1-f5be-49e1-87ce-62bbc28ad705"
	deviceID := strings.Repeat("a", 64)

	tests := []struct {
		name     string
		userID   string
		expected string
	}{
		{name: "json", userID: `{"device_id":"device-a","session_id":"` + sessionID + `"}`, expected: sessionID},
		{name: "legacy", userID: "user_" + deviceID + "_account__session_" + sessionID, expected: sessionID},
		{name: "invalid uuid", userID: `{"session_id":"session-a"}`},
		{name: "nil uuid", userID: `{"session_id":"00000000-0000-0000-0000-000000000000"}`},
		{name: "missing", userID: `{"device_id":"device-a"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, extractKiroConversationID(test.userID))
		})
	}
}

func TestBuildKiroPayloadForAccountFallsBackForInvalidSessionID(t *testing.T) {
	t.Setenv("SUB2API_KIRO_CONVERSATION_ID_MODE", "")
	svc := &GatewayService{}
	account := &Account{ID: 40, Credentials: map[string]any{"profile_arn": "profile-a"}}
	body := []byte(`{
		"model":"claude-sonnet-4-5",
		"metadata":{"user_id":"{\"session_id\":\"session-a\"}"},
		"system":"system prompt",
		"messages":[{"role":"user","content":"first"}]
	}`)
	parsed, err := ParseGatewayRequest(NewRequestBodyRef(body), "anthropic")
	require.NoError(t, err)

	first, err := svc.buildKiroPayloadForAccount(context.Background(), account, parsed, body, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)
	second, err := svc.buildKiroPayloadForAccount(context.Background(), account, parsed, body, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)

	firstID := gjson.GetBytes(first.Payload, "conversationState.conversationId").String()
	require.NotEmpty(t, firstID)
	require.NotEqual(t, "session-a", firstID)
	require.Equal(t, firstID, gjson.GetBytes(second.Payload, "conversationState.conversationId").String())
}

func TestBuildKiroPayloadForAccountSeparatesFallbackConversations(t *testing.T) {
	t.Setenv("SUB2API_KIRO_CONVERSATION_ID_MODE", "")
	svc := &GatewayService{}
	account := &Account{ID: 40, Credentials: map[string]any{"profile_arn": "profile-a"}}
	firstBody := []byte(`{
		"model":"claude-sonnet-4-5",
		"system":"shared system prompt",
		"messages":[{"role":"user","content":"first conversation"}]
	}`)
	continuedBody := []byte(`{
		"model":"claude-sonnet-4-5",
		"system":"shared system prompt",
		"messages":[
			{"role":"user","content":"first conversation"},
			{"role":"assistant","content":"answer"},
			{"role":"user","content":"continue"}
		]
	}`)
	otherBody := []byte(`{
		"model":"claude-sonnet-4-5",
		"system":"shared system prompt",
		"messages":[{"role":"user","content":"different conversation"}]
	}`)

	build := func(body []byte) string {
		result, err := svc.buildKiroPayloadForAccount(context.Background(), account, nil, body, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
		require.NoError(t, err)
		return gjson.GetBytes(result.Payload, "conversationState.conversationId").String()
	}

	firstID := build(firstBody)
	require.Equal(t, firstID, build(continuedBody))
	require.NotEqual(t, firstID, build(otherBody))
}

func TestBuildKiroPayloadForAccountReplaysFullMessagesIntoHistory(t *testing.T) {
	svc := &GatewayService{}
	account := &Account{ID: 40, Credentials: map[string]any{"profile_arn": "profile-a"}}
	body := []byte(`{
		"model":"claude-sonnet-4-5",
		"system":"system prompt",
		"messages":[
			{"role":"user","content":"first"},
			{"role":"assistant","content":"answer"},
			{"role":"user","content":"second"}
		]
	}`)

	result, err := svc.buildKiroPayloadForAccount(context.Background(), account, nil, body, "claude-sonnet-4.5", "token", "claude-sonnet-4-5", nil)
	require.NoError(t, err)

	history := gjson.GetBytes(result.Payload, "conversationState.history").Array()
	require.Len(t, history, 4)
	require.Contains(t, history[0].Get("userInputMessage.content").String(), "system prompt")
	require.Equal(t, "first", history[2].Get("userInputMessage.content").String())
	require.Equal(t, "answer", history[3].Get("assistantResponseMessage.content").String())
	require.Equal(t, "second", gjson.GetBytes(result.Payload, "conversationState.currentMessage.userInputMessage.content").String())
}
