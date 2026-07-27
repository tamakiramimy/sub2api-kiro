/** Kiro 默认模型映射；其键同时是白名单、分组和 API Key 对外可见的模型目录。 */
const kiroDefaultMappings: Array<{ from: string; to: string }> = [
  { from: 'claude-opus-5-0', to: 'claude-opus-5' },
  { from: 'claude-opus-5-0-thinking', to: 'claude-opus-5' },
  { from: 'claude-opus-4-8', to: 'claude-opus-4.8' },
  { from: 'claude-opus-4-8-thinking', to: 'claude-opus-4.8' },
  { from: 'claude-opus-4-7', to: 'claude-opus-4.7' },
  { from: 'claude-opus-4-7-thinking', to: 'claude-opus-4.7' },
	{ from: 'claude-sonnet-5-0', to: 'claude-sonnet-5' },
	{ from: 'claude-sonnet-5-0-thinking', to: 'claude-sonnet-5' },
  { from: 'claude-opus-4-6', to: 'claude-opus-4.6' },
  { from: 'claude-opus-4-6-thinking', to: 'claude-opus-4.6' },
  { from: 'claude-sonnet-4-6', to: 'claude-sonnet-4.6' },
  { from: 'claude-sonnet-4-6-thinking', to: 'claude-sonnet-4.6' },
  { from: 'claude-opus-4-5-20251101', to: 'claude-opus-4.5' },
  { from: 'claude-opus-4-5-20251101-thinking', to: 'claude-opus-4.5' },
  { from: 'claude-sonnet-4-5-20250929', to: 'claude-sonnet-4.5' },
  { from: 'claude-sonnet-4-5-20250929-thinking', to: 'claude-sonnet-4.5' },
  { from: 'claude-haiku-4-5-20251001', to: 'claude-haiku-4.5' },
  { from: 'claude-haiku-4-5-20251001-thinking', to: 'claude-haiku-4.5' },
  { from: 'gpt-5.6-sol', to: 'gpt-5.6-sol' },
  { from: 'gpt-5.6-terra', to: 'gpt-5.6-terra' },
  { from: 'gpt-5.6-luna', to: 'gpt-5.6-luna' }
]

/**
 * Kiro 平台默认支持的模型 ID 列表。
 *
 * 与后端 `backend/internal/domain/constants.go` 的 `DefaultKiroModelMapping` 键保持一致。
 * 从默认映射派生，避免白名单和模型映射的模型目录发生漂移。
 */
export const kiroModels = kiroDefaultMappings.map(({ from }) => from)

const kiroClaude5CanonicalModelIDs = new Map([
  ['claude-opus-5-0', 'claude-opus-5'],
  ['claude-opus-5-0-thinking', 'claude-opus-5'],
  ['claude-opus-5.0', 'claude-opus-5'],
  ['claude-opus-5', 'claude-opus-5'],
  ['claude-opus-5-thinking', 'claude-opus-5'],
  ['claude-sonnet-5-0', 'claude-sonnet-5'],
  ['claude-sonnet-5-0-thinking', 'claude-sonnet-5'],
  ['claude-sonnet-5.0', 'claude-sonnet-5'],
  ['claude-sonnet-5', 'claude-sonnet-5'],
  ['claude-sonnet-5-thinking', 'claude-sonnet-5']
])

export function sanitizeKiroModelMapping(rawMapping?: Record<string, unknown>): Record<string, unknown> {
  if (!rawMapping || typeof rawMapping !== 'object') return {}
  return Object.fromEntries(
    Object.entries(rawMapping).map(([model, target]) => {
      const canonical = kiroClaude5CanonicalModelIDs.get(model.trim())
      return [model, canonical ?? target]
    })
  )
}

/** Kiro 模型映射预设，格式与其他平台的快捷映射按钮保持一致。 */
export const kiroPresetMappings = kiroDefaultMappings.map(({ from, to }) => ({
  label: from,
  from,
  to,
  color: 'bg-sky-100 text-sky-700 hover:bg-sky-200 dark:bg-sky-900/30 dark:text-sky-400'
}))

export function isKiroCompleteDefaultMapping(rawMapping?: Record<string, unknown>): boolean {
  const mapping = sanitizeKiroModelMapping(rawMapping)
  return kiroPresetMappings.every(({ from, to }) => mapping[from] === to)
}

export function getKiroDefaultModelMappings(): Array<{ from: string; to: string }> {
  return kiroPresetMappings.map(({ from, to }) => ({ from, to }))
}
