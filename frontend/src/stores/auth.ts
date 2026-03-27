import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const storedUser = localStorage.getItem('optiyoo_user')
  const initialUser = storedUser ? JSON.parse(storedUser) : null

  const user = ref<{ id: string, name: string, email: string, username?: string, can_create_multi_question_surveys?: boolean } | null>(initialUser)
  const isAuthenticated = ref(!!initialUser)

  const setUser = (userData: any) => {
    user.value = userData
    isAuthenticated.value = true
    localStorage.setItem('optiyoo_user', JSON.stringify(userData))
  }

  const logout = () => {
    user.value = null
    isAuthenticated.value = false
    localStorage.removeItem('optiyoo_user')
  }

  watch(user, (newVal) => {
      if (newVal) {
          localStorage.setItem('optiyoo_user', JSON.stringify(newVal))
      }
  }, { deep: true })

  return { user, isAuthenticated, setUser, logout }
})
