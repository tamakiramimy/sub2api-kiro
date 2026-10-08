//go:build integration

package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	kiropkg "github.com/Wei-Shaw/sub2api/internal/kiro"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestKiroTypeSafeQuotaMigrations(t *testing.T) {
	readSQL := func(name string) string {
		body, err := migrations.FS.ReadFile(name)
		require.NoError(t, err)
		return strings.TrimSpace(string(body))
	}
	for _, testCase := range []struct {
		name, appliedMigration, platform string
	}{
		{name: "fresh"},
		{name: "kiro_upgrade", appliedMigration: "241_user_platform_quotas_add_kiro.sql", platform: "kiro"},
		{name: "typesafe_upgrade", appliedMigration: "241_add_typesafe_platform.sql", platform: "typesafe"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tx := testTx(t)
			ctx := t.Context()
			_, err := tx.ExecContext(ctx, `CREATE TEMP TABLE user_platform_quotas (platform TEXT NOT NULL) ON COMMIT DROP;
CREATE TEMP TABLE composite_model_routes (target_platform TEXT NOT NULL) ON COMMIT DROP;`)
			require.NoError(t, err)
			if testCase.appliedMigration != "" {
				_, err = tx.ExecContext(ctx, readSQL(testCase.appliedMigration))
				require.NoError(t, err)
				_, err = tx.ExecContext(ctx, "INSERT INTO user_platform_quotas (platform) VALUES ($1)", testCase.platform)
				require.NoError(t, err)
			}
			for _, name := range []string{"241_add_typesafe_platform.sql", "241_user_platform_quotas_add_kiro.sql"} {
				if name != testCase.appliedMigration {
					_, err = tx.ExecContext(ctx, kiropkg.AdaptPlatformQuotaMigrationSQL(name, readSQL(name)))
					require.NoError(t, err)
				}
			}
			_, err = tx.ExecContext(ctx, readSQL("242_user_platform_quotas_kiro_typesafe.sql"))
			require.NoError(t, err)
			_, err = tx.ExecContext(ctx, "INSERT INTO user_platform_quotas (platform) VALUES ('kiro'), ('typesafe')")
			require.NoError(t, err)
			if testCase.platform != "" {
				var count int
				require.NoError(t, tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_platform_quotas WHERE platform = $1", testCase.platform).Scan(&count))
				require.Equal(t, 2, count)
			}
		})
	}
	for _, name := range []string{"241_add_typesafe_platform.sql", "241_user_platform_quotas_add_kiro.sql"} {
		var stored string
		require.NoError(t, integrationDB.QueryRowContext(t.Context(), "SELECT checksum FROM schema_migrations WHERE filename = $1", name).Scan(&stored))
		sum := sha256.Sum256([]byte(readSQL(name)))
		require.Equal(t, hex.EncodeToString(sum[:]), stored)
	}
}
