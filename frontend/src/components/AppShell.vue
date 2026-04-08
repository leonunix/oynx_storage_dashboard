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
        <RouterLink v-for="item in items" :key="item.to" :to="item.to" class="nav-link-card">
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
import { RouterLink } from 'vue-router'

defineProps({
  title: { type: String, required: true },
  eyebrow: { type: String, default: 'Onyx Control Plane' },
  user: { type: Object, default: null },
})

defineEmits(['logout'])

const items = [
  { to: '/overview', label: '总览', icon: 'bi bi-speedometer2' },
  { to: '/storage', label: '存储编排', icon: 'bi bi-diagram-3' },
  { to: '/volumes', label: 'Volumes', icon: 'bi bi-hdd-stack' },
  { to: '/metrics', label: 'Metrics', icon: 'bi bi-activity' },
  { to: '/audit', label: '审计', icon: 'bi bi-journal-text' },
]
</script>
