package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGatewayHandlerGetUserMsgQueueModeSupportsKiroOAuth(t *testing.T) {
	handler := &GatewayHandler{
		userMsgQueueHelper: &UserMsgQueueHelper{},
		cfg: &config.Config{
			Gateway: config.GatewayConfig{
				UserMessageQueue: config.UserMessageQueueConfig{Mode: config.UMQModeSerialize},
			},
		},
	}
	userMessage, err := service.ParseGatewayRequest(
		service.NewRequestBodyRef([]byte(`{"messages":[{"role":"user","content":"hello"}]}`)),
		service.PlatformAnthropic,
	)
	require.NoError(t, err)

	kiroOAuth := &service.Account{Platform: service.PlatformKiro, Type: service.AccountTypeOAuth}
	kiroAPIKey := &service.Account{Platform: service.PlatformKiro, Type: service.AccountTypeAPIKey}
	anthropicOAuth := &service.Account{Platform: service.PlatformAnthropic, Type: service.AccountTypeOAuth}

	require.Equal(t, config.UMQModeSerialize, handler.getUserMsgQueueMode(kiroOAuth, userMessage))
	require.Empty(t, handler.getUserMsgQueueMode(kiroAPIKey, userMessage))
	require.Equal(t, config.UMQModeSerialize, handler.getUserMsgQueueMode(anthropicOAuth, userMessage))
}
