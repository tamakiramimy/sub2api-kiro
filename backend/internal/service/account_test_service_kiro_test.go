//go:build unit

package service

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountTestService_KiroUsesAmazonQRequestAndParsesResponse(t *testing.T) {
	ctx, recorder := newTestContext()
	account := &Account{
		ID:          91,
		Platform:    PlatformKiro,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":  "kiro-access-token",
			"refresh_token": "kiro-refresh-token",
			"expires_at":    time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			"auth_method":   "idc",
			"provider":      "AWS",
			"model_mapping": map[string]any{
				"gpt-5.6-terra": "gpt-5.6-terra",
			},
		},
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(kiroBridgeSuccessEventStream(t))),
		},
	}}
	svc := &AccountTestService{
		accountRepo:       repo,
		kiroTokenProvider: NewKiroTokenProvider(repo, nil, nil),
		httpUpstream:      upstream,
	}

	require.NoError(t, svc.TestAccountConnection(ctx, account.ID, "gpt-5.6-terra", "", AccountTestModeDefault))
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://q.us-east-1.amazonaws.com/generateAssistantResponse", upstream.requests[0].URL.String())
	require.Equal(t, "Bearer kiro-access-token", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "vibe", upstream.requests[0].Header.Get("x-amzn-kiro-agent-mode"))
	body, err := io.ReadAll(upstream.requests[0].Body)
	require.NoError(t, err)
	require.Equal(t, "gpt-5.6-terra", gjson.GetBytes(body, "conversationState.currentMessage.userInputMessage.modelId").String())
	require.Contains(t, recorder.Body.String(), "Hello from Kiro GPT")
	require.Contains(t, recorder.Body.String(), `"type":"test_complete"`)
}

func TestAccountTestService_KiroNormalizesLegacySonnet5Mapping(t *testing.T) {
	ctx, recorder := newTestContext()
	account := &Account{
		ID:          93,
		Platform:    PlatformKiro,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":  "kiro-access-token",
			"refresh_token": "kiro-refresh-token",
			"expires_at":    time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			"auth_method":   "idc",
			"provider":      "AWS",
			"model_mapping": map[string]any{
				"claude-sonnet-5-0": "claude-sonnet-5.0",
			},
		},
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(kiroBridgeSuccessEventStream(t))),
		},
	}}
	svc := &AccountTestService{
		accountRepo:       repo,
		kiroTokenProvider: NewKiroTokenProvider(repo, nil, nil),
		httpUpstream:      upstream,
	}

	require.NoError(t, svc.TestAccountConnection(ctx, account.ID, "claude-sonnet-5-0", "", AccountTestModeDefault))
	require.Len(t, upstream.requests, 1)
	body, err := io.ReadAll(upstream.requests[0].Body)
	require.NoError(t, err)
	require.Equal(t, "claude-sonnet-5", gjson.GetBytes(body, "conversationState.currentMessage.userInputMessage.modelId").String())
	require.Contains(t, recorder.Body.String(), `"type":"test_complete"`)
}

func TestAccountTestService_KiroRefreshesAndRetriesAfterUnauthorized(t *testing.T) {
	ctx, recorder := newTestContext()
	account := &Account{
		ID:          92,
		Platform:    PlatformKiro,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":  "expired-access-token",
			"refresh_token": "refresh-token",
			"expires_at":    time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			"auth_method":   "idc",
			"provider":      "AWS",
		},
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	upstream := &queuedHTTPUpstream{responses: []*http.Response{
		newJSONResponse(http.StatusUnauthorized, `{"type":"error","error":{"message":"Invalid bearer token"}}`),
		{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(kiroBridgeSuccessEventStream(t))),
		},
	}}
	provider := NewKiroTokenProvider(repo, nil, nil)
	provider.kiroOAuthService = &stubKiroAccountTokenRefresher{tokenInfo: &KiroTokenInfo{
		AccessToken: "refreshed-access-token",
		ExpiresAt:   time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	}}
	svc := &AccountTestService{
		accountRepo:       repo,
		kiroTokenProvider: provider,
		httpUpstream:      upstream,
	}

	require.NoError(t, svc.TestAccountConnection(ctx, account.ID, "claude-sonnet-4-6", "", AccountTestModeDefault))
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "Bearer expired-access-token", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "Bearer refreshed-access-token", upstream.requests[1].Header.Get("Authorization"))
	require.Contains(t, recorder.Body.String(), `"type":"test_complete"`)
}