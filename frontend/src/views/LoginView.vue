<template>
  <div class="login-page">
    <div class="login-shell">
      <section class="login-hero">
        <div class="brand-mark mb-3">ONYX CONTROL</div>
        <h1 class="page-title mb-3">清爽、可靠、面向企业的存储控制平面</h1>
        <p class="text-secondary mb-4">
          从设备编排、volume 生命周期，到引擎状态、审计与权限管理，统一落在一个更适合团队协作的控制面里。
        </p>
        <div class="hero-points">
          <div class="hero-point">
            <strong>初始化即治理</strong>
            <span>首次启动必须创建管理员，避免默认口令直接进入系统。</span>
          </div>
          <div class="hero-point">
            <strong>控制面数据库独立</strong>
            <span>用户、角色与审计走 SQLite，和 Onyx 数据路径解耦。</span>
          </div>
          <div class="hero-point">
            <strong>企业风格主题</strong>
            <span>浅色、蓝绿主色、明确层次，更适合长期运维和团队协作。</span>
          </div>
        </div>
      </section>

      <section class="login-card">
        <div class="tiny-label mb-2">{{ initialized ? 'Sign In' : 'First-Time Setup' }}</div>
        <h2 class="login-title mb-2">{{ initialized ? '控制平面登录' : '初始化管理员账号' }}</h2>
        <p class="text-secondary mb-4">
          {{
            initialized
              ? '初始化完成后，从这里登录控制面。后续可以继续接 LDAP / OIDC。'
              : '当前系统还没有任何用户。请先创建首个管理员账号，初始化完成后再进入 Dashboard。'
          }}
        </p>

        <form v-if="!initialized" class="d-grid gap-3" @submit.prevent="initialize">
          <div>
            <label class="form-label">管理员用户名</label>
            <input v-model="setupForm.username" class="form-control form-control-lg" />
          </div>
          <div>
            <label class="form-label">显示名称</label>
            <input v-model="setupForm.displayName" class="form-control form-control-lg" />
          </div>
          <div>
            <label class="form-label">初始密码</label>
            <input v-model="setupForm.password" type="password" class="form-control form-control-lg" />
          </div>
          <div v-if="error" class="alert alert-danger py-2 mb-0">{{ error }}</div>
          <button class="btn btn-accent btn-lg" :disabled="submitting">
            {{ submitting ? '初始化中...' : '完成初始化' }}
          </button>
        </form>

        <form v-else class="d-grid gap-3" @submit.prevent="submit">
          <div>
            <label class="form-label">用户名</label>
            <input v-model="form.username" class="form-control form-control-lg" />
          </div>
          <div>
            <label class="form-label">密码</label>
            <input v-model="form.password" type="password" class="form-control form-control-lg" />
          </div>
          <div v-if="error" class="alert alert-danger py-2 mb-0">{{ error }}</div>
          <button class="btn btn-accent btn-lg" :disabled="auth.loading">
            {{ auth.loading ? '登录中...' : '进入 Dashboard' }}
          </button>
        </form>
      </section>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const error = ref('')
const submitting = ref(false)
const initialized = ref(true)
const form = reactive({
  username: '',
  password: '',
})
const setupForm = reactive({
  username: 'admin',
  displayName: 'Onyx Administrator',
  password: '',
})

const submit = async () => {
  error.value = ''
  try {
    await auth.login(form)
    await auth.fetchMe()
    router.push('/overview')
  } catch (err) {
    if (err?.response?.data?.code === 'setup_required') {
      initialized.value = false
      error.value = '系统尚未初始化，请先创建管理员账号。'
      return
    }
    error.value = err?.response?.data?.error || err.message || '登录失败'
  }
}

const loadSetupStatus = async () => {
  const { data } = await http.get('/setup/status')
  initialized.value = !!data.initialized
  setupForm.username = data.suggestedUsername || 'admin'
}

const initialize = async () => {
  error.value = ''
  submitting.value = true
  try {
    await http.post('/setup/initialize', setupForm)
    initialized.value = true
    form.username = setupForm.username
    form.password = setupForm.password
    await submit()
  } catch (err) {
    error.value = err?.response?.data?.error || err.message || '初始化失败'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await loadSetupStatus()
})
</script>
