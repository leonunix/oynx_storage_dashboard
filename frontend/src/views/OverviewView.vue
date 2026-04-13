<template>
  <AppShell title="overview.title" eyebrow="Flow-first Operations" :user="auth.user" @logout="logout">
    <template #header-actions>
      <div class="d-flex align-items-center gap-2 flex-wrap justify-content-end">
        <span class="badge text-bg-dark">Auto refresh 15s</span>
        <span class="badge text-bg-light border">Updated {{ lastRefreshLabel }}</span>
        <button class="btn btn-accent" @click="load">{{ $t('overview.refreshStatus') }}</button>
      </div>
    </template>

    <FlowPipeline :snapshot="latest" :rates="telemetry.rates" :window-label="historyWindowLabel" />

    <div class="overview-kpis">
      <KpiTrendCard
        v-for="card in kpiCards"
        :key="card.label"
        :icon="card.icon"
        :label="card.label"
        :value="card.value"
        :note="card.note"
        :series="card.series"
        :color="card.color"
        :baseline="card.baseline || 'zero'"
      />
    </div>

    <div class="row g-4">
      <div class="col-12 col-xl-8">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Write path throughput</h3>
              <p class="chart-note">Follow logical ingress, cache absorption and durable LV3 output in one view.</p>
            </div>
            <span class="badge text-bg-dark">{{ historyWindowLabel }}</span>
          </div>
          <div class="chart-legend">
            <span v-for="serie in writeFlowSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="writeFlowSeries" :height="250" format="bytesRate" />
        </div>
      </div>

      <div class="col-12 col-xl-4">
        <div class="content-card chart-card h-100">
          <div class="section-header">
            <div>
              <h3>Buffer + allocator posture</h3>
              <p class="chart-note">Pressure indicators that tell us how much headroom the hot path still has.</p>
            </div>
          </div>
          <div class="metric-stack">
            <div class="metric-strip">
              <span>Engine mode</span>
              <strong>{{ latest?.engineMode || overview.engineMode || '-' }}</strong>
            </div>
            <div class="metric-strip">
              <span>Buffer memory</span>
              <strong>{{ formatBytes(latest?.bufferPayloadBytes) }} / {{ formatBytes(latest?.bufferPayloadLimit) }}</strong>
            </div>
            <div class="metric-strip">
              <span>Allocator usage</span>
              <strong>{{ formatPercent(latest?.allocatorUsagePercent) }}</strong>
            </div>
            <div class="metric-strip">
              <span>ublk devices</span>
              <strong>{{ ublkDevices.length ? ublkDevices.map((id) => `/dev/ublkb${id}`).join(', ') : 'none' }}</strong>
            </div>
          </div>
          <div class="chart-legend mt-3">
            <span v-for="serie in bufferFillSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="bufferFillSeries" :height="215" format="percent" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Read path and efficiency</h3>
              <p class="chart-note">Read latency starts with hit rate, but the whole path stays visible here.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in readFlowSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="readFlowSeries" :height="220" format="bytesRate" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Flush backlog</h3>
              <p class="chart-note">Queue depth is easier to reason about when it is not mixed with percentages or bytes.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in pendingEntriesSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="pendingEntriesSeries" :height="220" format="number" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Reduction ratios</h3>
              <p class="chart-note">Compression and total data reduction now share a chart because they use the same unit.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in ratioSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="ratioSeries" :height="220" baseline="fit" format="ratio" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Dedup hit rate</h3>
              <p class="chart-note">This now stands on its own scale, so the percentage curve stays readable.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in dedupRateSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="dedupRateSeries" :height="220" baseline="fit" format="percent" />
        </div>
      </div>

      <div class="col-12">
        <div class="content-card">
          <div class="section-header">
            <div>
              <h3>Buffer lanes</h3>
              <p class="chart-note">Shard-level visibility helps spot skew, stuck heads, or an unhealthy flush tail immediately.</p>
            </div>
            <span class="badge text-bg-dark">{{ bufferShards.length }} shards</span>
          </div>
          <div v-if="bufferShards.length" class="table-responsive">
            <table class="table align-middle table-sm">
              <thead>
                <tr>
                  <th>Lane</th>
                  <th>Fill</th>
                  <th>Used</th>
                  <th>Pending</th>
                  <th>Queue</th>
                  <th>Flushed stuck</th>
                  <th>Head age</th>
                  <th>Residency</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="shard in bufferShards" :key="shard.shard_idx">
                  <td><strong>#{{ shard.shard_idx }}</strong></td>
                  <td>
                    <div class="d-flex align-items-center gap-2">
                      <div class="fill-bar">
                        <div class="fill-bar-inner" :style="{ width: `${shard.fill_pct}%` }" :class="fillClass(shard.fill_pct)"></div>
                      </div>
                      <code>{{ shard.fill_pct }}%</code>
                    </div>
                  </td>
                  <td><code>{{ formatBytes(shard.used_bytes) }}</code></td>
                  <td><code>{{ shard.pending_entries }}</code></td>
                  <td><code>{{ shard.log_order_len ?? 0 }}</code></td>
                  <td><code :class="{ 'text-danger': shard.flushed_seqs_len > 100 }">{{ shard.flushed_seqs_len ?? 0 }}</code></td>
                  <td><code>{{ shard.head_age_ms != null ? `${Math.round(shard.head_age_ms / 1000)}s` : '-' }}</code></td>
                  <td><code>{{ shard.head_residency_ms != null ? `${Math.round(shard.head_residency_ms / 1000)}s` : '-' }}</code></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-note">No shard telemetry reported yet.</div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import FlowPipeline from '../components/FlowPipeline.vue'
import KpiTrendCard from '../components/KpiTrendCard.vue'
import TrendChart from '../components/TrendChart.vue'
import {
  buildSeries,
  formatBytes,
  formatBytesPerSec,
  formatDateTime,
  formatPercent,
  formatRatio,
  formatWindowLabel,
  seriesForKey,
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
const ublkDevices = computed(() => overview.value.ublkDevices || [])
const historyWindowLabel = computed(() => formatWindowLabel(HISTORY_WINDOW))
const lastRefreshLabel = computed(() => formatDateTime(lastLoadedAt.value))

const writeFlowSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'client_write_bps', label: 'Client ingress', color: '#2563eb' },
    { key: 'buffer_write_bps', label: 'Buffer absorb', color: '#0d9488' },
    { key: 'lv3_write_bps', label: 'LV3 durable write', color: '#0891b2' },
  ]),
)

const readFlowSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'client_read_bps', label: 'Client reads', color: '#475569' },
    { key: 'buffer_read_bps', label: 'Buffer hits', color: '#3b82f6' },
    { key: 'lv3_read_bps', label: 'LV3 reads', color: '#f59e0b' },
  ]),
)

const bufferFillSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'buffer_fill_pct', label: 'Buffer fill', color: '#ef4444' },
  ]),
)

const pendingEntriesSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'buffer_pending_entries', label: 'Pending entries', color: '#2563eb' },
  ]),
)

const ratioSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'compression_ratio', label: 'Compression', color: '#2563eb' },
    { key: 'data_reduction_ratio', label: 'Data reduction', color: '#f59e0b' },
  ]),
)

const dedupRateSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'dedup_hit_rate_pct', label: 'Dedup hit rate', color: '#0d9488' },
  ]),
)

const kpiCards = computed(() => [
  {
    icon: 'bi bi-arrow-right-circle',
    label: 'Client write',
    value: formatBytesPerSec(telemetry.value.rates?.clientWriteBps),
    note: 'Logical ingress over the minute-level history cadence.',
    series: [{ key: 'client-write', points: seriesForKey(telemetry.value, 'client_write_bps') }],
    color: '#2563eb',
  },
  {
    icon: 'bi bi-hourglass-split',
    label: 'Buffer fill',
    value: formatPercent(latest.value?.bufferFillPercent, 0),
    note: 'The first pressure signal to watch when traffic spikes.',
    series: [{ key: 'buffer-fill', points: seriesForKey(telemetry.value, 'buffer_fill_pct') }],
    color: '#ef4444',
  },
  {
    icon: 'bi bi-magic',
    label: 'Compression',
    value: formatRatio(latest.value?.compressionRatio),
    note: 'Compression ratio tracked from persisted samples, not browser-side deltas.',
    series: [{ key: 'compression', points: seriesForKey(telemetry.value, 'compression_ratio') }],
    color: '#0d9488',
    baseline: 'fit',
  },
  {
    icon: 'bi bi-intersect',
    label: 'Dedup hit rate',
    value: formatPercent(latest.value?.dedupHitRatePct),
    note: 'A clearer picture of how much identical content the pipeline is finding.',
    series: [{ key: 'dedup', points: seriesForKey(telemetry.value, 'dedup_hit_rate_pct') }],
    color: '#f59e0b',
    baseline: 'fit',
  },
  {
    icon: 'bi bi-bezier2',
    label: 'Data reduction',
    value: formatRatio(latest.value?.dataReductionRatio),
    note: 'Compression plus dedup presented as one end-to-end gain number.',
    series: [{ key: 'reduction', points: seriesForKey(telemetry.value, 'data_reduction_ratio') }],
    color: '#475569',
    baseline: 'fit',
  },
  {
    icon: 'bi bi-hdd-network',
    label: 'LV3 write',
    value: formatBytesPerSec(telemetry.value.rates?.lv3WriteBps),
    note: 'The durable side of the system, after packing and placement.',
    series: [{ key: 'lv3-write', points: seriesForKey(telemetry.value, 'lv3_write_bps') }],
    color: '#0891b2',
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
.overview-kpis {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.chart-card {
  display: grid;
  gap: 0.75rem;
}

.chart-note {
  margin: 0.125rem 0 0;
  color: var(--onyx-muted);
  font-size: 0.8125rem;
}

.chart-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  color: var(--onyx-muted);
  font-size: 0.8125rem;
}

.chart-legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
}

.legend-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  display: inline-block;
}

.metric-stack {
  display: grid;
  gap: 0.5rem;
}

.metric-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 0.75rem;
  border-radius: var(--onyx-radius-sm);
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
  font-size: 0.8125rem;
}

.metric-strip span {
  color: var(--onyx-muted);
}

.metric-strip strong {
  text-align: right;
  font-size: 0.8125rem;
}

.fill-bar {
  width: 72px;
  height: 6px;
  background: var(--onyx-border);
  border-radius: 3px;
  overflow: hidden;
}

.fill-bar-inner {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.fill-ok { background: var(--onyx-accent); }
.fill-warning { background: var(--onyx-warm); }
.fill-danger { background: var(--onyx-danger); }

.empty-note {
  color: var(--onyx-muted);
  padding: 0.75rem 0;
  font-size: 0.875rem;
}

@media (max-width: 1100px) {
  .overview-kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .overview-kpis {
    grid-template-columns: 1fr;
  }
}
</style>
