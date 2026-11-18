<template>
  <div class="border-t pt-4" data-testid="kiro-cache-config-fields">
    <div class="flex items-center gap-2">
      <KiroIcon size="md" class="text-sky-600 dark:text-sky-400" />
      <span class="text-sm font-medium text-gray-900 dark:text-white">
        {{ t('admin.groups.kiroCacheEmulation.title') }}
      </span>
    </div>

    <label class="mt-3 flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
      <input
        :checked="enabled"
        type="checkbox"
        class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        @change="updateEnabled"
      />
      <span>{{ t('admin.groups.kiroCacheEmulation.enable') }}</span>
    </label>
    <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
      {{
        enabled
          ? t('admin.groups.kiroCacheEmulation.enabledHint')
          : t('admin.groups.kiroCacheEmulation.disabledHint')
      }}
    </p>

    <div v-if="enabled" class="mt-3 max-w-xs">
      <label class="input-label">
        {{ t('admin.groups.kiroCacheEmulation.ratio') }}
      </label>
      <div class="relative">
        <input
          :value="ratioPercent"
          type="number"
          step="0.1"
          min="0.1"
          max="100"
          class="input pr-9"
          @input="updateRatio"
        />
        <span
          class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm text-gray-500 dark:text-gray-400"
        >%</span>
      </div>
      <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
        {{ t('admin.groups.kiroCacheEmulation.ratioHint') }}
      </p>
    </div>

    <p class="mt-3 border-l-2 border-amber-400 pl-3 text-xs leading-5 text-amber-800 dark:text-amber-300">
      {{ t('admin.groups.kiroCacheEmulation.localOnlyWarning') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import KiroIcon from './KiroIcon.vue'

defineProps<{
  enabled: boolean
  ratioPercent: number | string | null
}>()

const emit = defineEmits<{
  'update:enabled': [value: boolean]
  'update:ratioPercent': [value: number | string]
}>()

const { t } = useI18n()

const updateEnabled = (event: Event) => {
  emit('update:enabled', (event.target as HTMLInputElement).checked)
}

const updateRatio = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  emit('update:ratioPercent', value === '' ? '' : Number(value))
}
</script>