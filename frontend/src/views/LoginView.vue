<template>
  <div class="login-page">
    <div class="login-shell">
      <section class="login-hero">
        <div class="brand-mark mb-3">ONYX CONTROL</div>
        <h1 class="page-title mb-3">{{ $t('login.heroTitle') }}</h1>
        <p class="text-secondary mb-4">
          {{ $t('login.heroDesc') }}
        </p>
        <div class="hero-points">
          <div class="hero-point">
            <strong>{{ $t('login.initGovernance') }}</strong>
            <span>{{ $t('login.initGovernanceDesc') }}</span>
          </div>
          <div class="hero-point">
            <strong>{{ $t('login.independentDB') }}</strong>
            <span>{{ $t('login.independentDBDesc') }}</span>
          </div>
          <div class="hero-point">
            <strong>{{ $t('login.enterpriseTheme') }}</strong>
            <span>{{ $t('login.enterpriseThemeDesc') }}</span>
          </div>
        </div>
      </section>

      <section class="login-card">
        <div class="tiny-label mb-2">{{ initialized ? $t('login.signIn') : $t('login.firstTimeSetup') }}</div>
        <h2 class="login-title mb-2">{{ initialized ? $t('login.loginTitle') : $t('login.initTitle') }}</h2>
        <p class="text-secondary mb-4">
          {{ initialized ? $t('login.loginDesc') : $t('login.initDesc') }}
        </p>

        <form v-if="!initialized" class="d-grid gap-3" @submit.prevent="initialize">
          <div>
            <label class="form-label">{{ $t('login.adminUsername') }}</label>
            <input v-model="setupForm.username" class="form-control form-control-lg" />
          </div>
          <div>
            <label class="form-label">{{ $t('login.displayName') }}</label>
            <input v-model="setupForm.displayName" class="form-control form-control-lg" />
          </div>
          <div>
            <label class="form-label">{{ $t('login.initialPassword') }}</label>
            <input v-model="setupForm.password" type="password" class="form-control form-control-lg" />
          </div>
          <div v-if="error" class="alert alert-danger py-2 mb-0">{{ error }}</div>
          <button class="btn btn-accent btn-lg" :disabled="submitting">
            {{ submitting ? $t('login.initializing') : $t('login.completeInit') }}
          </button>
        </form>

        <form v-else class="d-grid gap-3" @submit.prevent="submit">
          <div>
            <label class="form-label">{{ $t('login.username') }}</label>
            <input v-model="form.username" class="form-control form-control-lg" />
          </div>
          <div>
            <label class="form-label">{{ $t('login.password') }}</label>
            <input v-model="form.password" type="password" class="form-control form-control-lg" />
          </div>
          <div v-if="error" class="alert alert-danger py-2 mb-0">{{ error }}</div>
          <button class="btn btn-accent btn-lg" :disabled="auth.loading">
            {{ auth.loading ? $t('login.loggingIn') : $t('login.enterDashboard') }}
          </button>
        </form>
      </section>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import http from '../api/http'
import { useAuthStore } from '../stores/auth'
import { translateError } from '../i18n/errorMap'

const { t } = useI18n()
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
      error.value = t('login.setupRequired')
      return
    }
    error.value = translateError(err?.response?.data?.error, t) || t('login.loginFailed')
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
    error.value = translateError(err?.response?.data?.error, t) || t('login.initFailed')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await loadSetupStatus()
})
</script>
