package kiro

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultModels_MatchesKiroReferenceModelsPlusGPTExtension(t *testing.T) {
	ids := make([]string, 0, len(DefaultModels))
	for _, model := range DefaultModels {
		ids = append(ids, model.ID)
	}

	// 基线为 myfork 参考实现的 16 个 Claude 模型，额外主动扩展了 Kiro 近期新增的 3 个
	// OpenAI GPT-5.6 代理模型（sol/terra/luna）——myfork 回退时还没有这 3 个模型。
	require.Equal(t, []string{
		"claude-opus-4-8",
		"claude-opus-4-8-thinking",
		"claude-opus-4-7",
		"claude-opus-4-7-thinking",
		"claude-sonnet-5-0",
		"claude-sonnet-5-0-thinking",
		"claude-opus-4-6",
		"claude-opus-4-6-thinking",
		"claude-sonnet-4-6",
		"claude-sonnet-4-6-thinking",
		"claude-opus-4-5-20251101",
		"claude-opus-4-5-20251101-thinking",
		"claude-sonnet-4-5-20250929",
		"claude-sonnet-4-5-20250929-thinking",
		"claude-haiku-4-5-20251001",
		"claude-haiku-4-5-20251001-thinking",
		"gpt-5.6-sol",
		"gpt-5.6-terra",
		"gpt-5.6-luna",
	}, ids)

	require.Contains(t, ids, "claude-sonnet-4-6")
	require.Contains(t, ids, "claude-opus-4-7")
	require.Contains(t, ids, "claude-opus-4-8")
	require.Contains(t, ids, "claude-sonnet-5-0")
	require.Contains(t, ids, "claude-haiku-4-5-20251001-thinking")
	require.Contains(t, ids, "gpt-5.6-sol")
	require.Contains(t, ids, "gpt-5.6-terra")
	require.Contains(t, ids, "gpt-5.6-luna")
	require.NotContains(t, ids, "auto")
	require.NotContains(t, ids, "claude-sonnet-4")
	require.NotContains(t, ids, "gpt-4o")
	require.NotContains(t, ids, "deepseek-3-2")
	require.NotContains(t, ids, "minimax-m2-1")
	require.NotContains(t, ids, "qwen3-coder-next")
	require.NotContains(t, ids, "claude-sonnet-4-6-chat")
	for _, id := range ids {
		require.NotContains(t, id, "kiro-")
		require.NotContains(t, id, "-agentic")
		require.NotContains(t, id, "-chat")
	}
}
