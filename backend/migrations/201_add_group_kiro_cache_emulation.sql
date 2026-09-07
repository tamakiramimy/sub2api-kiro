ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS kiro_cache_emulation_enabled BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS kiro_cache_emulation_ratio NUMERIC(4,3) NOT NULL DEFAULT 1.0;

COMMENT ON COLUMN groups.kiro_cache_emulation_enabled IS '是否为该分组启用 Kiro prompt cache 计费模拟（仅 kiro 平台生效）';
COMMENT ON COLUMN groups.kiro_cache_emulation_ratio IS 'Kiro cache 命中计费模拟比例，取值 [0,1]，仅在 kiro_cache_emulation_enabled 为 true 时生效';
