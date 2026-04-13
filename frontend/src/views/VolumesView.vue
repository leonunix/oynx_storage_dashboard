<template>
  <AppShell title="volumes.title" eyebrow="Volume Operations" :user="auth.user" @logout="logout">
    <div class="row g-4">
      <div class="col-12 col-lg-5">
        <div class="content-card">
          <div class="section-header">
            <h3>{{ $t('volumes.createVolume') }}</h3>
            <span class="badge text-bg-success">{{ $t('common.writeOp') }}</span>
          </div>
          <form class="d-grid gap-3" @submit.prevent="createVolume">
            <input v-model="form.name" class="form-control" placeholder="volume name" />
            <input v-model.number="form.sizeBytes" type="number" class="form-control" placeholder="size bytes" />
            <select v-model="form.compression" class="form-select">
              <option value="lz4">lz4</option>
              <option value="zstd">zstd</option>
              <option value="none">none</option>
            </select>
            <button class="btn btn-accent">{{ $t('common.create') }}</button>
          </form>
        </div>
      </div>

      <div class="col-12 col-lg-7">
        <div class="content-card">
          <div class="section-header">
            <h3>{{ $t('volumes.existingVolumes') }}</h3>
            <button class="btn btn-sm btn-outline-light" @click="load">{{ $t('common.refresh') }}</button>
          </div>
          <div class="table-responsive">
            <table class="table align-middle">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Size</th>
                  <th>Zones</th>
                  <th>Compression</th>
                  <th>Status</th>
                  <th>IO</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in items" :key="item.name">
                  <td>
                    <strong>{{ item.name }}</strong>
                  </td>
                  <td>{{ fmtSize(item.sizeBytes) }}</td>
                  <td>{{ item.zoneCount }}</td>
                  <td><code>{{ item.compression }}</code></td>
                  <td>
                    <span :class="statusBadge(item.status)">{{ item.status }}</span>
                  </td>
                  <td class="io-cell">
                    <template v-if="item.metrics">
                      <span class="io-tag">R {{ item.metrics.readOps }} ops</span>
                      <span class="io-tag">W {{ item.metrics.writeOps }} ops</span>
                      <span class="io-tag">{{ fmtSize(item.metrics.readBytes) }} read</span>
                      <span class="io-tag">{{ fmtSize(item.metrics.writeBytes) }} write</span>
                      <span v-if="item.metrics.readErrors" class="io-tag io-err">{{ item.metrics.readErrors }} R err</span>
                      <span v-if="item.metrics.writeErrors" class="io-tag io-err">{{ item.metrics.writeErrors }} W err</span>
                    </template>
                    <span v-else class="text-muted">-</span>
                  </td>
                  <td class="text-end">
                    <button class="btn btn-sm btn-outline-danger" @click="removeVolume(item.name)">{{ $t('common.delete') }}</button>
                  </td>
                </tr>
                <tr v-if="items.length === 0">
                  <td colspan="7" class="text-center text-secondary py-4">{{ $t('volumes.noVolumes') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const items = ref([])
const form = reactive({
  name: '',
  sizeBytes: 1073741824,
  compression: 'lz4',
})

const fmtSize = (bytes) => {
  if (!bytes) return '0'
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GiB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MiB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(0) + ' KiB'
  return bytes + ' B'
}

const statusBadge = (s) => {
  if (s === 'online') return 'badge text-bg-success'
  if (s === 'configured') return 'badge text-bg-warning'
  return 'badge text-bg-secondary'
}

const load = async () => {
  const { data } = await http.get('/volumes')
  items.value = data.items || []
}

const createVolume = async () => {
  await http.post('/volumes', form)
  form.name = ''
  await load()
}

const removeVolume = async (name) => {
  await http.delete(`/volumes/${name}`)
  await load()
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

<style scoped>
.io-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.io-tag {
  display: inline-block;
  font-size: 0.6875rem;
  padding: 0.125rem 0.375rem;
  border-radius: var(--onyx-radius-xs);
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
  color: var(--onyx-text-secondary);
  white-space: nowrap;
}

.io-err {
  background: #fef2f2;
  border-color: #fecaca;
  color: var(--onyx-danger);
}
</style>
