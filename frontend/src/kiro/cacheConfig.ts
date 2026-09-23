export type KiroCacheConfigFormState = {
  platform: string
  kiro_cache_emulation_enabled: boolean
  kiro_cache_emulation_ratio_percent: number | string | null
}

export const kiroCacheRatioToPercent = (
  value: number | null | undefined
): number => {
  const ratio = Number(value)
  if (!Number.isFinite(ratio) || ratio <= 0) return 100
  return Math.round(Math.min(ratio, 1) * 1000) / 10
}

export const kiroCachePercentToRatio = (
  value: number | string | null | undefined
): number => {
  const percent = Number(value)
  if (!Number.isFinite(percent) || percent <= 0) return 0
  return Math.round(Math.min(percent, 100) * 10) / 1000
}

export const validateKiroCacheConfig = (
  form: KiroCacheConfigFormState
): string | null => {
  if (form.platform !== 'kiro' || !form.kiro_cache_emulation_enabled) {
    return null
  }
  const percent = Number(form.kiro_cache_emulation_ratio_percent)
  if (!Number.isFinite(percent) || percent <= 0 || percent > 100) {
    return 'ratioRangeError'
  }
  const ratio = kiroCachePercentToRatio(percent)
  return ratio > 0 && ratio <= 1 ? null : 'ratioRangeError'
}