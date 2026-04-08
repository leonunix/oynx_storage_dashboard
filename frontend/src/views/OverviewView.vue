<template>
  <AppShell title="引擎总览" eyebrow="Service Health" :user="auth.user" @logout="logout">
    <template #header-actions>
      <button class="btn btn-accent" @click="load">刷新状态</button>
    </template>

    <div class="stat-grid">
      <StatCard label="Engine Mode" :value="overview.engineMode || '-'" note="full / meta-only / unknown" />
      <StatCard label="Volumes" :value="overview.volumeCount ?? '-'" note="当前配置卷数量" />
      <StatCard label="Zone Workers" :value="overview.zoneCount ?? '-'" note="并发工作单元" />
      <StatCard label="Buffer Fill" :value="`${overview.bufferFillPercent ?? 0}%`" note="接近高水位时应重点关注" />
    </div>

    <div class="row g-4">
      <div class="col-12 col-xl-7">
        <div class="content-card">
          <div class="section-header">
            <h3>关键指标</h3>
            <span class="badge text-bg-dark">实时快照</span>
          </div>
          <div class="metric-list">
            <div v-for="(value, key) in overview.metrics" :key="key" class="metric-row">
              <span>{{ key }}</span>
              <code>{{ formatMetric(value) }}</code>
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-xl-5">
        <div class="content-card h-100">
          <div class="section-header">
            <h3>原始状态输出</h3>
            <span class="badge text-bg-warning">CLI / IPC bridge</span>
          </div>
          <pre class="raw-block">{{ overview.rawStatus || '暂无数据' }}</pre>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup>
import { onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import StatCard from '../components/StatCard.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const overview = reactive({})

const load = async () => {
  const { data } = await http.get('/dashboard/overview')
  Object.assign(overview, data)
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

const formatMetric = (value) => {
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return value
}

onMounted(async () => {
  await auth.fetchMe()
  await load()
})
</script>
