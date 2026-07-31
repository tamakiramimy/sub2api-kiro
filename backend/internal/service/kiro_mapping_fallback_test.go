package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAccountKiroDefaultMappingRestrictsUnsupportedModels(t *testing.T) {
	account := &Account{Platform: PlatformKiro}

	require.False(t, account.IsModelSupported("gpt-4o"))
	require.False(t, account.IsModelSupported("kiro-gpt-4o"))
	require.False(t, account.IsModelSupported("auto"))
	require.True(t, account.IsModelSupported("claude-opus-5"))
	require.True(t, account.IsModelSupported("claude-opus-5-thinking"))
	require.True(t, account.IsModelSupported("claude-sonnet-5"))
	require.True(t, account.IsModelSupported("claude-sonnet-5-thinking"))
	require.Equal(t, "claude-opus-5", account.GetMappedModel("claude-opus-5-thinking"))
	require.Equal(t, "claude-sonnet-5", account.GetMappedModel("claude-sonnet-5-thinking"))
	require.Equal(t, "claude-sonnet-4.6", account.GetMappedModel("claude-sonnet-4-6"))
}

func TestAccountKiroLegacyClaude5MappingsExposeCanonicalClientIDs(t *testing.T) {
	account := &Account{
		Platform: PlatformKiro,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"claude-opus-5-0":            "claude-opus-5.0",
				"claude-opus-5-0-thinking":   "claude-opus-5.0",
				"claude-sonnet-5-0":          "claude-sonnet-5.0",
				"claude-sonnet-5-0-thinking": "claude-sonnet-5.0",
			},
		},
	}

	for _, model := range []string{
		"claude-opus-5",
		"claude-opus-5-thinking",
		"claude-sonnet-5",
		"claude-sonnet-5-thinking",
	} {
		require.True(t, account.IsModelSupported(model), model)
	}
	require.Equal(t, "claude-opus-5", account.GetMappedModel("claude-opus-5-thinking"))
	require.Equal(t, "claude-sonnet-5", account.GetMappedModel("claude-sonnet-5-thinking"))
}

func TestGatewayServiceCalculateTokenCost_KiroAutoUsesConservativeFallback(t *testing.T) {
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1.1

	svc := NewGatewayService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		nil,
		NewBillingService(cfg, nil),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil, // channelService
		nil, // resolver
		nil, // balanceNotifyService
		nil, // userPlatformQuotaRepo
	)

	result := &ForwardResult{
		Model:         "auto",
		UpstreamModel: "auto",
		Usage: ClaudeUsage{
			InputTokens:  20,
			OutputTokens: 10,
		},
	}

	expected, err := svc.billingService.CalculateCost(kiroConservativeFallbackBillingModel, UsageTokens{
		InputTokens:  20,
		OutputTokens: 10,
	}, 1.1)
	require.NoError(t, err)

	cost := svc.calculateTokenCost(context.Background(), result, &APIKey{}, "auto", 1.1, &recordUsageOpts{IsKiroAccount: true})
	require.NotNil(t, cost)
	require.InDelta(t, expected.ActualCost, cost.ActualCost, 1e-12)
	require.InDelta(t, expected.TotalCost, cost.TotalCost, 1e-12)
}
