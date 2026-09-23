//go:build unit

package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountSummaryFromServiceIncludesPlatformWithoutCredentials(t *testing.T) {
	account := &service.Account{
		ID:          7,
		Name:        "kiro-account",
		Platform:    service.PlatformKiro,
		Credentials: map[string]any{"access_token": "secret"},
	}

	summary := AccountSummaryFromService(account)

	require.Equal(t, int64(7), summary.ID)
	require.Equal(t, "kiro-account", summary.Name)
	require.Equal(t, service.PlatformKiro, summary.Platform)
	fields := marshalToMap(t, summary)
	require.NotContains(t, fields, "credentials")
}