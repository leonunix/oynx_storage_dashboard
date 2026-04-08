import { defineStore } from 'pinia'
import http from '../api/http'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: localStorage.getItem('onyx_dashboard_token') || '',
    loading: false,
  }),
  actions: {
    async login(credentials) {
      this.loading = true
      try {
        const { data } = await http.post('/auth/login', credentials)
        this.user = data.user
        this.token = data.token
        localStorage.setItem('onyx_dashboard_token', data.token)
        return true
      } finally {
        this.loading = false
      }
    },
    async fetchMe() {
      if (!this.token) {
        return null
      }
      const { data } = await http.get('/auth/me')
      this.user = data
      return data
    },
    logout() {
      this.user = null
      this.token = ''
      localStorage.removeItem('onyx_dashboard_token')
    },
  },
})
