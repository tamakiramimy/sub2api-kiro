-- Kiro 会话指纹仅用于管理员 usage 审计；不保存客户端会话原文或请求内容。
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS kiro_session_fingerprint VARCHAR(64);
