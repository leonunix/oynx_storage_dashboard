<template>
  <AppShell title="存储编排" eyebrow="dm / LVM / Provisioning" :user="auth.user" @logout="logout">
    <div class="row g-4">
      <div class="col-12 col-xl-7">
        <div class="content-card">
          <div class="section-header">
            <h3>节点存储拓扑</h3>
            <button class="btn btn-sm btn-outline-light" @click="loadLayout">刷新</button>
          </div>

          <h4 class="sub-title">Block Devices</h4>
          <div class="chip-grid mb-4">
            <div v-for="device in layout.blockDevices || []" :key="device.name" class="chip-card">
              <strong>{{ device.name }}</strong>
              <span>{{ device.type }} / {{ device.size }}</span>
              <span>{{ device.state || 'unknown' }}</span>
            </div>
          </div>

          <h4 class="sub-title">dm Targets</h4>
          <div class="list-group list-group-flush">
            <div v-for="item in layout.dmTargets || []" :key="item.name" class="list-group-item bg-transparent">
              {{ item.name }}
            </div>
          </div>

          <h4 class="sub-title mt-4">LVM</h4>
          <div class="list-group list-group-flush">
            <div v-for="item in layout.logicalVolumes || []" :key="`${item.vgName}-${item.name}`" class="list-group-item bg-transparent">
              {{ item.vgName }}/{{ item.name }} · {{ item.size }} · {{ item.attr }}
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-xl-5">
        <div class="content-card">
          <div class="section-header">
            <h3>Provision Plan</h3>
            <span class="badge text-bg-warning">Preview Only</span>
          </div>

          <form class="d-grid gap-3" @submit.prevent="previewPlan">
            <input v-model="provision.name" class="form-control" placeholder="storage name" />
            <input v-model="deviceInput" class="form-control" placeholder="/dev/nvme0n1,/dev/nvme1n1" />
            <select v-model="provision.raidType" class="form-select">
              <option value="raid5">raid5</option>
              <option value="raid6">raid6</option>
              <option value="mirror">mirror</option>
              <option value="none">none</option>
            </select>
            <input v-model.number="provision.stripSizeKb" type="number" class="form-control" placeholder="strip size KB" />
            <input v-model="provision.vgName" class="form-control" placeholder="vg name" />
            <input v-model="provision.metaLvName" class="form-control" placeholder="meta lv name" />
            <input v-model="provision.dataLvName" class="form-control" placeholder="data lv name" />
            <button class="btn btn-accent">生成计划</button>
          </form>

          <div v-if="plan.commands?.length" class="mt-4">
            <h4 class="sub-title">执行计划</h4>
            <ol class="ps-3">
              <li v-for="command in plan.commands" :key="command" class="mb-2">
                <code>{{ command }}</code>
              </li>
            </ol>
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
const layout = reactive({})
const plan = reactive({})
const deviceInput = ref('/dev/nvme0n1,/dev/nvme1n1,/dev/nvme2n1')
const provision = reactive({
  name: 'onyx-pool-a',
  raidType: 'raid5',
  stripSizeKb: 64,
  vgName: 'vg0',
  metaLvName: 'onyx-meta',
  dataLvName: 'onyx-data',
})

const loadLayout = async () => {
  const { data } = await http.get('/storage/layout')
  Object.assign(layout, data)
}

const previewPlan = async () => {
  const payload = {
    ...provision,
    devices: deviceInput.value.split(',').map((item) => item.trim()).filter(Boolean),
  }
  const { data } = await http.post('/storage/workflows/provision/preview', payload)
  Object.assign(plan, data)
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
