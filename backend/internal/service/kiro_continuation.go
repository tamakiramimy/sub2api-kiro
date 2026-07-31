package service

import (
	"context"
	"fmt"
	"strings"

	kiropkg "github.com/Wei-Shaw/sub2api/internal/kiro"
	kirocontinuation "github.com/Wei-Shaw/sub2api/internal/kiro/continuation"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// KiroContinuationStore persists upstream agent state between Claude Desktop
// turns. The scope includes the authenticated API key and selected account so
// continuation IDs cannot cross tenant or credential boundaries.
type KiroContinuationStore interface {
	Get(context.Context, string) (*kirocontinuation.State, error)
	Set(context.Context, string, kirocontinuation.State) error
	Delete(context.Context, string) error
}

func (s *GatewayService) kiroContinuationScope(account *Account, parsed *ParsedRequest) string {
	if s == nil || account == nil || parsed == nil || parsed.SessionContext == nil {
		return ""
	}
	if parsed.SessionContext.APIKeyID <= 0 {
		return ""
	}
	sessionHash := strings.TrimSpace(s.GenerateSessionHash(parsed))
	if sessionHash == "" {
		return ""
	}
	return fmt.Sprintf("account=%d|group=%d|api_key=%d|session=%s",
		account.ID,
		derefGroupID(parsed.GroupID),
		parsed.SessionContext.APIKeyID,
		sessionHash,
	)
}

func (s *GatewayService) loadKiroContinuation(ctx context.Context, account *Account, parsed *ParsedRequest) *kiropkg.KiroContinuation {
	if s == nil || s.kiroContinuationStore == nil {
		return nil
	}
	scope := s.kiroContinuationScope(account, parsed)
	if scope == "" {
		return nil
	}
	state, err := s.kiroContinuationStore.Get(ctx, scope)
	if err != nil || state == nil || !state.Valid() {
		if err != nil {
			logger.L().Debug("kiro continuation cache read failed")
		}
		return nil
	}
	return &kiropkg.KiroContinuation{
		ConversationID:      state.ConversationID,
		AgentContinuationID: state.AgentContinuationID,
	}
}

func (s *GatewayService) saveKiroContinuation(ctx context.Context, account *Account, parsed *ParsedRequest, continuation kiropkg.KiroContinuation) {
	if s == nil || s.kiroContinuationStore == nil || !continuation.Valid() {
		return
	}
	scope := s.kiroContinuationScope(account, parsed)
	if scope == "" {
		return
	}
	if err := s.kiroContinuationStore.Set(ctx, scope, kirocontinuation.State{
		ConversationID:      continuation.ConversationID,
		AgentContinuationID: continuation.AgentContinuationID,
	}); err != nil {
		logger.L().Debug("kiro continuation cache write failed")
	}
}
