<template>
  <AppShell title="Metrics" eyebrow="Pipeline Telemetry" :user="auth.user" @logout="logout">
    <template #header-actions>
      <button class="btn btn-accent" @click="load">刷新</button>
    </template>

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
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import AppShell from '../components/AppShell.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const raw = ref({})

const load = async () => {
  const { data } = await http.get('/metrics/summary')
  raw.value = data || {}
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
        { key: 'lookup_hits', label: 'lookup hits', value: d.buffer_lookup_hits },
        { key: 'lookup_misses', label: 'lookup misses', value: d.buffer_lookup_misses },
        { key: 'read_buf_hits', label: 'read buffer hits', value: d.read_buffer_hits },
        { key: 'read_lv3', label: 'read LV3 hits', value: d.read_lv3_hits },
        { key: 'read_unmapped', label: 'read unmapped', value: d.read_unmapped },
        { key: 'crc_errors', label: 'CRC errors', value: d.read_crc_errors },
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
        { key: 'flush_errors', label: 'flush errors', value: d.flush_errors },
      ],
    },
    {
      title: 'Dedup',
      items: [
        { key: 'dedup_rate', label: 'dedup hit rate', value: dedupRate.value },
        { key: 'hits', label: 'dedup hits', value: d.dedup_hits },
        { key: 'misses', label: 'dedup misses', value: d.dedup_misses },
        { key: 'saved', label: 'space saved by dedup', value: fmtBytes(d.dedup_hits * 4096) },
        { key: 'skipped', label: 'skipped units', value: d.dedup_skipped_units },
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
})
</script>
