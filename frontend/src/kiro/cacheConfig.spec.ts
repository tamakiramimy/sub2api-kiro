import { describe, expect, it } from 'vitest'

import {
  kiroCachePercentToRatio,
  kiroCacheRatioToPercent,
  validateKiroCacheConfig,
  type KiroCacheConfigFormState
} from './cacheConfig'

const formState = (
  overrides: Partial<KiroCacheConfigFormState> = {}
): KiroCacheConfigFormState => ({
  platform: 'kiro',
  kiro_cache_emulation_enabled: true,
  kiro_cache_emulation_ratio_percent: 100,
  ...overrides
})

describe('Kiro cache config conversions', () => {
  it('converts between the UI percent and the backend ratio', () => {
    expect(kiroCachePercentToRatio(65.4)).toBe(0.654)
    expect(kiroCacheRatioToPercent(0.654)).toBe(65.4)
    expect(kiroCachePercentToRatio(100)).toBe(1)
  })

  it('uses 100 percent for absent legacy ratios', () => {
    expect(kiroCacheRatioToPercent(undefined)).toBe(100)
    expect(kiroCacheRatioToPercent(0)).toBe(100)
  })
})

describe('validateKiroCacheConfig', () => {
  it('accepts enabled Kiro ratios in the backend range', () => {
    expect(validateKiroCacheConfig(formState())).toBeNull()
    expect(
      validateKiroCacheConfig(
        formState({ kiro_cache_emulation_ratio_percent: 0.1 })
      )
    ).toBeNull()
  })

  it('rejects empty, zero, non-finite and over-100 ratios', () => {
    for (const ratio of ['', 0, -1, 100.1, 'invalid']) {
      expect(
        validateKiroCacheConfig(
          formState({ kiro_cache_emulation_ratio_percent: ratio })
        )
      ).toBe('ratioRangeError')
    }
  })

  it('skips disabled and non-Kiro configurations', () => {
    expect(
      validateKiroCacheConfig(
        formState({
          kiro_cache_emulation_enabled: false,
          kiro_cache_emulation_ratio_percent: 0
        })
      )
    ).toBeNull()
    expect(
      validateKiroCacheConfig(
        formState({
          platform: 'openai',
          kiro_cache_emulation_ratio_percent: 0
        })
      )
    ).toBeNull()
  })
})