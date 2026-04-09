import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import OverviewView from '../views/OverviewView.vue'
import VolumesView from '../views/VolumesView.vue'
import StorageView from '../views/StorageView.vue'
import MetricsView from '../views/MetricsView.vue'
import AuditView from '../views/AuditView.vue'
import UsersView from '../views/UsersView.vue'

const routes = [
  { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
  { path: '/', redirect: '/overview' },
  { path: '/overview', name: 'overview', component: OverviewView },
  { path: '/storage', name: 'storage', component: StorageView },
  { path: '/volumes', name: 'volumes', component: VolumesView },
  { path: '/metrics', name: 'metrics', component: MetricsView },
  { path: '/audit', name: 'audit', component: AuditView },
  { path: '/users', name: 'users', component: UsersView },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const token = localStorage.getItem('onyx_dashboard_token')
  if (!to.meta.public && !token) {
    return { name: 'login' }
  }
  if (to.name === 'login' && token) {
    return { name: 'overview' }
  }
  return true
})

export default router
