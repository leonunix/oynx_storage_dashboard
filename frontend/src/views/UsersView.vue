<template>
  <AppShell title="用户与权限" eyebrow="RBAC Administration" :user="auth.user" @logout="logout">
    <div class="row g-4">
      <div class="col-12 col-xl-5">
        <div class="content-card">
          <div class="section-header">
            <h3>创建用户</h3>
            <span class="badge text-bg-info">SQLite</span>
          </div>

          <form class="d-grid gap-3" @submit.prevent="createUser">
            <input v-model="createForm.username" class="form-control" placeholder="username" />
            <input v-model="createForm.displayName" class="form-control" placeholder="display name" />
            <select v-model="createForm.role" class="form-select">
              <option v-for="role in roles" :key="role.name" :value="role.name">{{ role.name }}</option>
            </select>
            <input v-model="createForm.password" type="password" class="form-control" placeholder="initial password" />
            <button class="btn btn-accent">创建用户</button>
          </form>
        </div>

        <div class="content-card mt-4">
          <div class="section-header">
            <h3>角色权限</h3>
            <button class="btn btn-sm btn-outline-light" @click="loadRoles">刷新</button>
          </div>
          <div class="timeline">
            <div v-for="role in roles" :key="role.name" class="timeline-item">
              <div class="timeline-time text-uppercase">{{ role.name }}</div>
              <div class="d-flex flex-wrap gap-2">
                <span v-for="permission in role.permissions" :key="permission" class="badge rounded-pill text-bg-dark">
                  {{ permission }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-xl-7">
        <div class="content-card">
          <div class="section-header">
            <h3>用户列表</h3>
            <button class="btn btn-sm btn-outline-light" @click="loadUsers">刷新</button>
          </div>

          <div class="table-responsive">
            <table class="table align-middle">
              <thead>
                <tr>
                  <th>Username</th>
                  <th>Display</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in users" :key="user.username">
                  <td>{{ user.username }}</td>
                  <td>
                    <input
                      v-model="editState[user.username].displayName"
                      class="form-control form-control-sm"
                      :disabled="savingUser === user.username"
                    />
                  </td>
                  <td>
                    <select
                      v-model="editState[user.username].role"
                      class="form-select form-select-sm"
                      :disabled="savingUser === user.username"
                    >
                      <option v-for="role in roles" :key="role.name" :value="role.name">{{ role.name }}</option>
                    </select>
                  </td>
                  <td>
                    <div class="form-check form-switch">
                      <input
                        :id="`disabled-${user.username}`"
                        v-model="editState[user.username].disabled"
                        class="form-check-input"
                        type="checkbox"
                        :disabled="savingUser === user.username"
                      />
                      <label class="form-check-label" :for="`disabled-${user.username}`">
                        {{ editState[user.username].disabled ? 'disabled' : 'active' }}
                      </label>
                    </div>
                  </td>
                  <td class="text-end">
                    <div class="d-flex gap-2 justify-content-end">
                      <button class="btn btn-sm btn-outline-light" @click="saveUser(user.username)">保存</button>
                      <button class="btn btn-sm btn-outline-warning" @click="resetPassword(user.username)">重置密码</button>
                    </div>
                  </td>
                </tr>
                <tr v-if="users.length === 0">
                  <td colspan="5" class="text-center text-secondary py-4">暂无用户</td>
                </tr>
              </tbody>
            </table>
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
const users = ref([])
const roles = ref([])
const savingUser = ref('')
const editState = reactive({})

const createForm = reactive({
  username: '',
  displayName: '',
  role: 'viewer',
  password: '',
})

const ensureEditState = () => {
  for (const user of users.value) {
    editState[user.username] = {
      displayName: user.displayName,
      role: user.role,
      disabled: !!user.disabled,
    }
  }
}

const loadUsers = async () => {
  const { data } = await http.get('/users')
  users.value = data.items || []
  ensureEditState()
}

const loadRoles = async () => {
  const { data } = await http.get('/roles')
  roles.value = data.items || []
  if (!createForm.role && roles.value.length > 0) {
    createForm.role = roles.value[0].name
  }
}

const createUser = async () => {
  await http.post('/users', createForm)
  createForm.username = ''
  createForm.displayName = ''
  createForm.password = ''
  createForm.role = roles.value[0]?.name || 'viewer'
  await loadUsers()
}

const saveUser = async (username) => {
  savingUser.value = username
  try {
    await http.patch(`/users/${username}`, editState[username])
    await loadUsers()
  } finally {
    savingUser.value = ''
  }
}

const resetPassword = async (username) => {
  const password = window.prompt(`为 ${username} 设置新密码`)
  if (!password) {
    return
  }
  await http.post(`/users/${username}/reset-password`, { password })
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

onMounted(async () => {
  await auth.fetchMe()
  await Promise.all([loadRoles(), loadUsers()])
})
</script>
