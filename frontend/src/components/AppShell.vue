<template>
  <div class="shell">
    <aside class="sidebar">
      <div>
        <div class="brand-mark">ONYX</div>
        <h1 class="brand-title">Storage Dashboard</h1>
        <p class="brand-copy">
          从设备编排到 volume 生命周期，再到引擎观测与审计，全都收在一个控制面里。
        </p>
      </div>

      <nav class="nav-stack">
        <RouterLink v-for="item in visibleItems" :key="item.to" :to="item.to" class="nav-link-card">
          <i :class="item.icon"></i>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="status-panel">
        <div class="tiny-label">当前用户</div>
        <div class="status-value">{{ user?.displayName || '未登录' }}</div>
        <div class="tiny-label mt-3">角色</div>
        <div class="status-value">{{ user?.role || '-' }}</div>
        <button class="btn btn-sm btn-outline-light mt-3" @click="$emit('logout')">退出登录</button>
      </div>
    </aside>

    <main class="content">
      <header class="hero-bar">
        <div>
          <p class="eyebrow">{{ eyebrow }}</p>
          <h2 class="page-title">{{ title }}</h2>
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
import { RouterLink } from 'vue-router'

const props = defineProps({
  title: { type: String, required: true },
  eyebrow: { type: String, default: 'Onyx Control Plane' },
  user: { type: Object, default: null },
})

defineEmits(['logout'])

const items = [
  { to: '/overview', label: '总览', icon: 'bi bi-speedometer2', permission: 'overview:read' },
  { to: '/storage', label: '存储编排', icon: 'bi bi-diagram-3', permission: 'storage:read' },
  { to: '/config', label: '引擎配置', icon: 'bi bi-gear', permission: 'storage:write' },
  { to: '/volumes', label: 'Volumes', icon: 'bi bi-hdd-stack', permission: 'volumes:read' },
  { to: '/metrics', label: 'Metrics', icon: 'bi bi-activity', permission: 'metrics:read' },
  { to: '/audit', label: '审计', icon: 'bi bi-journal-text', permission: 'audit:read' },
  { to: '/users', label: '用户与权限', icon: 'bi bi-people', permission: 'users:manage' },
]

const visibleItems = computed(() => {
  const permissions = props.user?.permissions || []
  return items.filter((item) => permissions.includes(item.permission))
})
</script>
