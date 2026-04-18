<template>
  <AppShell title="overview.title" eyebrow="Flow-first Operations" :user="auth.user" @logout="logout">
    <template #header-actions>
      <div class="d-flex align-items-center gap-2 flex-wrap justify-content-end">
        <span class="badge text-bg-dark">Auto refresh 15s</span>
        <span class="badge text-bg-light border">Updated {{ lastRefreshLabel }}</span>
        <button class="btn btn-accent" @click="load">{{ $t('overview.refreshStatus') }}</button>
      </div>
    </template>

    <!-- Hero: animated data flow diagram -->
    <DataFlowDiagram :snapshot="latest" :rates="telemetry.rates" />

    <!-- Compact stat cards -->
    <div class="stat-row">
      <div v-for="card in statCards" :key="card.label" class="stat-pill">
        <div class="stat-pill-icon" :style="{ color: card.color }">
          <i :class="card.icon"></i>
        </div>
        <div>
          <div class="stat-pill-label">{{ card.label }}</div>
          <div class="stat-pill-value">{{ card.value }}</div>
        </div>
      </div>
    </div>

    <!-- User IO latency (avg, derived from volume_*_total_ns / volume_*_ops) -->
    <div class="content-card">
      <div class="section-header">
        <div>
          <h3>User IO latency</h3>
          <p class="chart-note">
            Avg ms/op · read: volume.read() 进入到 return · write: volume.write() 进入到 ack 返回。
          </p>
        </div>
        <div class="d-flex align-items-center gap-2">
          <span class="latency-chip latency-chip-read">
            R <strong>{{ formatLatency(telemetry.rates?.clientReadLatencyNs) }}</strong>
          </span>
          <span class="latency-chip latency-chip-write">
            W <strong>{{ formatLatency(telemetry.rates?.clientWriteLatencyNs) }}</strong>
          </span>
          <span class="badge text-bg-light border">{{ HISTORY_WINDOW }}</span>
        </div>
      </div>
      <TrendChart :series="latencySeries" :height="200" format="duration" />
    </div>

    <!-- Buffer lanes (visual) -->
    <div class="content-card">
      <div class="section-header">
        <div>
          <h3>Buffer lanes</h3>
          <p class="chart-note">Per-shard pressure, queue depth, and flush health at a glance.</p>
        </div>
        <span class="badge text-bg-dark">{{ bufferShards.length }} shards</span>
      </div>

      <div v-if="bufferShards.length" class="lane-grid">
        <div v-for="shard in bufferShards" :key="shard.shard_idx" class="lane-card" :class="laneClass(shard)">
          <!-- top: lane id + fill bar -->
          <div class="lane-top">
            <div class="lane-id">#{{ shard.shard_idx }}</div>
            <div class="lane-fill-track">
              <div class="lane-fill-bar" :class="fillClass(shard.fill_pct)" :style="{ width: `${shard.fill_pct}%` }">
                <span class="lane-fill-label" v-if="shard.fill_pct >= 20">{{ shard.fill_pct }}%</span>
              </div>
              <span class="lane-fill-label lane-fill-label-out" v-if="shard.fill_pct < 20">{{ shard.fill_pct }}%</span>
            </div>
          </div>
          <!-- metrics row -->
          <div class="lane-metrics">
            <div class="lane-metric">
              <span class="lane-metric-label">Used</span>
              <span class="lane-metric-value">{{ formatBytes(shard.used_bytes) }}</span>
            </div>
            <div class="lane-metric">
              <span class="lane-metric-label">Pending</span>
              <span class="lane-metric-value">{{ shard.pending_entries }}</span>
            </div>
            <div class="lane-metric">
              <span class="lane-metric-label">Queue</span>
              <span class="lane-metric-value">{{ shard.log_order_len ?? 0 }}</span>
            </div>
            <div class="lane-metric">
              <span class="lane-metric-label">Stuck</span>
              <span class="lane-metric-value" :class="{ 'text-danger': shard.flushed_seqs_len > 100 }">{{ shard.flushed_seqs_len ?? 0 }}</span>
            </div>
            <div class="lane-metric">
              <span class="lane-metric-label">Head age</span>
              <span class="lane-metric-value">{{ shard.head_age_ms != null ? `${Math.round(shard.head_age_ms / 1000)}s` : '-' }}</span>
            </div>
            <div class="lane-metric">
              <span class="lane-metric-label">Residency</span>
              <span class="lane-metric-value">{{ shard.head_residency_ms != null ? `${Math.round(shard.head_residency_ms / 1000)}s` : '-' }}</span>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-note">No shard telemetry reported yet.</div>
    </div>
  </AppShell>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import DataFlowDiagram from '../components/DataFlowDiagram.vue'
import TrendChart from '../components/TrendChart.vue'
import {
  formatBytes,
  formatDateTime,
  formatDurationNs,
  formatPercent,
  formatRatio,
} from '../lib/telemetry'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const AUTO_REFRESH_MS = 15000
const HISTORY_WINDOW = '6h'

const overview = ref({})
const telemetry = ref({ series: {}, rates: {}, latest: null })
const lastLoadedAt = ref(null)

let refreshHandle = null

const latest = computed(() => telemetry.value.latest)
const bufferShards = computed(() => overview.value.bufferShards || [])
const lastRefreshLabel = computed(() => formatDateTime(lastLoadedAt.value))

const latencySeries = computed(() => {
  const series = telemetry.value?.series || {}
  return [
    {
      key: 'client_read_latency',
      label: 'Read',
      color: '#0d9488',
      points: series.client_read_latency_ns || [],
    },
    {
      key: 'client_write_latency',
      label: 'Write',
      color: '#2563eb',
      points: series.client_write_latency_ns || [],
    },
  ]
})

const formatLatency = (value) => formatDurationNs(value)

const statCards = computed(() => [
  {
    icon: 'bi bi-bezier2',
    label: 'Data reduction',
    value: formatRatio(latest.value?.dataReductionRatio),
    color: '#0d9488',
  },
  {
    icon: 'bi bi-pie-chart',
    label: 'Allocator usage',
    value: formatPercent(latest.value?.allocatorUsagePercent),
    color: '#6366f1',
  },
  {
    icon: 'bi bi-hdd-stack',
    label: 'Volumes',
    value: `${latest.value?.volumeCount ?? 0}`,
    color: '#2563eb',
  },
  {
    icon: 'bi bi-gear-wide-connected',
    label: 'Engine mode',
    value: latest.value?.engineMode || '-',
    color: '#475569',
  },
])

const load = async () => {
  const [overviewResp, telemetryResp] = await Promise.all([
    http.get('/dashboard/overview'),
    http.get(`/dashboard/telemetry?window=${HISTORY_WINDOW}`),
  ])

  overview.value = overviewResp.data || {}
  telemetry.value = telemetryResp.data || { series: {}, rates: {}, latest: null }
  lastLoadedAt.value = new Date().toISOString()
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

const fillClass = (pct) => {
  if (pct >= 90) return 'fill-danger'
  if (pct >= 70) return 'fill-warning'
  return 'fill-ok'
}

const laneClass = (shard) => {
  if (shard.fill_pct >= 90 || shard.flushed_seqs_len > 100) return 'lane-alert'
  if (shard.fill_pct >= 70) return 'lane-warm'
  return ''
}

onMounted(async () => {
  await auth.fetchMe()
  await load()
  refreshHandle = window.setInterval(load, AUTO_REFRESH_MS)
})

onBeforeUnmount(() => {
  if (refreshHandle) {
    window.clearInterval(refreshHandle)
  }
})
</script>

<style scoped>
/* ─── Stat pills ────────────────────────────────── */

.stat-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.stat-pill {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  border-radius: var(--onyx-radius);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
  box-shadow: var(--onyx-shadow-sm);
}

.stat-pill-icon {
  display: grid;
  place-items: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--onyx-radius-xs);
  font-size: 1rem;
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
  flex-shrink: 0;
}

.stat-pill-label {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--onyx-muted);
}

.stat-pill-value {
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1.2;
  margin-top: 0.0625rem;
}

/* ─── Section helpers ───────────────────────────── */

.chart-note {
  margin: 0.125rem 0 0;
  color: var(--onyx-muted);
  font-size: 0.8125rem;
}

.empty-note {
  color: var(--onyx-muted);
  padding: 0.75rem 0;
  font-size: 0.875rem;
}

/* ─── Buffer lanes (visual cards) ───────────────── */

.lane-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 0.625rem;
}

.lane-card {
  display: grid;
  gap: 0.625rem;
  padding: 0.75rem 0.875rem;
  border-radius: var(--onyx-radius-sm);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
  transition: border-color 0.3s, box-shadow 0.3s;
}

.lane-warm {
  border-color: #fbbf2440;
  box-shadow: inset 0 0 0 1px #fbbf2418;
}

.lane-alert {
  border-color: #ef444440;
  box-shadow: inset 0 0 0 1px #ef444418;
  animation: pulse-alert 2s ease-in-out infinite;
}

@keyframes pulse-alert {
  0%, 100% { box-shadow: inset 0 0 0 1px #ef444418; }
  50% { box-shadow: inset 0 0 0 1px #ef444430, 0 0 8px #ef444412; }
}

/* ─── Lane top (id + fill bar) ──────────────────── */

.lane-top {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.lane-id {
  font-size: 0.75rem;
  font-weight: 800;
  color: var(--onyx-muted);
  min-width: 1.75rem;
}

.lane-fill-track {
  flex: 1;
  height: 18px;
  border-radius: 4px;
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
  overflow: hidden;
  position: relative;
  display: flex;
  align-items: center;
}

.lane-fill-bar {
  height: 100%;
  border-radius: 3px 0 0 3px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-right: 0.375rem;
  transition: width 0.5s ease;
  position: relative;
}

.lane-fill-bar.fill-ok {
  background: linear-gradient(90deg, #0d948822, var(--onyx-accent));
}

.lane-fill-bar.fill-warning {
  background: linear-gradient(90deg, #f59e0b22, var(--onyx-warm));
}

.lane-fill-bar.fill-danger {
  background: linear-gradient(90deg, #ef444422, var(--onyx-danger));
}

.lane-fill-label {
  font-size: 0.5625rem;
  font-weight: 800;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0,0,0,0.3);
  white-space: nowrap;
}

.lane-fill-label-out {
  color: var(--onyx-muted);
  text-shadow: none;
  margin-left: 0.375rem;
  position: relative;
}

/* ─── Lane metrics row ──────────────────────────── */

.lane-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.375rem;
}

.lane-metric {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0.25rem 0.375rem;
  border-radius: 3px;
  background: var(--onyx-surface-soft);
}

.lane-metric-label {
  font-size: 0.5625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--onyx-muted);
}

.lane-metric-value {
  font-size: 0.8125rem;
  font-weight: 700;
  line-height: 1.3;
}

.text-danger {
  color: var(--onyx-danger) !important;
}

/* ─── Latency chips ─────────────────────────────── */

.latency-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.18rem 0.55rem;
  border-radius: 999px;
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface-soft);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--onyx-muted);
}

.latency-chip strong {
  font-weight: 700;
  color: var(--onyx-text, #0f172a);
  letter-spacing: normal;
  text-transform: none;
}

.latency-chip-read {
  border-color: #0d948840;
  background: #0d948810;
}

.latency-chip-read strong {
  color: #0d9488;
}

.latency-chip-write {
  border-color: #2563eb40;
  background: #2563eb10;
}

.latency-chip-write strong {
  color: #2563eb;
}

/* ─── Responsive ────────────────────────────────── */

@media (max-width: 900px) {
  .stat-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .stat-row {
    grid-template-columns: 1fr;
  }

  .lane-grid {
    grid-template-columns: 1fr;
  }

  .lane-metrics {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
