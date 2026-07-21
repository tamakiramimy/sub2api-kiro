/**
 * Kiro 平台默认支持的模型 ID 列表。
 *
 * 与后端 `backend/internal/domain/constants.go` 的 `DefaultKiroModelMapping` 键保持一致。
 * `composables/useModelWhitelist.ts` 只导入并展开合并进全局模型下拉列表，不直接维护具体条目。
 */
export const kiroModels = [
  'claude-opus-4-8',
  'claude-opus-4-8-thinking',
  'claude-opus-4-7',
  'claude-opus-4-7-thinking',
  'claude-sonnet-5-0',
  'claude-sonnet-5-0-thinking',
  'claude-opus-4-6',
  'claude-opus-4-6-thinking',
  'claude-sonnet-4-6',
  'claude-sonnet-4-6-thinking',
  'claude-opus-4-5-20251101',
  'claude-opus-4-5-20251101-thinking',
  'claude-sonnet-4-5-20250929',
  'claude-sonnet-4-5-20250929-thinking',
  'claude-haiku-4-5-20251001',
  'claude-haiku-4-5-20251001-thinking'
]

/** Kiro 模型映射预设（用于账号编辑弹窗的"一键添加"按钮），键为对外模型名，值为 Kiro 上游模型名。 */
export const kiroPresetMappings: Array<{ from: string; to: string }> = [
  { from: 'claude-opus-4-8', to: 'claude-opus-4.8' },
  { from: 'claude-opus-4-7', to: 'claude-opus-4.7' },
  { from: 'claude-sonnet-5-0', to: 'claude-sonnet-5.0' },
  { from: 'claude-opus-4-6', to: 'claude-opus-4.6' },
  { from: 'claude-sonnet-4-6', to: 'claude-sonnet-4.6' },
  { from: 'claude-opus-4-5-20251101', to: 'claude-opus-4.5' },
  { from: 'claude-sonnet-4-5-20250929', to: 'claude-sonnet-4.5' },
  { from: 'claude-haiku-4-5-20251001', to: 'claude-haiku-4.5' }
]
