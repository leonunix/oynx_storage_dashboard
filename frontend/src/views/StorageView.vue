<template>
  <AppShell title="storage.title" eyebrow="dm / RAID / LVM / Provisioning" :user="auth.user" @logout="logout">
    <MutationGuard :allowed="layout.allowMutations" />

    <div v-if="error" class="alert alert-danger mb-3">{{ error }}</div>
    <div v-if="success" class="alert alert-success mb-3">{{ success }}</div>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-4">
      <li class="nav-item" v-for="tab in tabs" :key="tab.key">
        <button
          class="nav-link"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          <i :class="tab.icon" class="me-1"></i>{{ tab.label }}
        </button>
      </li>
    </ul>

    <!-- Tab 1: Topology -->
    <div v-if="activeTab === 'topology'">
      <div class="d-flex justify-content-end mb-3">
        <button class="btn btn-sm btn-outline-light" @click="loadLayout">{{ $t('common.refresh') }}</button>
      </div>

      <div class="row g-4">
        <div class="col-12 col-xl-6">
          <div class="content-card">
            <h3 class="sub-title">Block Devices</h3>
            <div class="chip-grid">
              <div v-for="d in layout.blockDevices || []" :key="d.name" class="chip-card">
                <strong>{{ d.name }}</strong>
                <span>{{ d.type }} / {{ d.size }}</span>
                <span>{{ d.state || 'unknown' }}</span>
              </div>
            </div>
            <div v-if="!(layout.blockDevices?.length)" class="text-muted">{{ $t('storage.noBlockDevices') }}</div>
          </div>
        </div>

        <div class="col-12 col-xl-6">
          <div class="content-card">
            <h3 class="sub-title">dm Targets</h3>
            <div class="list-group list-group-flush">
              <div v-for="item in layout.dmTargets || []" :key="item.name" class="list-group-item bg-transparent">
                {{ item.name }}
              </div>
            </div>
            <div v-if="!(layout.dmTargets?.length)" class="text-muted">{{ $t('storage.noDmTargets') }}</div>
          </div>
        </div>

        <div class="col-12 col-xl-6">
          <div class="content-card">
            <h3 class="sub-title">RAID Arrays</h3>
            <table v-if="layout.raidArrays?.length" class="table table-sm mb-0">
              <thead><tr><th>Name</th><th>Level</th><th>State</th><th>Size</th><th>Devices</th></tr></thead>
              <tbody>
                <tr v-for="r in layout.raidArrays" :key="r.name">
                  <td><code>{{ r.name }}</code></td>
                  <td>{{ r.level }}</td>
                  <td><span :class="raidStateBadge(r.state)" class="badge">{{ r.state }}</span></td>
                  <td>{{ r.size }}</td>
                  <td>{{ r.activeDevs }}/{{ r.totalDevs }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="text-muted">{{ $t('storage.noRaidArrays') }}</div>
          </div>
        </div>

        <div class="col-12 col-xl-6">
          <div class="content-card">
            <h3 class="sub-title">Volume Groups</h3>
            <table v-if="layout.volumeGroups?.length" class="table table-sm mb-0">
              <thead><tr><th>Name</th><th>Size</th><th>Free</th><th>PVs</th><th>LVs</th></tr></thead>
              <tbody>
                <tr v-for="vg in layout.volumeGroups" :key="vg.name">
                  <td><strong>{{ vg.name }}</strong></td>
                  <td>{{ vg.size }}</td>
                  <td>{{ vg.free }}</td>
                  <td>{{ vg.pvCount }}</td>
                  <td>{{ vg.lvCount }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="text-muted">{{ $t('storage.noVolumeGroups') }}</div>
          </div>
        </div>

        <div class="col-12 col-xl-6">
          <div class="content-card">
            <h3 class="sub-title">Physical Volumes</h3>
            <table v-if="layout.physicalVolumes?.length" class="table table-sm mb-0">
              <thead><tr><th>Name</th><th>VG</th><th>Size</th><th>Free</th></tr></thead>
              <tbody>
                <tr v-for="pv in layout.physicalVolumes" :key="pv.name">
                  <td><code>{{ pv.name }}</code></td>
                  <td>{{ pv.vgName || '(orphan)' }}</td>
                  <td>{{ pv.size }}</td>
                  <td>{{ pv.free }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="text-muted">{{ $t('storage.noPhysicalVolumes') }}</div>
          </div>
        </div>

        <div class="col-12 col-xl-6">
          <div class="content-card">
            <h3 class="sub-title">Logical Volumes</h3>
            <table v-if="layout.logicalVolumes?.length" class="table table-sm mb-0">
              <thead><tr><th>VG</th><th>Name</th><th>Size</th><th>Attr</th></tr></thead>
              <tbody>
                <tr v-for="lv in layout.logicalVolumes" :key="`${lv.vgName}-${lv.name}`">
                  <td>{{ lv.vgName }}</td>
                  <td><strong>{{ lv.name }}</strong></td>
                  <td>{{ lv.size }}</td>
                  <td><code>{{ lv.attr }}</code></td>
                </tr>
              </tbody>
            </table>
            <div v-else class="text-muted">{{ $t('storage.noLogicalVolumes') }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 2: RAID -->
    <div v-if="activeTab === 'raid'">
      <div class="row g-4">
        <div class="col-12 col-xl-7">
          <div class="content-card">
            <div class="section-header">
              <h3>{{ $t('storage.existingRaid') }}</h3>
              <button class="btn btn-sm btn-outline-light" @click="loadLayout">{{ $t('common.refresh') }}</button>
            </div>

            <table v-if="layout.raidArrays?.length" class="table table-sm">
              <thead>
                <tr><th>Name</th><th>Level</th><th>State</th><th>Size</th><th>Devices</th><th>Actions</th></tr>
              </thead>
              <tbody>
                <tr v-for="r in layout.raidArrays" :key="r.name">
                  <td><code>{{ r.name }}</code></td>
                  <td>{{ r.level }}</td>
                  <td><span :class="raidStateBadge(r.state)" class="badge">{{ r.state }}</span></td>
                  <td>{{ r.size }}</td>
                  <td>
                    <span v-for="d in r.devices" :key="d" class="badge text-bg-light me-1">{{ d }}</span>
                    <span v-if="!r.devices?.length">{{ r.activeDevs }}/{{ r.totalDevs }}</span>
                  </td>
                  <td>
                    <ConfirmAction
                      :label="$t('common.stop')"
                      :confirm-text="$t('storage.stopRaid')"
                      :disabled="!layout.allowMutations"
                      @confirm="({ resolve, reject }) => stopRaid(r.name, resolve, reject)"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-else class="text-muted">{{ $t('storage.noRaidArrays') }}</div>
          </div>
        </div>

        <div class="col-12 col-xl-5">
          <div class="content-card">
            <h3 class="sub-title">{{ $t('storage.createRaid') }}</h3>
            <form class="d-grid gap-3" @submit.prevent="createRaid">
              <input v-model="raidForm.name" class="form-control" placeholder="/dev/md0" />
              <select v-model="raidForm.level" class="form-select">
                <option value="raid0">raid0</option>
                <option value="raid1">raid1</option>
                <option value="raid5">raid5</option>
                <option value="raid6">raid6</option>
                <option value="raid10">raid10</option>
              </select>
              <input v-model="raidDeviceInput" class="form-control" placeholder="/dev/sdb,/dev/sdc,/dev/sdd" />
              <input v-model.number="raidForm.chunkKb" type="number" class="form-control" placeholder="chunk size KB (default: 512)" />
              <label class="form-check">
                <input v-model="raidForm.force" type="checkbox" class="form-check-input" />
                <span class="form-check-label">{{ $t('common.force') }} (--run --force)</span>
              </label>
              <button class="btn btn-accent" :disabled="!layout.allowMutations || loading">
                {{ loading ? $t('common.creating') : $t('storage.createRaidBtn') }}
              </button>
            </form>
            <p class="text-muted mt-2" style="font-size: 0.8rem">
              {{ $t('storage.raidNote') }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 3: LVM -->
    <div v-if="activeTab === 'lvm'">
      <div class="d-flex justify-content-end mb-3">
        <button class="btn btn-sm btn-outline-light" @click="loadLayout">{{ $t('common.refresh') }}</button>
      </div>

      <!-- PV Section -->
      <details class="config-group mb-3" open>
        <summary>Physical Volumes (PV)</summary>
        <div class="config-body">
          <table v-if="layout.physicalVolumes?.length" class="table table-sm mb-3">
            <thead><tr><th>Name</th><th>VG</th><th>Size</th><th>Free</th><th>Actions</th></tr></thead>
            <tbody>
              <tr v-for="pv in layout.physicalVolumes" :key="pv.name">
                <td><code>{{ pv.name }}</code></td>
                <td>{{ pv.vgName || '(orphan)' }}</td>
                <td>{{ pv.size }}</td>
                <td>{{ pv.free }}</td>
                <td>
                  <ConfirmAction
                    :label="$t('common.remove')"
                    :confirm-text="$t('storage.removePVConfirm')"
                    :disabled="!layout.allowMutations || !!pv.vgName"
                    @confirm="({ resolve, reject }) => removePV(pv.name, resolve, reject)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="text-muted mb-3">{{ $t('storage.noPhysicalVolumes') }}</div>

          <form class="d-flex gap-2 align-items-end" @submit.prevent="createPV">
            <input v-model="pvForm.device" class="form-control form-control-sm" placeholder="/dev/sdb" style="max-width: 300px" />
            <label class="form-check form-check-inline mb-0">
              <input v-model="pvForm.force" type="checkbox" class="form-check-input" />
              <span class="form-check-label">{{ $t('common.force') }}</span>
            </label>
            <button class="btn btn-sm btn-accent" :disabled="!layout.allowMutations || loading">{{ $t('storage.createPV') }}</button>
          </form>
        </div>
      </details>

      <!-- VG Section -->
      <details class="config-group mb-3" open>
        <summary>Volume Groups (VG)</summary>
        <div class="config-body">
          <table v-if="layout.volumeGroups?.length" class="table table-sm mb-3">
            <thead><tr><th>Name</th><th>Size</th><th>Free</th><th>PVs</th><th>LVs</th><th>Actions</th></tr></thead>
            <tbody>
              <tr v-for="vg in layout.volumeGroups" :key="vg.name">
                <td><strong>{{ vg.name }}</strong></td>
                <td>{{ vg.size }}</td>
                <td>{{ vg.free }}</td>
                <td>{{ vg.pvCount }}</td>
                <td>{{ vg.lvCount }}</td>
                <td>
                  <ConfirmAction
                    :label="$t('common.remove')"
                    :confirm-text="$t('storage.removeVGConfirm')"
                    :disabled="!layout.allowMutations || vg.lvCount > 0"
                    @confirm="({ resolve, reject }) => removeVG(vg.name, resolve, reject)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="text-muted mb-3">{{ $t('storage.noVolumeGroups') }}</div>

          <form class="d-flex gap-2 align-items-end" @submit.prevent="createVG">
            <input v-model="vgForm.name" class="form-control form-control-sm" placeholder="vg0" style="max-width: 200px" />
            <input v-model="vgDeviceInput" class="form-control form-control-sm" placeholder="/dev/sdb,/dev/sdc" style="max-width: 300px" />
            <button class="btn btn-sm btn-accent" :disabled="!layout.allowMutations || loading">{{ $t('storage.createVG') }}</button>
          </form>
        </div>
      </details>

      <!-- LV Section -->
      <details class="config-group mb-3" open>
        <summary>Logical Volumes (LV)</summary>
        <div class="config-body">
          <table v-if="layout.logicalVolumes?.length" class="table table-sm mb-3">
            <thead><tr><th>VG</th><th>Name</th><th>Size</th><th>Attr</th><th>Actions</th></tr></thead>
            <tbody>
              <tr v-for="lv in layout.logicalVolumes" :key="`${lv.vgName}-${lv.name}`">
                <td>{{ lv.vgName }}</td>
                <td><strong>{{ lv.name }}</strong></td>
                <td>{{ lv.size }}</td>
                <td><code>{{ lv.attr }}</code></td>
                <td class="d-flex gap-2">
                  <ConfirmAction
                    :label="$t('common.remove')"
                    :confirm-text="$t('storage.removeLVConfirm')"
                    :disabled="!layout.allowMutations"
                    @confirm="({ resolve, reject }) => removeLV(lv.vgName, lv.name, resolve, reject)"
                  />
                  <button
                    class="btn btn-sm btn-outline-secondary"
                    :disabled="!layout.allowMutations"
                    @click="openResize(lv)"
                  >{{ $t('storage.resizeLabel') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="text-muted mb-3">{{ $t('storage.noLogicalVolumes') }}</div>

          <!-- Resize inline form -->
          <div v-if="resizeTarget" class="alert alert-info d-flex gap-2 align-items-center mb-3">
            <span>{{ $t('storage.resizeLabel') }} <strong>{{ resizeTarget.vgName }}/{{ resizeTarget.name }}</strong>:</span>
            <input v-model="resizeSize" class="form-control form-control-sm" placeholder="+10G or 500G" style="max-width: 200px" />
            <button class="btn btn-sm btn-accent" :disabled="loading" @click="resizeLV">{{ $t('common.ok') }}</button>
            <button class="btn btn-sm btn-outline-secondary" @click="resizeTarget = null">{{ $t('common.cancel') }}</button>
          </div>

          <!-- Create LV form -->
          <form class="d-flex gap-2 align-items-end" @submit.prevent="createLV">
            <input v-model="lvForm.name" class="form-control form-control-sm" placeholder="lv name" style="max-width: 200px" />
            <select v-model="lvForm.vgName" class="form-select form-select-sm" style="max-width: 200px">
              <option value="" disabled>select VG</option>
              <option v-for="vg in layout.volumeGroups || []" :key="vg.name" :value="vg.name">{{ vg.name }}</option>
            </select>
            <input v-model="lvForm.size" class="form-control form-control-sm" placeholder="100%FREE or 500G" style="max-width: 200px" />
            <button class="btn btn-sm btn-accent" :disabled="!layout.allowMutations || loading">{{ $t('storage.createLV') }}</button>
          </form>
        </div>
      </details>
    </div>

    <!-- Tab 4: Provision -->
    <div v-if="activeTab === 'provision'">
      <div class="row g-4">
        <div class="col-12 col-xl-6">
          <div class="content-card">
            <div class="section-header">
              <h3>Provision Plan</h3>
              <span v-if="!layout.allowMutations" class="badge text-bg-warning">{{ $t('storage.previewOnly') }}</span>
              <span v-else class="badge text-bg-success">{{ $t('storage.executionReady') }}</span>
            </div>

            <form class="d-grid gap-3" @submit.prevent="previewPlan">
              <input v-model="provision.name" class="form-control" placeholder="storage name" />
              <input v-model="deviceInput" class="form-control" placeholder="/dev/nvme0n1,/dev/nvme1n1" />
              <select v-model="provision.raidType" class="form-select">
                <option value="raid5">raid5</option>
                <option value="raid6">raid6</option>
                <option value="mirror">mirror</option>
                <option value="raid10">raid10</option>
                <option value="none">none</option>
              </select>
              <input v-model.number="provision.stripSizeKb" type="number" class="form-control" placeholder="strip size KB" />
              <input v-model="provision.vgName" class="form-control" placeholder="vg name" />
              <input v-model="provision.metaLvName" class="form-control" placeholder="meta lv name" />
              <input v-model="provision.dataLvName" class="form-control" placeholder="data lv name" />
              <button class="btn btn-accent">{{ $t('storage.generatePlan') }}</button>
            </form>
          </div>
        </div>

        <div class="col-12 col-xl-6">
          <div v-if="plan.commands?.length" class="content-card">
            <h3 class="sub-title">{{ $t('storage.executePlan') }}</h3>

            <div v-if="plan.safetyChecks?.length" class="mb-3">
              <strong>Safety Checks:</strong>
              <ul class="ps-3 mb-0">
                <li v-for="check in plan.safetyChecks" :key="check">{{ check }}</li>
              </ul>
            </div>

            <ol class="ps-3 mb-3">
              <li v-for="command in plan.commands" :key="command" class="mb-2">
                <code>{{ command }}</code>
              </li>
            </ol>

            <div v-if="plan.warnings?.length" class="mb-3">
              <div v-for="w in plan.warnings" :key="w" class="text-warning" style="font-size: 0.85rem">
                <i class="bi bi-exclamation-triangle me-1"></i>{{ w }}
              </div>
            </div>

            <div v-if="layout.allowMutations && plan.executionReady">
              <ConfirmAction
                :label="$t('storage.executePlan')"
                :confirm-text="$t('storage.executePlanConfirm')"
                button-class="btn btn-danger"
                @confirm="({ resolve, reject }) => executePlan(resolve, reject)"
              />
            </div>
          </div>

          <!-- Execution results -->
          <div v-if="execResults" class="content-card mt-4">
            <h3 class="sub-title">
              {{ $t('storage.executionResult') }}
              <span :class="execResults.success ? 'badge text-bg-success' : 'badge text-bg-danger'" class="ms-2">
                {{ execResults.success ? $t('storage.allSuccess') : $t('storage.partialFail') }}
              </span>
            </h3>
            <div v-for="(cr, idx) in execResults.results" :key="idx" class="mb-3 p-2 rounded" :class="cr.error ? 'bg-danger bg-opacity-10' : 'bg-success bg-opacity-10'">
              <div class="d-flex align-items-center gap-2 mb-1">
                <span :class="cr.error ? 'badge text-bg-danger' : 'badge text-bg-success'">{{ cr.error ? 'FAIL' : 'OK' }}</span>
                <code style="font-size: 0.85rem">{{ cr.command }}</code>
              </div>
              <pre v-if="cr.stdout" class="mb-0 ps-3" style="font-size: 0.8rem; white-space: pre-wrap">{{ cr.stdout }}</pre>
              <pre v-if="cr.error" class="mb-0 ps-3 text-danger" style="font-size: 0.8rem; white-space: pre-wrap">{{ cr.error }}</pre>
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
import ConfirmAction from '../components/ConfirmAction.vue'
import MutationGuard from '../components/MutationGuard.vue'
import { useAuthStore } from '../stores/auth'
import { translateError } from '../i18n/errorMap'

const LONG_TIMEOUT = { timeout: 120000 }

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const layout = reactive({})
const plan = reactive({})
const execResults = ref(null)
const activeTab = ref('topology')
const loading = ref(false)
const error = ref('')
const success = ref('')

const tabs = computed(() => [
  { key: 'topology', label: t('storage.tabs.topology'), icon: 'bi bi-diagram-3' },
  { key: 'raid', label: t('storage.tabs.raid'), icon: 'bi bi-hdd-rack' },
  { key: 'lvm', label: t('storage.tabs.lvm'), icon: 'bi bi-layers' },
  { key: 'provision', label: t('storage.tabs.provision'), icon: 'bi bi-tools' },
])

// ── Forms ──────────────────────────────────────────────────────────

const raidForm = reactive({ name: '/dev/md0', level: 'raid5', chunkKb: 512, force: false })
const raidDeviceInput = ref('/dev/sdb,/dev/sdc,/dev/sdd')
const pvForm = reactive({ device: '', force: false })
const vgForm = reactive({ name: '', })
const vgDeviceInput = ref('')
const lvForm = reactive({ name: '', vgName: '', size: '' })
const resizeTarget = ref(null)
const resizeSize = ref('')

const provision = reactive({
  name: 'onyx-pool-a',
  raidType: 'raid5',
  stripSizeKb: 64,
  vgName: 'vg0',
  metaLvName: 'onyx-meta',
  dataLvName: 'onyx-data',
})
const deviceInput = ref('/dev/nvme0n1,/dev/nvme1n1,/dev/nvme2n1')

// ── Helpers ────────────────────────────────────────────────────────

const clearMessages = () => { error.value = ''; success.value = '' }

const showError = (msg) => { error.value = msg; success.value = '' }
const showSuccess = (msg) => { success.value = msg; error.value = '' }

const raidStateBadge = (state) => {
  if (!state) return 'text-bg-secondary'
  const s = state.toLowerCase()
  if (s.includes('active') || s === 'clean') return 'text-bg-success'
  if (s.includes('degrad')) return 'text-bg-warning'
  return 'text-bg-danger'
}

const parseDevices = (input) =>
  input.split(',').map(s => s.trim()).filter(Boolean)

const openResize = (lv) => {
  resizeTarget.value = lv
  resizeSize.value = ''
}

// ── API calls ──────────────────────────────────────────────────────

const loadLayout = async () => {
  clearMessages()
  try {
    const { data } = await http.get('/storage/layout')
    Object.assign(layout, data)
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  }
}

// RAID
const createRaid = async () => {
  clearMessages()
  loading.value = true
  try {
    await http.post('/storage/raid', {
      name: raidForm.name,
      level: raidForm.level,
      devices: parseDevices(raidDeviceInput.value),
      chunkKb: raidForm.chunkKb || undefined,
      force: raidForm.force,
    }, LONG_TIMEOUT)
    showSuccess('RAID array created successfully')
    await loadLayout()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  } finally {
    loading.value = false
  }
}

const stopRaid = async (name, resolve, reject) => {
  clearMessages()
  try {
    const devName = name.startsWith('/dev/') ? name.replace('/dev/', '') : name
    await http.delete(`/storage/raid/${devName}`)
    showSuccess(`RAID ${name} stopped`)
    await loadLayout()
    resolve()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
    reject(err)
  }
}

// PV
const createPV = async () => {
  clearMessages()
  loading.value = true
  try {
    await http.post('/storage/pv', { device: pvForm.device, force: pvForm.force })
    showSuccess(`PV ${pvForm.device} created`)
    pvForm.device = ''
    pvForm.force = false
    await loadLayout()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  } finally {
    loading.value = false
  }
}

const removePV = async (device, resolve, reject) => {
  clearMessages()
  try {
    await http.delete('/storage/pv', { data: { device } })
    showSuccess(`PV ${device} removed`)
    await loadLayout()
    resolve()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
    reject(err)
  }
}

// VG
const createVG = async () => {
  clearMessages()
  loading.value = true
  try {
    await http.post('/storage/vg', {
      name: vgForm.name,
      devices: parseDevices(vgDeviceInput.value),
    })
    showSuccess(`VG ${vgForm.name} created`)
    vgForm.name = ''
    vgDeviceInput.value = ''
    await loadLayout()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  } finally {
    loading.value = false
  }
}

const removeVG = async (name, resolve, reject) => {
  clearMessages()
  try {
    await http.delete(`/storage/vg/${name}`)
    showSuccess(`VG ${name} removed`)
    await loadLayout()
    resolve()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
    reject(err)
  }
}

// LV
const createLV = async () => {
  clearMessages()
  loading.value = true
  try {
    await http.post('/storage/lv', {
      name: lvForm.name,
      vgName: lvForm.vgName,
      size: lvForm.size,
    })
    showSuccess(`LV ${lvForm.vgName}/${lvForm.name} created`)
    lvForm.name = ''
    lvForm.size = ''
    await loadLayout()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  } finally {
    loading.value = false
  }
}

const removeLV = async (vgName, name, resolve, reject) => {
  clearMessages()
  try {
    await http.delete('/storage/lv', { data: { name, vgName } })
    showSuccess(`LV ${vgName}/${name} removed`)
    await loadLayout()
    resolve()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
    reject(err)
  }
}

const resizeLV = async () => {
  if (!resizeTarget.value) return
  clearMessages()
  loading.value = true
  try {
    await http.post('/storage/lv/resize', {
      name: resizeTarget.value.name,
      vgName: resizeTarget.value.vgName,
      size: resizeSize.value,
    }, LONG_TIMEOUT)
    showSuccess(`LV ${resizeTarget.value.vgName}/${resizeTarget.value.name} resized`)
    resizeTarget.value = null
    await loadLayout()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  } finally {
    loading.value = false
  }
}

// Provision
const previewPlan = async () => {
  clearMessages()
  execResults.value = null
  const payload = {
    ...provision,
    devices: parseDevices(deviceInput.value),
  }
  try {
    const { data } = await http.post('/storage/workflows/provision/preview', payload)
    Object.assign(plan, data)
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
  }
}

const executePlan = async (resolve, reject) => {
  clearMessages()
  try {
    const { data } = await http.post('/storage/workflows/provision/execute', {
      commands: plan.commands,
    }, LONG_TIMEOUT)
    execResults.value = data
    if (data.success) {
      showSuccess(t('storage.provisionSuccess'))
    } else {
      showError(t('storage.provisionFail'))
    }
    await loadLayout()
    resolve()
  } catch (err) {
    showError(translateError(err.response?.data?.error, t) || err.message)
    reject(err)
  }
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

onMounted(async () => {
  await auth.fetchMe()
  await loadLayout()
})
</script>

<style scoped>
.nav-tabs .nav-link {
  color: var(--onyx-muted);
  border: none;
  border-bottom: 2px solid transparent;
  background: none;
  padding: 0.6rem 1.2rem;
  font-weight: 500;
  transition: all 0.2s;
}
.nav-tabs .nav-link:hover {
  color: var(--onyx-text);
  border-bottom-color: var(--onyx-border);
}
.nav-tabs .nav-link.active {
  color: var(--onyx-accent);
  border-bottom-color: var(--onyx-accent);
  background: none;
}
.config-group {
  border: 1px solid var(--onyx-border);
  border-radius: 0.75rem;
  overflow: hidden;
}
.config-group summary {
  padding: 0.75rem 1rem;
  font-weight: 600;
  cursor: pointer;
  background: var(--onyx-surface-soft);
}
.config-body {
  padding: 1rem;
}
</style>
