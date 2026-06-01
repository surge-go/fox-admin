import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('fox-admin-token') ?? '')
  const username = ref(localStorage.getItem('fox-admin-username') ?? 'admin')

  const isLoggedIn = computed(() => token.value.length > 0)

  function setSession(nextToken: string, nextUsername = 'admin') {
    token.value = nextToken
    username.value = nextUsername
    localStorage.setItem('fox-admin-token', nextToken)
    localStorage.setItem('fox-admin-username', nextUsername)
  }

  function clearSession() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('fox-admin-token')
    localStorage.removeItem('fox-admin-username')
  }

  return {
    token,
    username,
    isLoggedIn,
    setSession,
    clearSession,
  }
})
