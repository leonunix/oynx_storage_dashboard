<template>
  <AppShell title="audit.title" eyebrow="Security & Operations" :user="auth.user" @logout="logout">
    <div class="content-card">
      <div class="section-header">
        <h3>{{ $t('audit.recentEvents') }}</h3>
        <button class="btn btn-sm btn-outline-light" @click="load">{{ $t('common.refresh') }}</button>
      </div>
      <div class="timeline">
        <div v-for="item in items" :key="item.id" class="timeline-item">
          <div class="timeline-time">{{ item.at }}</div>
          <div>
            <strong>{{ item.actor }}</strong> · {{ item.action }} · {{ item.resource }}
            <div class="text-secondary">{{ item.result }} · {{ item.description }}</div>
          </div>
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
const items = ref([])

const load = async () => {
  const { data } = await http.get('/audit/events')
  items.value = data.items || []
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

onMounted(async () => {
  await auth.fetchMe()
  await load()
})
</script>
