import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'
import { createI18n } from 'vue-i18n'
import { describe, expect, it } from 'vitest'

import OAuthAuthorizationFlow from '../OAuthAuthorizationFlow.vue'

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en: {} },
  missingWarn: false,
  fallbackWarn: false,
})

function mountKiroFlow() {
  return mount(OAuthAuthorizationFlow, {
    props: {
      addMethod: 'oauth',
      platform: 'kiro',
      authUrl: 'https://example.test/authorize',
      sessionId: 'kiro-session',
    },
    global: {
      plugins: [createPinia(), i18n],
      stubs: { Icon: { template: '<span />' } },
    },
  })
}

describe('OAuthAuthorizationFlow Kiro callback parsing', () => {
  it('extracts code and state from a complete callback URL', async () => {
    const wrapper = mountKiroFlow()
    const callbackURL = 'http://127.0.0.1:9876/oauth/callback?code=kiro-auth-code&state=kiro-oauth-state'

    await wrapper.get('textarea').setValue(callbackURL)
    await nextTick()

    expect((wrapper.vm as unknown as { authCode: string }).authCode).toBe('kiro-auth-code')
    expect((wrapper.vm as unknown as { oauthState: string }).oauthState).toBe('kiro-oauth-state')
  })

  it('extracts code and state from a callback query string', async () => {
    const wrapper = mountKiroFlow()

    await wrapper.get('textarea').setValue('?code=kiro-query-code&state=kiro-query-state')
    await nextTick()

    expect((wrapper.vm as unknown as { authCode: string }).authCode).toBe('kiro-query-code')
    expect((wrapper.vm as unknown as { oauthState: string }).oauthState).toBe('kiro-query-state')
  })

  it('preserves a bare Kiro authorization code', async () => {
    const wrapper = mountKiroFlow()

    await wrapper.get('textarea').setValue('kiro-bare-code')
    await nextTick()

    expect((wrapper.vm as unknown as { authCode: string }).authCode).toBe('kiro-bare-code')
    expect((wrapper.vm as unknown as { oauthState: string }).oauthState).toBe('')
  })
})