import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN.js'
import en from './locales/en.js'

const savedLocale = localStorage.getItem('onyx_dashboard_locale') || 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'zh-CN',
  messages: { 'zh-CN': zhCN, en },
})

export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem('onyx_dashboard_locale', locale)
}

export default i18n
