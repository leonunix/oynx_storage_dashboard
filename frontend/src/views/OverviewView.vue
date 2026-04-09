<template>
  <AppShell title="overview.title" eyebrow="Service Health" :user="auth.user" @logout="logout">
    <template #header-actions>
      <button class="btn btn-accent" @click="load">{{ $t('overview.refreshStatus') }}</button>
    </template>

    <div class="stat-grid">
      <StatCard label="Engine Mode" :value="overview.engineMode || '-'" :note="modeNote" />
      <StatCard label="Volumes" :value="overview.volumeCount ?? '-'" :note="$t('overview.volumeCountNote')" />
      <StatCard label="Zone Workers" :value="overview.zoneCount ?? '-'" :note="$t('overview.zoneWorkersNote')" />
      <StatCard label="Buffer Fill" :value="`${overview.bufferFillPercent ?? 0}%`" :note="$t('overview.bufferFillNote')" />
      <StatCard :label="$t('overview.compressionRatio')" :value="fmtRatio(overview.compressionRatio)" :note="$t('overview.compressionNote')" />
      <StatCard :label="$t('overview.dedupHitRate')" :value="fmtPercent(overview.dedupHitRate)" :note="$t('overview.dedupNote')" />
      <StatCard :label="$t('overview.dataReduction')" :value="fmtRatio(overview.dataReductionRatio)" :note="$t('overview.dataReductionNote')" />
    </div>

    <div class="row g-4">
      <!-- ublk devices -->
      <div v-if="ublkDevices.length" class="col-12">
        <div class="content-card">
          <div class="section-header">
            <h3>ublk Devices</h3>
            <span class="badge text-bg-success">{{ ublkDevices.length }} active</span>
          </div>
          <div class="chip-grid">
            <div v-for="devId in ublkDevices" :key="devId" class="chip-card">
              <strong>/dev/ublkb{{ devId }}</strong>
              <span>device id {{ devId }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Per-lane buffer stats -->
      <div v-if="bufferShards.length" class="col-12">
        <div class="content-card">
          <div class="section-header">
            <h3>Buffer Lanes (per-shard)</h3>
            <span class="badge text-bg-dark">{{ bufferShards.length }} shards</span>
          </div>
          <div class="table-responsive">
            <table class="table align-middle table-sm">
              <thead>
                <tr>
                  <th>Lane</th>
                  <th>Fill</th>
                  <th>Used</th>
                  <th>Capacity</th>
                  <th>Pending</th>
                  <th>Head</th>
                  <th>Tail</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="s in bufferShards" :key="s.shard_idx">
                  <td><strong>#{{ s.shard_idx }}</strong></td>
                  <td>
                    <div class="d-flex align-items-center gap-2">
                      <div class="fill-bar">
                        <div class="fill-bar-inner" :style="{ width: s.fill_pct + '%' }" :class="fillClass(s.fill_pct)"></div>
                      </div>
                      <code>{{ s.fill_pct }}%</code>
                    </div>
                  </td>
                  <td><code>{{ fmtSize(s.used_bytes) }}</code></td>
                  <td><code>{{ fmtSize(s.capacity_bytes) }}</code></td>
                  <td><code>{{ s.pending_entries }}</code></td>
                  <td><code>{{ s.head_offset }}</code></td>
                  <td><code>{{ s.tail_offset }}</code></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="col-12 col-xl-7">
        <div class="content-card">
          <div class="section-header">
            <h3>{{ $t('overview.keyMetrics') }}</h3>
            <span class="badge text-bg-dark">{{ $t('overview.liveSnapshot') }}</span>
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
            <h3>Allocator (LV3)</h3>
          </div>
          <div class="metric-list">
            <div class="metric-row">
              <span>free_blocks</span>
              <code>{{ overview.allocatorFreeBlocks ?? '-' }}</code>
            </div>
            <div class="metric-row">
              <span>total_blocks</span>
              <code>{{ overview.allocatorTotalBlocks ?? '-' }}</code>
            </div>
            <div class="metric-row">
              <span>usage</span>
              <code>{{ allocatorUsage }}</code>
            </div>
            <div class="metric-row">
              <span>buffer_pending</span>
              <code>{{ overview.bufferPendingEntries ?? '-' }}</code>
            </div>
            <div class="metric-row">
              <span>live_handles</span>
              <code>{{ overview.liveHandleCount ?? '-' }}</code>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import StatCard from '../components/StatCard.vue'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const overview = reactive({})
const ublkDevices = ref([])
const bufferShards = ref([])

const fmtRatio = (v) => {
  if (!v || v <= 0) return '1.00x'
  return v.toFixed(2) + 'x'
}

const fmtPercent = (v) => {
  if (!v && v !== 0) return '0%'
  return (v * 100).toFixed(1) + '%'
}

const modeNote = computed(() => {
  switch (overview.engineMode) {
    case 'active': return t('overview.modeNoteActive')
    case 'standby': return t('overview.modeNoteStandby')
    case 'bare': return t('overview.modeNoteBare')
    default: return ''
  }
})

const allocatorUsage = computed(() => {
  const free = overview.allocatorFreeBlocks
  const total = overview.allocatorTotalBlocks
  if (!total) return '-'
  const used = total - free
  const pct = ((used / total) * 100).toFixed(1)
  return `${pct}% (${used}/${total})`
})

const fmtSize = (bytes) => {
  if (!bytes) return '0'
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GiB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' MiB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(0) + ' KiB'
  return bytes + ' B'
}

const fillClass = (pct) => {
  if (pct >= 90) return 'fill-danger'
  if (pct >= 70) return 'fill-warning'
  return 'fill-ok'
}

const load = async () => {
  const { data } = await http.get('/dashboard/overview')
  Object.assign(overview, data)
  ublkDevices.value = data.ublkDevices || []
  bufferShards.value = data.bufferShards || []
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

<style scoped>
.fill-bar {
  width: 80px;
  height: 8px;
  background: rgba(0, 0, 0, 0.06);
  border-radius: 4px;
  overflow: hidden;
}

.fill-bar-inner {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.fill-ok { background: var(--onyx-accent); }
.fill-warning { background: var(--onyx-warm); }
.fill-danger { background: #dc3545; }
</style>
