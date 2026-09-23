//go:build unit

package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestKiroUsageSessionFingerprintIsStableAndNonReversible(t *testing.T) {
	t.Parallel()

	svc := &GatewayService{cfg: &config.Config{JWT: config.JWTConfig{Secret: "test-jwt-secret"}}}
	first := svc.KiroUsageSessionFingerprint("stable-session-input")

	require.Len(t, first, 64)
	require.Equal(t, first, svc.KiroUsageSessionFingerprint("stable-session-input"))
	require.NotEqual(t, first, svc.KiroUsageSessionFingerprint("other-session-input"))
	require.NotContains(t, first, "stable-session-input")
	require.Empty(t, svc.KiroUsageSessionFingerprint(""))
}