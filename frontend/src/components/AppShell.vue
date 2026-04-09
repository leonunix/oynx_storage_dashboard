<template>
  <div class="shell">
    <aside class="sidebar">
      <div>
        <div class="brand-mark">ONYX</div>
        <h1 class="brand-title">Storage Dashboard</h1>
        <p class="brand-copy">
          {{ $t('shell.brandCopy') }}
        </p>
      </div>

      <nav class="nav-stack">
        <RouterLink v-for="item in visibleItems" :key="item.to" :to="item.to" class="nav-link-card">
          <i :class="item.icon"></i>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="status-panel">
        <div class="tiny-label">{{ $t('shell.currentUser') }}</div>
        <div class="status-value">{{ user?.displayName || $t('shell.notLoggedIn') }}</div>
        <div class="tiny-label mt-3">{{ $t('shell.role') }}</div>
        <div class="status-value">{{ user?.role || '-' }}</div>
        <div class="d-flex align-items-center gap-2 mt-3">
          <button class="btn btn-sm btn-outline-light" @click="$emit('logout')">{{ $t('shell.logout') }}</button>
          <LanguageSwitcher />
        </div>
      </div>
    </aside>

    <main class="content">
      <header class="hero-bar">
        <div>
          <p class="eyebrow">{{ eyebrow }}</p>
          <h2 class="page-title">{{ $t(title) }}</h2>
        </div>
        <slot name="header-actions" />
      </header>

      <section class="page-panel">
        <slot />
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import LanguageSwitcher from './LanguageSwitcher.vue'

const { t } = useI18n()

const props = defineProps({
  title: { type: String, required: true },
  eyebrow: { type: String, default: 'Onyx Control Plane' },
  user: { type: Object, default: null },
})

defineEmits(['logout'])

const items = computed(() => [
  { to: '/overview', label: t('shell.nav.overview'), icon: 'bi bi-speedometer2', permission: 'overview:read' },
  { to: '/storage', label: t('shell.nav.storage'), icon: 'bi bi-diagram-3', permission: 'storage:read' },
  { to: '/config', label: t('shell.nav.config'), icon: 'bi bi-gear', permission: 'storage:write' },
  { to: '/volumes', label: t('shell.nav.volumes'), icon: 'bi bi-hdd-stack', permission: 'volumes:read' },
  { to: '/metrics', label: t('shell.nav.metrics'), icon: 'bi bi-activity', permission: 'metrics:read' },
  { to: '/audit', label: t('shell.nav.audit'), icon: 'bi bi-journal-text', permission: 'audit:read' },
  { to: '/users', label: t('shell.nav.users'), icon: 'bi bi-people', permission: 'users:manage' },
])

const visibleItems = computed(() => {
  const permissions = props.user?.permissions || []
  return items.value.filter((item) => permissions.includes(item.permission))
})
</script>
