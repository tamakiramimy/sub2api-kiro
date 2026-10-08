package kiro

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestAdaptPlatformQuotaMigrationSQL(t *testing.T) {
	for _, name := range []string{"241_add_typesafe_platform.sql", "241_user_platform_quotas_add_kiro.sql"} {
		body, err := migrations.FS.ReadFile(name)
		require.NoError(t, err)
		content := string(body)
		adapted := AdaptPlatformQuotaMigrationSQL(name, content)
		require.Contains(t, adapted, "'opencode_go', 'typesafe', 'kiro'))")
		require.Equal(t, adapted, AdaptPlatformQuotaMigrationSQL(name, adapted))
		require.Equal(t, content, AdaptPlatformQuotaMigrationSQL("unknown.sql", content))
		if name == "241_add_typesafe_platform.sql" {
			require.Contains(t, adapted, "CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok',\n                               'kimi', 'zhipu', 'deepseek', 'minimax', 'opencode_go', 'typesafe'))")
		}
	}
}
