package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const kiroUsageSessionFingerprintDomain = "sub2api:kiro-usage-session:v1:"

// KiroUsageSessionFingerprint derives a stable, non-reversible audit identifier.
func (s *GatewayService) KiroUsageSessionFingerprint(sessionHash string) string {
	sessionHash = strings.TrimSpace(sessionHash)
	if s == nil || s.cfg == nil || sessionHash == "" || strings.TrimSpace(s.cfg.JWT.Secret) == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.JWT.Secret))
	_, _ = mac.Write([]byte(kiroUsageSessionFingerprintDomain))
	_, _ = mac.Write([]byte(sessionHash))
	return hex.EncodeToString(mac.Sum(nil))
}