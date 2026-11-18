package kiro

import "strings"

func AdaptPlatformQuotaMigrationSQL(name, content string) string {
	switch name {
	case "241_add_typesafe_platform.sql", "241_user_platform_quotas_add_kiro.sql", "242_user_platform_quotas_kiro_typesafe.sql":
		// Current releases validate platforms through domain.Platforms. Keep
		// legacy migration checksums and DROP statements, and skip the obsolete
		// ADD so replays can run over both Kiro and newer provider rows.
		const addQuotaCheck = "ALTER TABLE user_platform_quotas\n    ADD CONSTRAINT user_platform_quotas_platform_check\n"
		before, after, ok := strings.Cut(content, addQuotaCheck)
		if ok {
			_, rest, complete := strings.Cut(after, ";")
			if complete {
				return before + rest
			}
		}
	}
	return content
}
