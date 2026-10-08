package kiro

import "strings"

func AdaptPlatformQuotaMigrationSQL(name, content string) string {
	const quotaCheckPrefix = "CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok',\n" +
		"                        'kimi', 'zhipu', 'deepseek', 'minimax', 'opencode_go', "
	switch name {
	case "241_add_typesafe_platform.sql":
		return strings.Replace(content, quotaCheckPrefix+"'typesafe'))", quotaCheckPrefix+"'typesafe', 'kiro'))", 1)
	case "241_user_platform_quotas_add_kiro.sql":
		return strings.Replace(content, quotaCheckPrefix+"'kiro'))", quotaCheckPrefix+"'typesafe', 'kiro'))", 1)
	default:
		return content
	}
}
