<template>
  <AppShell title="config.title" eyebrow="TOML Configuration" :user="auth.user" @logout="logout">
    <div class="row g-4">
      <div class="col-12 col-xl-7">
        <div class="content-card">
          <div class="section-header">
            <h3>{{ $t('config.configEdit') }}</h3>
            <button class="btn btn-sm btn-outline-light" @click="loadConfig">{{ $t('common.refresh') }}</button>
          </div>

          <div v-if="loading" class="text-center py-5 text-muted">{{ $t('common.loading') }}</div>

          <div v-else class="config-sections">
            <!-- Meta -->
            <details class="config-group" open>
              <summary>Meta (RocksDB)</summary>
              <div class="config-fields">
                <label>rocksdb_path</label>
                <input v-model="config.meta.rocksdb_path" class="form-control" placeholder="/var/lib/onyx-storage/meta" />
                <label>block_cache_mb</label>
                <input v-model.number="config.meta.block_cache_mb" type="number" class="form-control" placeholder="256" />
                <label>wal_dir</label>
                <input v-model="config.meta.wal_dir" class="form-control" placeholder="(optional)" />
              </div>
            </details>

            <!-- Storage -->
            <details class="config-group" open>
              <summary>Storage (LV3 Data Device)</summary>
              <div class="config-fields">
                <label>data_device</label>
                <input v-model="config.storage.data_device" class="form-control" placeholder="/dev/vg0/onyx-data" />
                <label>block_size</label>
                <input v-model.number="config.storage.block_size" type="number" class="form-control" placeholder="4096" />
                <label>use_hugepages</label>
                <div class="form-check form-switch">
                  <input v-model="config.storage.use_hugepages" class="form-check-input" type="checkbox" />
                </div>
                <label>default_compression</label>
                <select v-model="config.storage.default_compression" class="form-select">
                  <option value="Lz4">LZ4</option>
                  <option value="Zstd">ZSTD</option>
                  <option value="None">None</option>
                </select>
              </div>
            </details>

            <!-- Buffer -->
            <details class="config-group" open>
              <summary>Buffer (LV2 Write Buffer)</summary>
              <div class="config-fields">
                <label>device</label>
                <input v-model="config.buffer.device" class="form-control" placeholder="/dev/vg0/onyx-buffer" />
                <label>capacity_mb</label>
                <input v-model.number="config.buffer.capacity_mb" type="number" class="form-control" placeholder="16384" />
                <label>flush_watermark_pct</label>
                <input v-model.number="config.buffer.flush_watermark_pct" type="number" class="form-control" placeholder="80" />
                <label>group_commit_wait_us</label>
                <input v-model.number="config.buffer.group_commit_wait_us" type="number" class="form-control" placeholder="250" />
                <label>shards</label>
                <input v-model.number="config.buffer.shards" type="number" class="form-control" placeholder="4" />
              </div>
            </details>

            <!-- Engine -->
            <details class="config-group">
              <summary>Engine</summary>
              <div class="config-fields">
                <label>zone_count</label>
                <input v-model.number="config.engine.zone_count" type="number" class="form-control" placeholder="4" />
                <label>zone_size_blocks</label>
                <input v-model.number="config.engine.zone_size_blocks" type="number" class="form-control" placeholder="256" />
              </div>
            </details>

            <!-- GC -->
            <details class="config-group">
              <summary>GC</summary>
              <div class="config-fields">
                <label>enabled</label>
                <div class="form-check form-switch">
                  <input v-model="config.gc.enabled" class="form-check-input" type="checkbox" />
                </div>
                <label>scan_interval_ms</label>
                <input v-model.number="config.gc.scan_interval_ms" type="number" class="form-control" placeholder="5000" />
                <label>dead_ratio_threshold</label>
                <input v-model.number="config.gc.dead_ratio_threshold" type="number" step="0.01" class="form-control" placeholder="0.25" />
                <label>buffer_usage_max_pct</label>
                <input v-model.number="config.gc.buffer_usage_max_pct" type="number" class="form-control" placeholder="80" />
                <label>buffer_usage_resume_pct</label>
                <input v-model.number="config.gc.buffer_usage_resume_pct" type="number" class="form-control" placeholder="50" />
                <label>max_rewrite_per_cycle</label>
                <input v-model.number="config.gc.max_rewrite_per_cycle" type="number" class="form-control" placeholder="64" />
              </div>
            </details>

            <!-- Dedup -->
            <details class="config-group">
              <summary>Dedup</summary>
              <div class="config-fields">
                <label>enabled</label>
                <div class="form-check form-switch">
                  <input v-model="config.dedup.enabled" class="form-check-input" type="checkbox" />
                </div>
                <label>workers</label>
                <input v-model.number="config.dedup.workers" type="number" class="form-control" placeholder="2" />
                <label>buffer_skip_threshold_pct</label>
                <input v-model.number="config.dedup.buffer_skip_threshold_pct" type="number" class="form-control" placeholder="90" />
                <label>rescan_interval_ms</label>
                <input v-model.number="config.dedup.rescan_interval_ms" type="number" class="form-control" placeholder="30000" />
                <label>max_rescan_per_cycle</label>
                <input v-model.number="config.dedup.max_rescan_per_cycle" type="number" class="form-control" placeholder="256" />
              </div>
            </details>

            <!-- Flush -->
            <details class="config-group">
              <summary>Flush</summary>
              <div class="config-fields">
                <label>compress_workers</label>
                <input v-model.number="config.flush.compress_workers" type="number" class="form-control" placeholder="2" />
                <label>coalesce_max_raw_bytes</label>
                <input v-model.number="config.flush.coalesce_max_raw_bytes" type="number" class="form-control" placeholder="131072" />
                <label>coalesce_max_lbas</label>
                <input v-model.number="config.flush.coalesce_max_lbas" type="number" class="form-control" placeholder="32" />
              </div>
            </details>

            <!-- Ublk -->
            <details class="config-group">
              <summary>Ublk</summary>
              <div class="config-fields">
                <label>nr_queues</label>
                <input v-model.number="config.ublk.nr_queues" type="number" class="form-control" placeholder="4" />
                <label>queue_depth</label>
                <input v-model.number="config.ublk.queue_depth" type="number" class="form-control" placeholder="128" />
                <label>io_buf_bytes</label>
                <input v-model.number="config.ublk.io_buf_bytes" type="number" class="form-control" placeholder="1048576" />
              </div>
            </details>

            <!-- Service -->
            <details class="config-group">
              <summary>Service</summary>
              <div class="config-fields">
                <label>socket_path</label>
                <input v-model="config.service.socket_path" class="form-control" placeholder="/var/run/onyx-storage.sock" />
              </div>
            </details>
          </div>
        </div>
      </div>

      <!-- Right panel: Status + Actions -->
      <div class="col-12 col-xl-5">
        <div class="content-card">
          <div class="section-header">
            <h3>{{ $t('config.engineStatus') }}</h3>
          </div>

          <div class="mode-display mb-4">
            <div class="d-flex align-items-center gap-2 mb-2">
              <span class="fw-semibold">{{ $t('config.currentMode') }}</span>
              <span :class="modeBadgeClass">{{ mode }}</span>
            </div>
            <p class="text-muted small mb-0">{{ modeHint }}</p>
          </div>

          <div class="mode-steps mb-4">
            <div :class="['mode-step', { active: modeIndex >= 0 }]">
              <i class="bi bi-check-circle-fill" v-if="modeIndex >= 0"></i>
              <i class="bi bi-circle" v-else></i>
              <span>IPC Socket</span>
            </div>
            <div :class="['mode-step', { active: modeIndex >= 1 }]">
              <i class="bi bi-check-circle-fill" v-if="modeIndex >= 1"></i>
              <i class="bi bi-circle" v-else></i>
              <span>Meta (RocksDB)</span>
            </div>
            <div :class="['mode-step', { active: modeIndex >= 2 }]">
              <i class="bi bi-check-circle-fill" v-if="modeIndex >= 2"></i>
              <i class="bi bi-circle" v-else></i>
              <span>Storage + Buffer</span>
            </div>
          </div>

          <hr />

          <div class="d-grid gap-2">
            <button class="btn btn-accent" @click="saveConfig" :disabled="saving">
              <i class="bi bi-save me-1"></i>
              {{ saving ? $t('config.saving') : $t('config.saveConfig') }}
            </button>
            <button class="btn btn-primary" @click="reloadEngine" :disabled="reloading">
              <i class="bi bi-arrow-clockwise me-1"></i>
              {{ reloading ? $t('config.reloading') : $t('config.hotReload') }}
            </button>
            <button class="btn btn-outline-danger" @click="restartService" :disabled="restarting">
              <i class="bi bi-power me-1"></i>
              {{ restarting ? $t('config.restarting') : $t('config.restartService') }}
            </button>
          </div>
          <p class="text-muted small mt-2 mb-0">
            {{ $t('config.hotReloadNote') }}<br>
            {{ $t('config.restartNote') }}
          </p>

          <div v-if="message" :class="['alert', messageClass, 'mt-3', 'mb-0']" role="alert">
            {{ message }}
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AppShell from '../components/AppShell.vue'
import http from '../api/http'
import { translateError } from '../i18n/errorMap'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()

const loading = ref(true)
const saving = ref(false)
const reloading = ref(false)
const restarting = ref(false)
const message = ref('')
const messageClass = ref('alert-success')
const mode = ref('unknown')

const config = ref({
  meta: {},
  storage: {},
  buffer: {},
  ublk: {},
  flush: {},
  engine: {},
  gc: {},
  dedup: {},
  service: {},
})

const modeIndex = computed(() => {
  switch (mode.value) {
    case 'active': return 2
    case 'standby': return 1
    case 'bare': return 0
    default: return -1
  }
})

const modeBadgeClass = computed(() => {
  switch (mode.value) {
    case 'active': return 'badge text-bg-success'
    case 'standby': return 'badge text-bg-warning'
    case 'bare': return 'badge text-bg-secondary'
    default: return 'badge text-bg-danger'
  }
})

const modeHint = computed(() => {
  switch (mode.value) {
    case 'bare': return t('config.bareHint')
    case 'standby': return t('config.standbyHint')
    case 'active': return t('config.activeHint')
    default: return t('config.unknownHint')
  }
})

function showMessage(text, isError = false) {
  message.value = text
  messageClass.value = isError ? 'alert-danger' : 'alert-success'
  setTimeout(() => { message.value = '' }, 5000)
}

async function loadConfig() {
  loading.value = true
  try {
    const { data } = await http.get('/config')
    config.value = data.config || config.value
    mode.value = data.mode || 'unknown'
  } catch (e) {
    showMessage(t('config.loadFailed') + ': ' + (translateError(e.response?.data?.error, t) || e.message), true)
  } finally {
    loading.value = false
  }
}

function cleanConfig(obj) {
  const cleaned = {}
  for (const [key, value] of Object.entries(obj)) {
    if (value === null || value === undefined || value === '') continue
    if (typeof value === 'object' && !Array.isArray(value)) {
      const sub = cleanConfig(value)
      if (Object.keys(sub).length > 0) {
        cleaned[key] = sub
      }
    } else {
      cleaned[key] = value
    }
  }
  return cleaned
}

async function saveConfig() {
  saving.value = true
  try {
    const payload = cleanConfig(config.value)
    await http.put('/config', payload)
    showMessage(t('config.configSaved'))
  } catch (e) {
    showMessage(t('config.saveFailed') + ': ' + (translateError(e.response?.data?.error, t) || e.message), true)
  } finally {
    saving.value = false
  }
}

async function restartService() {
  restarting.value = true
  try {
    await http.post('/config/restart')
    showMessage(t('config.restartStarted'))
  } catch (e) {
    showMessage(t('config.restartFailed') + ': ' + (translateError(e.response?.data?.error, t) || e.message), true)
  } finally {
    restarting.value = false
  }
}

async function reloadEngine() {
  reloading.value = true
  try {
    const { data } = await http.post('/config/reload')
    mode.value = data.mode || mode.value
    showMessage(t('config.reloadSuccess', { mode: data.mode }))
  } catch (e) {
    showMessage(t('config.reloadFailed') + ': ' + (translateError(e.response?.data?.error, t) || e.message), true)
  } finally {
    reloading.value = false
  }
}

function logout() {
  auth.logout()
  router.push('/login')
}

onMounted(loadConfig)
</script>

<style scoped>
.config-sections {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.config-group {
  border: 1px solid var(--onyx-border);
  border-radius: 8px;
  overflow: hidden;
}

.config-group summary {
  padding: 0.6rem 1rem;
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  background: rgba(18, 90, 138, 0.04);
  user-select: none;
}

.config-group summary:hover {
  background: rgba(18, 90, 138, 0.08);
}

.config-fields {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 0.5rem 1rem;
  padding: 0.75rem 1rem;
  align-items: center;
}

.config-fields label {
  font-size: 0.82rem;
  font-family: 'SF Mono', 'Fira Code', monospace;
  color: var(--onyx-muted);
  white-space: nowrap;
}

.config-fields .form-control,
.config-fields .form-select {
  font-size: 0.85rem;
  padding: 0.35rem 0.6rem;
}

.mode-display {
  text-align: center;
}

.mode-steps {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.mode-step {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  color: var(--onyx-muted);
  background: rgba(0, 0, 0, 0.02);
}

.mode-step.active {
  color: var(--onyx-accent);
  background: rgba(30, 167, 161, 0.08);
  font-weight: 600;
}

.btn-accent {
  background: var(--onyx-accent);
  color: #fff;
  border: none;
}

.btn-accent:hover {
  background: #17908b;
  color: #fff;
}
</style>
