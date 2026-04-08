<template>
  <AppShell title="Metrics" eyebrow="Pipeline Telemetry" :user="auth.user" @logout="logout">
    <div class="content-card">
      <div class="section-header">
        <h3>引擎快照</h3>
        <button class="btn btn-sm btn-outline-light" @click="load">刷新</button>
      </div>
      <div class="metric-list">
        <div v-for="(value, key) in metrics" :key="key" class="metric-row">
          <span>{{ key }}</span>
          <code>{{ format(value) }}</code>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const metrics = ref({})

const load = async () => {
  const { data } = await http.get('/metrics/summary')
  metrics.value = data.metrics || {}
}

const format = (value) => (typeof value === 'object' ? JSON.stringify(value) : value)

const logout = () => {
  auth.logout()
  router.push('/login')
}

onMounted(async () => {
  await auth.fetchMe()
  await load()
})
</script>
