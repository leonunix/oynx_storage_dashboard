<template>
  <AppShell title="metrics.title" eyebrow="Minute-level Telemetry" :user="auth.user" @logout="logout">
    <template #header-actions>
      <div class="d-flex gap-2 align-items-center flex-wrap justify-content-end">
        <div class="window-switch">
          <button
            v-for="item in telemetryWindows"
            :key="item.key"
            class="btn btn-sm"
            :class="selectedWindow === item.key ? 'btn-accent' : 'btn-outline-light'"
            @click="selectWindow(item.key)"
          >
            {{ item.label }}
          </button>
        </div>
        <span class="badge text-bg-dark">Auto refresh 20s</span>
        <button class="btn btn-accent" @click="load">{{ $t('common.refresh') }}</button>
      </div>
    </template>

    <FlowPipeline :snapshot="latest" :rates="telemetry.rates" :window-label="windowLabel" />

    <div class="lane-grid">
      <div v-for="lane in laneCards" :key="lane.key" class="content-card lane-card">
        <div class="lane-head">
          <div>
            <div class="tiny-label">{{ lane.kicker }}</div>
            <h3>{{ lane.title }}</h3>
          </div>
          <div class="lane-icon" :style="{ color: lane.color, background: lane.glow }">
            <i :class="lane.icon"></i>
          </div>
        </div>
        <p class="chart-note">{{ lane.note }}</p>
        <div class="lane-stats">
          <div>
            <span>Read</span>
            <strong>{{ lane.read }}</strong>
          </div>
          <div>
            <span>Write</span>
            <strong>{{ lane.write }}</strong>
          </div>
        </div>
        <TrendChart compact :height="82" :series="lane.series" />
      </div>
    </div>

    <MetadbStatusPanel :stats="metadbStats" />

    <div class="row g-4">
      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Write throughput</h3>
              <p class="chart-note">Logical ingest, buffer absorb and durable LV3 writes on one scale.</p>
            </div>
            <span class="badge text-bg-dark">{{ windowLabel }}</span>
          </div>
          <div class="chart-legend">
            <span v-for="serie in writeThroughputSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="writeThroughputSeries" :height="270" format="bytesRate" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Read throughput</h3>
              <p class="chart-note">Read demand, buffer hits and LV3 fallthrough now stay readable on the same unit.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in readThroughputSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="readThroughputSeries" :height="270" format="bytesRate" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Write IOPS</h3>
              <p class="chart-note">When throughput is steady but latency is not, this is usually the next chart to check.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in writeIopsSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="writeIopsSeries" :height="240" format="opsRate" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Read IOPS</h3>
              <p class="chart-note">Separating reads from writes keeps the scale from collapsing into noise.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in readIopsSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="readIopsSeries" :height="240" format="opsRate" />
        </div>
      </div>

      <div class="col-12 col-xl-4">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Buffer fill</h3>
              <p class="chart-note">Pure percentage chart, no queue depth mixed in.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in bufferFillSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="bufferFillSeries" :height="220" format="percent" />
        </div>
      </div>

      <div class="col-12 col-xl-4">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Pending entries</h3>
              <p class="chart-note">Queue depth rendered alone, so spikes are finally obvious.</p>
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

      <div class="col-12 col-xl-4">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Payload memory</h3>
              <p class="chart-note">Resident buffer bytes get their own scale now.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in payloadBytesSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="payloadBytesSeries" :height="220" format="bytes" />
        </div>
      </div>

      <div class="col-12 col-xl-6">
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Reduction ratios</h3>
              <p class="chart-note">Compression and total reduction share a ratio scale, so they can be compared directly.</p>
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
              <p class="chart-note">A single percent chart reads much better than mixing it into ratios.</p>
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
        <div class="content-card chart-card">
          <div class="section-header">
            <div>
              <h3>Background activity</h3>
              <p class="chart-note">Dedup misses, GC rewrites and backpressure explain why the engine feels busy.</p>
            </div>
          </div>
          <div class="chart-legend">
            <span v-for="serie in backgroundSeries" :key="serie.key">
              <i class="legend-dot" :style="{ background: serie.color }"></i>
              {{ serie.label }}
            </span>
          </div>
          <TrendChart :series="backgroundSeries" :height="220" />
        </div>
      </div>
    </div>

    <div class="metric-summary-grid">
      <div v-for="group in metricGroups" :key="group.title" class="content-card summary-card">
        <div class="section-header">
          <h3>{{ group.title }}</h3>
          <i :class="group.icon"></i>
        </div>
        <div class="summary-chip-grid">
          <div v-for="item in group.items" :key="item.label" class="summary-chip">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
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
import MetadbStatusPanel from '../components/MetadbStatusPanel.vue'
import TrendChart from '../components/TrendChart.vue'
import {
  buildSeries,
  formatBytes,
  formatBytesPerSec,
  formatNumber,
  formatOpsPerSec,
  formatPercent,
  formatRatio,
  formatWindowLabel,
  seriesForKey,
  telemetryWindows,
} from '../lib/telemetry'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const AUTO_REFRESH_MS = 20000

const telemetry = ref({ series: {}, rates: {}, latest: null })
const raw = ref({})
const overview = ref({})
const selectedWindow = ref('24h')

let refreshHandle = null

const latest = computed(() => telemetry.value.latest)
const windowLabel = computed(() => formatWindowLabel(selectedWindow.value))
const metadbStats = computed(() =>
  overview.value?.metadb || overview.value?.metrics?.metadb || overview.value?.metrics?.metadb_memory || null,
)

const writeThroughputSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'client_write_bps', label: 'Client write', color: '#2563eb' },
    { key: 'buffer_write_bps', label: 'Buffer write', color: '#0d9488' },
    { key: 'lv3_write_bps', label: 'LV3 write', color: '#0891b2' },
  ]),
)

const readThroughputSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'client_read_bps', label: 'Client read', color: '#475569' },
    { key: 'buffer_read_bps', label: 'Buffer read', color: '#3b82f6' },
    { key: 'lv3_read_bps', label: 'LV3 disk (compressed)', color: '#f59e0b' },
    { key: 'lv3_read_decompressed_bps', label: 'LV3 user (decompressed)', color: '#0d9488' },
  ]),
)

const writeIopsSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'client_write_iops', label: 'Client write IOPS', color: '#2563eb' },
    { key: 'buffer_write_iops', label: 'Buffer write IOPS', color: '#0d9488' },
    { key: 'lv3_write_iops', label: 'LV3 write IOPS', color: '#0891b2' },
  ]),
)

const readIopsSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'client_read_iops', label: 'Client read IOPS', color: '#475569' },
    { key: 'buffer_read_iops', label: 'Buffer read IOPS', color: '#3b82f6' },
    { key: 'lv3_read_iops', label: 'LV3 read IOPS', color: '#f59e0b' },
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

const payloadBytesSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'buffer_payload_bytes', label: 'Payload bytes', color: '#0d9488' },
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

const backgroundSeries = computed(() =>
  buildSeries(telemetry.value, [
    { key: 'dedup_hits_per_min', label: 'Dedup hits/min', color: '#0d9488' },
    { key: 'dedup_misses_per_min', label: 'Dedup misses/min', color: '#3b82f6' },
    { key: 'gc_rewrites_per_min', label: 'GC rewrites/min', color: '#f59e0b' },
    { key: 'backpressure_events_per_min', label: 'Backpressure/min', color: '#ef4444' },
  ]),
)

const laneCards = computed(() => [
  {
    key: 'client',
    kicker: 'Logical traffic',
    title: 'Client IO',
    note: 'What the host actually asked the engine to do.',
    icon: 'bi bi-arrow-down-up',
    color: '#2563eb',
    glow: 'linear-gradient(135deg, rgba(46, 111, 206, 0.18), rgba(46, 111, 206, 0.06))',
    read: `${formatBytesPerSec(telemetry.value.rates?.clientReadBps)} · ${formatOpsPerSec(telemetry.value.rates?.clientReadIops)}`,
    write: `${formatBytesPerSec(telemetry.value.rates?.clientWriteBps)} · ${formatOpsPerSec(telemetry.value.rates?.clientWriteIops)}`,
    series: [
      { key: 'client-read', color: '#475569', points: seriesForKey(telemetry.value, 'client_read_bps') },
      { key: 'client-write', color: '#2563eb', points: seriesForKey(telemetry.value, 'client_write_bps') },
    ],
  },
  {
    key: 'buffer',
    kicker: 'Hot path',
    title: 'Buffer IO',
    note: 'Append throughput and read hits served before LV3 gets involved.',
    icon: 'bi bi-layers',
    color: '#0d9488',
    glow: 'linear-gradient(135deg, rgba(30, 167, 161, 0.18), rgba(30, 167, 161, 0.06))',
    read: `${formatBytesPerSec(telemetry.value.rates?.bufferReadBps)} · ${formatOpsPerSec(telemetry.value.rates?.bufferReadIops)}`,
    write: `${formatBytesPerSec(telemetry.value.rates?.bufferWriteBps)} · ${formatOpsPerSec(telemetry.value.rates?.bufferWriteIops)}`,
    series: [
      { key: 'buffer-read', color: '#3b82f6', points: seriesForKey(telemetry.value, 'buffer_read_bps') },
      { key: 'buffer-write', color: '#0d9488', points: seriesForKey(telemetry.value, 'buffer_write_bps') },
    ],
  },
  {
    key: 'lv3',
    kicker: 'Durable IO',
    title: 'LV3',
    note: 'Physical reads and writes after packing, dedup and background work.',
    icon: 'bi bi-hdd-network',
    color: '#0891b2',
    glow: 'linear-gradient(135deg, rgba(20, 126, 141, 0.18), rgba(20, 126, 141, 0.06))',
    read: `${formatBytesPerSec(telemetry.value.rates?.lv3ReadBps)} · ${formatOpsPerSec(telemetry.value.rates?.lv3ReadIops)}`,
    write: `${formatBytesPerSec(telemetry.value.rates?.lv3WriteBps)} · ${formatOpsPerSec(telemetry.value.rates?.lv3WriteIops)}`,
    series: [
      { key: 'lv3-read', color: '#f59e0b', points: seriesForKey(telemetry.value, 'lv3_read_bps') },
      { key: 'lv3-write', color: '#0891b2', points: seriesForKey(telemetry.value, 'lv3_write_bps') },
    ],
  },
])

const metricGroups = computed(() => {
  const metrics = raw.value || {}
  return [
    {
      title: 'Volume edge',
      icon: 'bi bi-hdd-stack',
      items: [
        { label: 'Uptime', value: `${formatNumber(metrics.uptime_secs || 0)}s` },
        { label: 'Volumes', value: latest.value?.volumeCount ?? 0 },
        { label: 'Volume reads', value: formatNumber(metrics.volume_read_ops || 0) },
        { label: 'Volume writes', value: formatNumber(metrics.volume_write_ops || 0) },
      ],
    },
    {
      title: 'Buffer layer',
      icon: 'bi bi-layers',
      items: [
        { label: 'Appends', value: formatNumber(metrics.buffer_appends || 0) },
        { label: 'Pending', value: formatNumber(latest.value?.bufferPendingEntries || 0) },
        { label: 'Payload', value: formatBytes(latest.value?.bufferPayloadBytes || 0) },
        { label: 'Backpressure', value: formatNumber(metrics.buffer_backpressure_events || 0) },
      ],
    },
    {
      title: 'Optimization',
      icon: 'bi bi-stars',
      items: [
        { label: 'Compression', value: formatRatio(latest.value?.compressionRatio) },
        { label: 'Dedup rate', value: formatPercent(latest.value?.dedupHitRatePct || 0) },
        { label: 'Dedup hits', value: formatNumber(metrics.dedup_hits || 0) },
        { label: 'Dedup misses', value: formatNumber(metrics.dedup_misses || 0) },
      ],
    },
    {
      title: 'Background work',
      icon: 'bi bi-gear-wide-connected',
      items: [
        { label: 'GC rewrites', value: formatNumber(metrics.gc_blocks_rewritten || 0) },
        { label: 'GC cycles', value: formatNumber(metrics.gc_cycles || 0) },
        { label: 'Flush errors', value: formatNumber(metrics.flush_errors || 0) },
        { label: 'CRC errors', value: formatNumber(metrics.read_crc_errors || 0) },
      ],
    },
  ]
})

const load = async () => {
  const [telemetryResp, metricsResp, overviewResp] = await Promise.all([
    http.get(`/metrics/timeseries?window=${selectedWindow.value}`),
    http.get('/metrics/summary'),
    http.get('/dashboard/overview'),
  ])

  telemetry.value = telemetryResp.data || { series: {}, rates: {}, latest: null }
  raw.value = metricsResp.data || {}
  overview.value = overviewResp.data || {}
}

const selectWindow = async (nextWindow) => {
  if (selectedWindow.value === nextWindow) return
  selectedWindow.value = nextWindow
  await load()
}

const logout = () => {
  auth.logout()
  router.push('/login')
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
.window-switch {
  display: inline-flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}

.lane-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.lane-card,
.chart-card {
  display: grid;
  gap: 0.75rem;
}

.lane-head,
.lane-stats,
.chart-legend {
  display: flex;
}

.lane-head {
  justify-content: space-between;
  gap: 1rem;
}

.lane-head h3 {
  margin: 0.125rem 0 0;
  font-size: 1rem;
  font-weight: 600;
}

.lane-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: grid;
  place-items: center;
  border-radius: var(--onyx-radius-xs);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface-soft);
  font-size: 1rem;
}

.lane-stats {
  gap: 0.5rem;
}

.lane-stats > div {
  flex: 1 1 0;
  display: grid;
  gap: 0.125rem;
  padding: 0.625rem 0.75rem;
  border-radius: var(--onyx-radius-sm);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface-soft);
  font-size: 0.8125rem;
}

.lane-stats span,
.chart-note {
  color: var(--onyx-muted);
}

.chart-note {
  margin: 0.125rem 0 0;
  font-size: 0.8125rem;
}

.chart-legend {
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

.metric-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.summary-card i {
  color: var(--onyx-primary);
  font-size: 1rem;
}

.summary-chip-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.summary-chip {
  display: grid;
  gap: 0.125rem;
  padding: 0.625rem 0.75rem;
  border-radius: var(--onyx-radius-sm);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface-soft);
}

.summary-chip span {
  color: var(--onyx-muted);
  font-size: 0.75rem;
}

.summary-chip strong {
  font-size: 0.9375rem;
}

@media (max-width: 1100px) {
  .lane-grid,
  .metric-summary-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .summary-chip-grid {
    grid-template-columns: 1fr;
  }

  .lane-stats {
    flex-direction: column;
  }
}
</style>
