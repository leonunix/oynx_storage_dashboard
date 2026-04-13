<template>
  <button class="lang-switcher" @click="toggle" :title="locale === 'zh-CN' ? 'Switch to English' : '切换到中文'">
    {{ locale === 'zh-CN' ? 'EN' : '中' }}
  </button>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { setLocale } from '../i18n'

const { locale } = useI18n()

function toggle() {
  setLocale(locale.value === 'zh-CN' ? 'en' : 'zh-CN')
}
</script>

<style scoped>
.lang-switcher {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 28px;
  border: 1px solid var(--onyx-border);
  border-radius: var(--onyx-radius-xs);
  background: var(--onyx-surface);
  color: var(--onyx-text-secondary);
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.lang-switcher:hover {
  background: var(--onyx-surface-soft);
  border-color: #cbd5e1;
}
</style>
