<template>
  <section class="content-card metadb-panel">
    <div class="section-header">
      <div>
        <div class="tiny-label">MetaDB</div>
        <h3>Database state</h3>
      </div>
      <div class="metadb-badges">
        <span class="badge text-bg-dark">LSN {{ formatNumber(n(stats?.last_applied_lsn)) }}</span>
        <span class="badge text-bg-dark">{{ formatNumber(n(stats?.high_water_pages)) }} pages</span>
      </div>
    </div>

    <div v-if="!stats" class="empty-note">MetaDB status is not reported by this engine yet.</div>

    <template v-else>
      <div class="metadb-grid">
        <div class="metadb-block signal-block">
          <div class="block-title">
            <i class="bi bi-activity"></i>
            <span>Bottlenecks</span>
          </div>
          <div class="signal-list">
            <div v-for="signal in bottleneckSignals" :key="signal.label" class="signal-row" :class="signal.severity">
              <span class="signal-dot"></span>
              <div>
                <strong>{{ signal.label }}</strong>
                <span>{{ signal.value }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="metadb-block">
          <div class="block-title">
            <i class="bi bi-cpu"></i>
            <span>Page cache</span>
          </div>
          <div class="cache-meter">
            <div class="meter-head">
              <span>Hit rate</span>
              <strong>{{ percentOrDash(cacheHitRate) }}</strong>
            </div>
            <div class="meter-track">
              <div class="meter-fill hit" :style="{ width: boundedWidth(cacheHitRate) }"></div>
            </div>
          </div>
          <div class="cache-meter">
            <div class="meter-head">
              <span>Usage</span>
              <strong>{{ cacheUsageLabel }}</strong>
            </div>
            <div class="meter-track">
              <div class="meter-fill usage" :style="{ width: boundedWidth(cacheUsagePct) }"></div>
            </div>
          </div>
          <div class="cache-stats">
            <div>
              <span>Hits</span>
              <strong>{{ formatNumber(n(stats.cache_hits)) }}</strong>
            </div>
            <div>
              <span>Misses</span>
              <strong>{{ formatNumber(n(stats.cache_misses)) }}</strong>
            </div>
            <div>
              <span>Evicts</span>
              <strong>{{ formatNumber(n(stats.cache_evictions)) }}</strong>
            </div>
            <div>
              <span>Pinned</span>
              <strong>{{ formatNumber(n(stats.cache_pinned_pages)) }}</strong>
            </div>
          </div>
        </div>
      </div>

      <div class="metadb-table-grid">
        <div class="metric-table">
          <div class="block-title">
            <i class="bi bi-check2-square"></i>
            <span>Commit</span>
          </div>
          <div v-for="row in commitRows" :key="row.label" class="metric-row">
            <span>{{ row.label }}</span>
            <strong>{{ row.value }}</strong>
          </div>
        </div>

        <div class="metric-table">
          <div class="block-title">
            <i class="bi bi-journal-arrow-down"></i>
            <span>WAL</span>
          </div>
          <div v-for="row in walRows" :key="row.label" class="metric-row">
            <span>{{ row.label }}</span>
            <strong>{{ row.value }}</strong>
          </div>
        </div>

        <div class="metric-table">
          <div class="block-title">
            <i class="bi bi-eraser"></i>
            <span>Range delete</span>
          </div>
          <div v-for="row in rangeDeleteRows" :key="row.label" class="metric-row">
            <span>{{ row.label }}</span>
            <strong>{{ row.value }}</strong>
          </div>
        </div>

        <div class="metric-table">
          <div class="block-title">
            <i class="bi bi-recycle"></i>
            <span>Cleanup</span>
          </div>
          <div v-for="row in cleanupRows" :key="row.label" class="metric-row">
            <span>{{ row.label }}</span>
            <strong>{{ row.value }}</strong>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { formatBytes, formatNumber, formatPercent } from '../lib/telemetry'

const props = defineProps({
  stats: { type: Object, default: null },
})

const n = (value) => Number(value ?? 0)
const ratio = (part, total) => (total > 0 ? (part / total) * 100 : null)
const avg = (total, count) => (count > 0 ? total / count : 0)

const cacheCapacity = computed(() =>
  n(props.stats?.block_cache_capacity_bytes ?? props.stats?.cache_capacity_bytes),
)
const cacheUsage = computed(() => n(props.stats?.block_cache_usage_bytes))
const cacheUsagePct = computed(() => ratio(cacheUsage.value, cacheCapacity.value))
const cacheHitRate = computed(() => {
  const hits = n(props.stats?.cache_hits)
  const misses = n(props.stats?.cache_misses)
  return ratio(hits, hits + misses)
})
const cacheMissRate = computed(() => (cacheHitRate.value == null ? null : 100 - cacheHitRate.value))
const pinnedBytesPct = computed(() =>
  ratio(n(props.stats?.block_cache_pinned_usage_bytes), n(props.stats?.cache_pin_budget_bytes)),
)
const commitAvgUs = computed(() => avg(n(props.stats?.commit_total_us), n(props.stats?.commit_attempts)))
const commitWaitAvgUs = computed(() =>
  avg(n(props.stats?.commit_apply_wait_us), n(props.stats?.commit_attempts)),
)
const commitApplyAvgUs = computed(() => avg(n(props.stats?.commit_apply_us), n(props.stats?.commit_attempts)))
const walWriteAvgUs = computed(() => avg(n(props.stats?.wal_write_us), n(props.stats?.wal_batches)))
const walFsyncAvgUs = computed(() => avg(n(props.stats?.wal_fsync_us), n(props.stats?.wal_fsyncs)))
const rangeDeleteAvgUs = computed(() =>
  avg(n(props.stats?.range_delete_total_us), n(props.stats?.range_delete_calls)),
)
const cleanupAvgUs = computed(() => avg(n(props.stats?.cleanup_total_us), n(props.stats?.cleanup_calls)))

const cacheUsageLabel = computed(() => {
  if (!cacheCapacity.value) return '--'
  return `${formatPercent(cacheUsagePct.value)} · ${formatBytes(cacheUsage.value)} / ${formatBytes(cacheCapacity.value)}`
})

const bottleneckSignals = computed(() => {
  if (!props.stats) return []

  const signals = []
  const errors =
    n(props.stats.commit_errors) + n(props.stats.range_delete_errors) + n(props.stats.cleanup_errors)

  if (errors > 0) {
    signals.push({ severity: 'danger', label: 'Metadata errors', value: formatNumber(errors) })
  }
  if (cacheMissRate.value != null && cacheMissRate.value >= 15) {
    signals.push({ severity: 'danger', label: 'Cache miss pressure', value: formatPercent(cacheMissRate.value) })
  } else if (cacheMissRate.value != null && cacheMissRate.value >= 5) {
    signals.push({ severity: 'warn', label: 'Cache miss pressure', value: formatPercent(cacheMissRate.value) })
  }
  if (cacheUsagePct.value != null && cacheUsagePct.value >= 95) {
    signals.push({ severity: 'danger', label: 'Cache almost full', value: formatPercent(cacheUsagePct.value) })
  } else if (cacheUsagePct.value != null && cacheUsagePct.value >= 85) {
    signals.push({ severity: 'warn', label: 'Cache filling', value: formatPercent(cacheUsagePct.value) })
  }
  if (pinnedBytesPct.value != null && pinnedBytesPct.value >= 90) {
    signals.push({ severity: 'warn', label: 'Pinned page budget', value: formatPercent(pinnedBytesPct.value) })
  }
  if (commitWaitAvgUs.value >= 5000) {
    signals.push({ severity: 'danger', label: 'Commit wait', value: formatDurationUs(commitWaitAvgUs.value) })
  } else if (commitWaitAvgUs.value >= 1000) {
    signals.push({ severity: 'warn', label: 'Commit wait', value: formatDurationUs(commitWaitAvgUs.value) })
  }
  if (walFsyncAvgUs.value >= 20000) {
    signals.push({ severity: 'danger', label: 'WAL fsync', value: formatDurationUs(walFsyncAvgUs.value) })
  } else if (walFsyncAvgUs.value >= 5000) {
    signals.push({ severity: 'warn', label: 'WAL fsync', value: formatDurationUs(walFsyncAvgUs.value) })
  }
  if (signals.length === 0) {
    signals.push({ severity: 'ok', label: 'No obvious bottleneck', value: 'green' })
  }

  return signals
})

const commitRows = computed(() => [
  { label: 'Attempts', value: formatNumber(n(props.stats?.commit_attempts)) },
  { label: 'Success / errors', value: `${formatNumber(n(props.stats?.commit_success))} / ${formatNumber(n(props.stats?.commit_errors))}` },
  { label: 'Avg / max', value: `${formatDurationUs(commitAvgUs.value)} / ${formatDurationUs(n(props.stats?.commit_total_max_us))}` },
  { label: 'Apply wait avg / max', value: `${formatDurationUs(commitWaitAvgUs.value)} / ${formatDurationUs(n(props.stats?.commit_apply_wait_max_us))}` },
  { label: 'Apply avg / max', value: `${formatDurationUs(commitApplyAvgUs.value)} / ${formatDurationUs(n(props.stats?.commit_apply_max_us))}` },
])

const walRows = computed(() => [
  { label: 'Batches / records', value: `${formatNumber(n(props.stats?.wal_batches))} / ${formatNumber(n(props.stats?.wal_records))}` },
  { label: 'Bytes', value: formatBytes(n(props.stats?.wal_bytes)) },
  { label: 'Write avg / max', value: `${formatDurationUs(walWriteAvgUs.value)} / ${formatDurationUs(n(props.stats?.wal_write_max_us))}` },
  { label: 'Fsync avg / max', value: `${formatDurationUs(walFsyncAvgUs.value)} / ${formatDurationUs(n(props.stats?.wal_fsync_max_us))}` },
  { label: 'Rotates / fsyncs', value: `${formatNumber(n(props.stats?.wal_rotates))} / ${formatNumber(n(props.stats?.wal_fsyncs))}` },
])

const rangeDeleteRows = computed(() => [
  { label: 'Calls / noop', value: `${formatNumber(n(props.stats?.range_delete_calls))} / ${formatNumber(n(props.stats?.range_delete_noop))}` },
  { label: 'Captured', value: formatNumber(n(props.stats?.range_delete_captured_entries)) },
  { label: 'Chunks', value: formatNumber(n(props.stats?.range_delete_chunks)) },
  { label: 'Avg / max', value: `${formatDurationUs(rangeDeleteAvgUs.value)} / ${formatDurationUs(n(props.stats?.range_delete_total_max_us))}` },
  { label: 'Errors', value: formatNumber(n(props.stats?.range_delete_errors)) },
])

const cleanupRows = computed(() => [
  { label: 'Calls / noop', value: `${formatNumber(n(props.stats?.cleanup_calls))} / ${formatNumber(n(props.stats?.cleanup_noop))}` },
  { label: 'PBAs / hashes', value: `${formatNumber(n(props.stats?.cleanup_pbas))} / ${formatNumber(n(props.stats?.cleanup_hashes_found))}` },
  { label: 'Tombstones', value: formatNumber(n(props.stats?.cleanup_tombstones_emitted)) },
  { label: 'Tx ops', value: formatNumber(n(props.stats?.cleanup_tx_ops)) },
  { label: 'Avg / max', value: `${formatDurationUs(cleanupAvgUs.value)} / ${formatDurationUs(n(props.stats?.cleanup_total_max_us))}` },
])

function boundedWidth(value) {
  if (value == null || !Number.isFinite(value)) return '0%'
  return `${Math.max(0, Math.min(100, value)).toFixed(1)}%`
}

function percentOrDash(value) {
  if (value == null || !Number.isFinite(value)) return '--'
  return formatPercent(value)
}

function formatDurationUs(value) {
  const numeric = Number(value || 0)
  if (numeric >= 1_000_000) return `${(numeric / 1_000_000).toFixed(2)} s`
  if (numeric >= 1000) return `${(numeric / 1000).toFixed(2)} ms`
  return `${numeric.toFixed(0)} us`
}
</script>

<style scoped>
.metadb-panel {
  display: grid;
  gap: 1rem;
}

.metadb-badges {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.metadb-grid,
.metadb-table-grid {
  display: grid;
  gap: 0.75rem;
}

.metadb-grid {
  grid-template-columns: minmax(18rem, 0.85fr) minmax(0, 1.15fr);
}

.metadb-table-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.metadb-block,
.metric-table {
  display: grid;
  gap: 0.75rem;
  min-width: 0;
  padding: 0.875rem;
  border: 1px solid var(--onyx-border);
  border-radius: var(--onyx-radius-sm);
  background: var(--onyx-surface-soft);
}

.block-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--onyx-text-secondary);
  font-size: 0.8125rem;
  font-weight: 700;
}

.block-title i {
  color: var(--onyx-primary);
}

.signal-list {
  display: grid;
  gap: 0.5rem;
}

.signal-row {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.625rem;
  align-items: center;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--onyx-border);
  border-radius: var(--onyx-radius-xs);
  background: var(--onyx-surface);
}

.signal-row > div {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.signal-row span:last-child {
  color: var(--onyx-muted);
  white-space: nowrap;
}

.signal-dot {
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 50%;
  background: var(--onyx-primary);
}

.signal-row.ok .signal-dot {
  background: #0d9488;
}

.signal-row.warn .signal-dot {
  background: #f59e0b;
}

.signal-row.danger .signal-dot {
  background: #ef4444;
}

.cache-meter {
  display: grid;
  gap: 0.375rem;
}

.meter-head,
.metric-row,
.cache-stats {
  display: flex;
}

.meter-head,
.metric-row {
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
  color: var(--onyx-muted);
  font-size: 0.8125rem;
}

.meter-head strong,
.metric-row strong {
  color: var(--onyx-text);
  font-weight: 700;
  text-align: right;
  overflow-wrap: anywhere;
}

.meter-track {
  height: 0.5rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.18);
}

.meter-fill {
  height: 100%;
  border-radius: inherit;
}

.meter-fill.hit {
  background: #0d9488;
}

.meter-fill.usage {
  background: #2563eb;
}

.cache-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.5rem;
}

.cache-stats > div {
  display: grid;
  gap: 0.125rem;
  padding: 0.5rem;
  border: 1px solid var(--onyx-border);
  border-radius: var(--onyx-radius-xs);
  background: var(--onyx-surface);
}

.cache-stats span {
  color: var(--onyx-muted);
  font-size: 0.75rem;
}

.cache-stats strong {
  font-size: 0.875rem;
}

@media (max-width: 1180px) {
  .metadb-grid,
  .metadb-table-grid {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 760px) {
  .metadb-grid,
  .metadb-table-grid,
  .cache-stats {
    grid-template-columns: 1fr;
  }

  .signal-row > div,
  .metric-row,
  .meter-head {
    flex-direction: column;
    gap: 0.125rem;
  }

  .metric-row strong,
  .meter-head strong {
    text-align: left;
  }
}
</style>
