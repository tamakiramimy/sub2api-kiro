-- 仅索引 Kiro 行，支持管理员按会话回看账号切换且不阻塞热表写入。
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_kiro_session_fingerprint_created_at
    ON usage_logs (kiro_session_fingerprint, created_at, id)
    WHERE kiro_session_fingerprint IS NOT NULL;