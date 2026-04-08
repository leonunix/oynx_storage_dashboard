<template>
  <div class="login-page">
    <div class="login-card">
      <div class="brand-mark mb-3">ONYX</div>
      <h1 class="page-title mb-2">控制平面登录</h1>
      <p class="text-secondary mb-4">
        建议后续接 LDAP/OIDC。当前脚手架默认使用 bootstrap 管理员账户。
      </p>

      <form class="d-grid gap-3" @submit.prevent="submit">
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
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const error = ref('')
const form = reactive({
  username: 'admin',
  password: 'onyx-admin',
})

const submit = async () => {
  error.value = ''
  try {
    await auth.login(form)
    await auth.fetchMe()
    router.push('/overview')
  } catch (err) {
    error.value = err?.response?.data?.error || err.message || '登录失败'
  }
}
</script>
