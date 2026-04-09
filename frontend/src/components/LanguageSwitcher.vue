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
  width: 36px;
  height: 28px;
  border: 1px solid rgba(46, 111, 206, 0.18);
  border-radius: 6px;
  background: #ffffff;
  color: var(--onyx-primary, #2e6fce);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  letter-spacing: 0.02em;
}

.lang-switcher:hover {
  background: var(--onyx-primary-soft, #dceafe);
  border-color: rgba(46, 111, 206, 0.32);
}
</style>
