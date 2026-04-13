<template>
  <AppShell title="metrics.title" eyebrow="Pipeline Telemetry" :user="auth.user" @logout="logout">
    <template #header-actions>
      <button class="btn btn-accent" @click="load">{{ $t('common.refresh') }}</button>
    </template>

    <div class="row g-4 mb-1">
      <div v-for="lane in ioRateCards" :key="lane.key" class="col-12 col-lg-4">
        <div class="content-card h-100">
          <div class="section-header">
            <h3>{{ lane.title }}</h3>
            <span class="badge text-bg-dark">{{ rateWindowLabel }}</span>
          </div>
          <p class="lane-note">{{ lane.note }}</p>
          <div class="rate-grid">
            <div class="rate-item">
              <span>read throughput</span>
              <strong>{{ lane.readThroughput }}</strong>
            </div>
            <div class="rate-item">
              <span>read IOPS</span>
              <strong>{{ lane.readIops }}</strong>
            </div>
            <div class="rate-item">
              <span>write throughput</span>
              <strong>{{ lane.writeThroughput }}</strong>
            </div>
            <div class="rate-item">
              <span>write IOPS</span>
              <strong>{{ lane.writeIops }}</strong>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="row g-4">
      <div v-for="group in metricGroups" :key="group.title" class="col-12 col-xl-6">
        <div class="content-card">
          <div class="section-header">
            <h3>{{ group.title }}</h3>
          </div>
          <div class="metric-list">
            <div v-for="item in group.items" :key="item.key" class="metric-row">
              <span>{{ item.label }}</span>
              <code>{{ item.value }}</code>
            </div>
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
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const raw = ref({})
const previousSample = ref(null)
const latestSample = ref(null)

const POLL_INTERVAL_MS = 2000
let pollHandle = null

const load = async () => {
  const [metricsResp, overviewResp] = await Promise.allSettled([
    http.get('/metrics/summary'),
    http.get('/dashboard/overview'),
  ])

  const metrics =
    metricsResp.status === 'fulfilled' ? (metricsResp.value.data || {}) : {}
  const overview =
    overviewResp.status === 'fulfilled' ? (overviewResp.value.data || {}) : {}

  const next = {
    ...metrics,
    buffer_payload_bytes: overview.bufferPayloadBytes || 0,
    buffer_payload_limit: overview.bufferPayloadLimit || 0,
    buffer_pending_entries: overview.bufferPendingEntries || 0,
    buffer_fill_percent: overview.bufferFillPercent || 0,
  }

  raw.value = next
  if (latestSample.value) {
    previousSample.value = latestSample.value
  }
  latestSample.value = {
    capturedAtMs: Date.now(),
    metrics: next,
  }
}

function fmtBytes(b) {
  if (b >= 1073741824) return (b / 1073741824).toFixed(2) + ' GiB'
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MiB'
  if (b >= 1024) return (b / 1024).toFixed(0) + ' KiB'
  return b + ' B'
}

function fmtNs(ns) {
  if (ns >= 1e9) return (ns / 1e9).toFixed(2) + ' s'
  if (ns >= 1e6) return (ns / 1e6).toFixed(1) + ' ms'
  if (ns >= 1e3) return (ns / 1e3).toFixed(0) + ' us'
  return ns + ' ns'
}

function fmtAvgNs(totalNs, ops) {
  if (!ops) return 'n/a'
  return fmtNs(totalNs / ops)
}

function fmtBytesPair(used, total) {
  if (!total) return fmtBytes(used || 0)
  return `${fmtBytes(used || 0)} / ${fmtBytes(total)}`
}

function fmtOpsPerSec(value) {
  if (value == null) return 'n/a'
  if (value >= 1000) return value.toFixed(0) + '/s'
  if (value >= 100) return value.toFixed(1) + '/s'
  return value.toFixed(2) + '/s'
}

function fmtBytesPerSec(value) {
  if (value == null) return 'n/a'
  return fmtBytes(value) + '/s'
}

const m = computed(() => raw.value)

const compressRatio = computed(() => {
  const d = m.value
  if (!d.compress_output_bytes) return '1.00x'
  return (d.compress_input_bytes / d.compress_output_bytes).toFixed(2) + 'x'
})

const dedupRate = computed(() => {
  const d = m.value
  const total = (d.dedup_hits || 0) + (d.dedup_misses || 0)
  if (!total) return '0.0%'
  return ((d.dedup_hits / total) * 100).toFixed(1) + '%'
})

const rateWindowSecs = computed(() => {
  if (!previousSample.value || !latestSample.value) return 0
  const deltaMs = latestSample.value.capturedAtMs - previousSample.value.capturedAtMs
  if (deltaMs <= 0) return 0
  return deltaMs / 1000
})

const rateWindowLabel = computed(() => {
  if (!rateWindowSecs.value) return 'warming up'
  return `${rateWindowSecs.value.toFixed(1)}s window`
})

function metricDelta(field) {
  if (!previousSample.value || !latestSample.value || !rateWindowSecs.value) {
    return null
  }
  const now = latestSample.value.metrics?.[field] || 0
  const before = previousSample.value.metrics?.[field] || 0
  return Math.max(0, now - before)
}

function buildRateCard(key, title, note, readOpsField, readBytesField, writeOpsField, writeBytesField) {
  const secs = rateWindowSecs.value
  if (!secs) {
    return {
      key,
      title,
      note,
      readThroughput: 'n/a',
      readIops: 'n/a',
      writeThroughput: 'n/a',
      writeIops: 'n/a',
    }
  }

  const readOps = metricDelta(readOpsField) / secs
  const readBytes = metricDelta(readBytesField) / secs
  const writeOps = metricDelta(writeOpsField) / secs
  const writeBytes = metricDelta(writeBytesField) / secs

  return {
    key,
    title,
    note,
    readThroughput: fmtBytesPerSec(readBytes),
    readIops: fmtOpsPerSec(readOps),
    writeThroughput: fmtBytesPerSec(writeBytes),
    writeIops: fmtOpsPerSec(writeOps),
  }
}

const ioRateCards = computed(() => [
  buildRateCard(
    'client',
    'Client IO',
    'Front-end logical traffic seen by the volume API.',
    'volume_read_ops',
    'volume_read_bytes',
    'volume_write_ops',
    'volume_write_bytes',
  ),
  buildRateCard(
    'cache',
    'Cache IO',
    'LV2 buffer activity: append writes and read hits served from cache.',
    'buffer_read_ops',
    'buffer_read_bytes',
    'buffer_write_ops',
    'buffer_write_bytes',
  ),
  buildRateCard(
    'lv3',
    'LV3 IO',
    'Physical device traffic issued by IoEngine, including background work.',
    'lv3_read_ops',
    'lv3_read_bytes',
    'lv3_write_ops',
    'lv3_write_bytes',
  ),
])

const metricGroups = computed(() => {
  const d = m.value
  if (!d.uptime_secs && d.uptime_secs !== 0) return []
  return [
    {
      title: 'Volume IO',
      items: [
        { key: 'uptime', label: 'uptime', value: d.uptime_secs + 's' },
        { key: 'read_ops', label: 'read ops', value: d.volume_read_ops },
        { key: 'write_ops', label: 'write ops', value: d.volume_write_ops },
        { key: 'read_bytes', label: 'read bytes', value: fmtBytes(d.volume_read_bytes) },
        { key: 'write_bytes', label: 'write bytes', value: fmtBytes(d.volume_write_bytes) },
        { key: 'create', label: 'volume creates', value: d.volume_create_ops },
        { key: 'delete', label: 'volume deletes', value: d.volume_delete_ops },
      ],
    },
    {
      title: 'Buffer',
      items: [
        { key: 'appends', label: 'appends', value: d.buffer_appends },
        { key: 'append_bytes', label: 'append bytes', value: fmtBytes(d.buffer_append_bytes) },
        { key: 'write_ops', label: 'cache write ops', value: d.buffer_write_ops || 0 },
        { key: 'write_bytes', label: 'cache write bytes', value: fmtBytes(d.buffer_write_bytes || 0) },
        { key: 'read_ops', label: 'cache read ops', value: d.buffer_read_ops || 0 },
        { key: 'read_bytes', label: 'cache read bytes', value: fmtBytes(d.buffer_read_bytes || 0) },
        { key: 'payload_mem', label: 'payload memory', value: fmtBytesPair(d.buffer_payload_bytes, d.buffer_payload_limit) },
        { key: 'buffer_fill', label: 'buffer fill', value: `${d.buffer_fill_percent || 0}%` },
        { key: 'pending_entries', label: 'pending entries', value: d.buffer_pending_entries || 0 },
        { key: 'backpressure_events', label: 'backpressure events', value: d.buffer_backpressure_events || 0 },
        { key: 'backpressure_wait_total', label: 'backpressure wait total', value: fmtNs(d.buffer_backpressure_wait_ns || 0) },
        { key: 'backpressure_wait_avg', label: 'backpressure wait avg', value: fmtAvgNs(d.buffer_backpressure_wait_ns, d.buffer_backpressure_events) },
        { key: 'hydration_skips', label: 'hydration skips', value: d.buffer_hydration_skipped_due_to_mem_limit || 0 },
        { key: 'hydration_bypass', label: 'head bypass count', value: d.buffer_hydration_head_bypass_count || 0 },
        { key: 'lookup_hits', label: 'lookup hits', value: d.buffer_lookup_hits },
        { key: 'lookup_misses', label: 'lookup misses', value: d.buffer_lookup_misses },
        { key: 'read_buf_hits', label: 'read buffer hits', value: d.read_buffer_hits },
        { key: 'read_unmapped', label: 'read unmapped', value: d.read_unmapped },
        { key: 'crc_errors', label: 'CRC errors', value: d.read_crc_errors },
      ],
    },
    {
      title: 'LV3 IO',
      items: [
        { key: 'lv3_read_ops', label: 'lv3 read ops', value: d.lv3_read_ops || 0 },
        { key: 'lv3_read_bytes', label: 'lv3 read bytes', value: fmtBytes(d.lv3_read_bytes || 0) },
        { key: 'lv3_write_ops', label: 'lv3 write ops', value: d.lv3_write_ops || 0 },
        { key: 'lv3_write_bytes', label: 'lv3 write bytes', value: fmtBytes(d.lv3_write_bytes || 0) },
        { key: 'read_lv3_hits', label: 'logical LV3 hits', value: d.read_lv3_hits || 0 },
      ],
    },
    {
      title: 'Flush + Compress',
      items: [
        { key: 'compress_ratio', label: 'compression ratio', value: compressRatio.value },
        { key: 'coalesce', label: 'coalesce runs', value: d.coalesce_runs },
        { key: 'coalesced_u', label: 'coalesced units', value: d.coalesced_units },
        { key: 'coalesced_b', label: 'coalesced bytes', value: fmtBytes(d.coalesced_bytes) },
        { key: 'compress_u', label: 'compress units', value: d.compress_units },
        { key: 'compress_in', label: 'compress in', value: fmtBytes(d.compress_input_bytes) },
        { key: 'compress_out', label: 'compress out', value: fmtBytes(d.compress_output_bytes) },
        { key: 'flush_units', label: 'flush units written', value: d.flush_units_written },
        { key: 'flush_bytes', label: 'flush bytes', value: fmtBytes(d.flush_unit_bytes) },
        { key: 'packed_slots', label: 'packed slots', value: d.flush_packed_slots_written },
        { key: 'packed_bytes', label: 'packed bytes', value: fmtBytes(d.flush_packed_bytes) },
        { key: 'precheck_ops', label: 'live-pba prechecks', value: d.flush_writer_precheck_live_pba_ops || 0 },
        { key: 'precheck_failures', label: 'precheck failures', value: d.flush_writer_precheck_live_pba_failures || 0 },
        { key: 'precheck_avg', label: 'precheck avg latency', value: fmtAvgNs(d.flush_writer_precheck_live_pba_ns, d.flush_writer_precheck_live_pba_ops) },
        { key: 'flush_errors', label: 'flush errors', value: d.flush_errors },
      ],
    },
    {
      title: 'Dedup',
      items: [
        { key: 'dedup_rate', label: 'dedup hit rate', value: dedupRate.value },
        { key: 'hits', label: 'dedup hits', value: d.dedup_hits },
        { key: 'misses', label: 'dedup misses', value: d.dedup_misses },
        { key: 'hit_failures', label: 'dedup hit failures', value: d.dedup_hit_failures || 0 },
        { key: 'saved', label: 'space saved by dedup', value: fmtBytes(d.dedup_hits * 4096) },
        { key: 'skipped', label: 'skipped units', value: d.dedup_skipped_units },
        { key: 'lookup_ops', label: 'lookup ops', value: d.dedup_lookup_ops || 0 },
        { key: 'lookup_avg', label: 'lookup avg latency', value: fmtAvgNs(d.dedup_lookup_ns, d.dedup_lookup_ops) },
        { key: 'live_check_ops', label: 'live checks', value: d.dedup_live_check_ops || 0 },
        { key: 'live_check_avg', label: 'live check avg latency', value: fmtAvgNs(d.dedup_live_check_ns, d.dedup_live_check_ops) },
        { key: 'stale_entries', label: 'stale index entries', value: d.dedup_stale_index_entries || 0 },
        { key: 'stale_delete_total', label: 'stale delete total', value: fmtNs(d.dedup_stale_delete_ns || 0) },
        { key: 'stale_delete_avg', label: 'stale delete avg latency', value: fmtAvgNs(d.dedup_stale_delete_ns, d.dedup_stale_index_entries) },
        { key: 'hit_commit_ops', label: 'hit commits', value: d.dedup_hit_commit_ops || 0 },
        { key: 'hit_commit_avg', label: 'hit commit avg latency', value: fmtAvgNs(d.dedup_hit_commit_ns, d.dedup_hit_commit_ops) },
        { key: 'rescan_cycles', label: 'rescan cycles', value: d.dedup_rescan_cycles },
        { key: 'rescan_hits', label: 'rescan hits', value: d.dedup_rescan_hits },
        { key: 'rescan_errors', label: 'rescan errors', value: d.dedup_rescan_errors },
      ],
    },
    {
      title: 'GC',
      items: [
        { key: 'gc_cycles', label: 'scan cycles', value: d.gc_cycles },
        { key: 'gc_found', label: 'candidates found', value: d.gc_candidates_found },
        { key: 'gc_rewritten', label: 'blocks rewritten', value: d.gc_blocks_rewritten },
        { key: 'gc_errors', label: 'errors', value: d.gc_errors },
      ],
    },
  ]
})

const logout = () => {
  auth.logout()
  router.push('/login')
}

onMounted(async () => {
  await auth.fetchMe()
  await load()
  pollHandle = window.setInterval(load, POLL_INTERVAL_MS)
})

onBeforeUnmount(() => {
  if (pollHandle) {
    window.clearInterval(pollHandle)
  }
})
</script>

<style scoped>
.lane-note {
  margin: 0 0 1rem;
  color: var(--onyx-muted);
  font-size: 0.92rem;
}

.rate-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
}

.rate-item {
  display: grid;
  gap: 0.25rem;
  padding: 0.9rem 1rem;
  border-radius: 1rem;
  background: #f4f9fc;
  border: 1px solid rgba(46, 111, 206, 0.08);
}

.rate-item span {
  color: #55738f;
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.rate-item strong {
  color: var(--onyx-text);
  font-size: 1.05rem;
}

@media (max-width: 640px) {
  .rate-grid {
    grid-template-columns: 1fr;
  }
}
</style>
