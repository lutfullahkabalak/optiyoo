import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const TOKEN_KEY = 'optiyoo_token'

export const useAuthStore = defineStore('auth', () => {
  const storedUserRaw = localStorage.getItem('optiyoo_user')
  const storedToken = localStorage.getItem(TOKEN_KEY) || ''

  let initialUser: {
    id: string
    name: string
    email: string
    username?: string
    avatar_url?: string
    avatar_color?: string
    can_create_multi_question_surveys?: boolean
  } | null = null
  if (storedUserRaw) {
    try {
      initialUser = JSON.parse(storedUserRaw)
    } catch {
      initialUser = null
    }
  }
  if (initialUser && !storedToken) {
    initialUser = null
    localStorage.removeItem('optiyoo_user')
  }

  const user = ref<{
    id: string
    name: string
    email: string
    username?: string
    avatar_url?: string
    avatar_color?: string
    can_create_multi_question_surveys?: boolean
  } | null>(initialUser)

  const token = ref(storedToken)
  const isAuthenticated = ref(!!initialUser && !!storedToken)

  const authHeadersGet = (): HeadersInit =>
    token.value ? { Authorization: `Bearer ${token.value}` } : {}

  const authHeadersJson = (): HeadersInit => ({
    'Content-Type': 'application/json',
    ...(token.value ? { Authorization: `Bearer ${token.value}` } : {})
  })

  const authHeadersMultipart = (): HeadersInit =>
    token.value ? { Authorization: `Bearer ${token.value}` } : {}

  const setUser = (
    userData: {
      id: string
      name: string
      email: string
      username?: string
      avatar_url?: string
      avatar_color?: string
      can_create_multi_question_surveys?: boolean
    },
    accessToken?: string
  ) => {
    user.value = userData
    isAuthenticated.value = true
    localStorage.setItem('optiyoo_user', JSON.stringify(userData))
    if (accessToken !== undefined && accessToken !== '') {
      token.value = accessToken
      localStorage.setItem(TOKEN_KEY, accessToken)
    }
  }

  const logout = () => {
    user.value = null
    isAuthenticated.value = false
    token.value = ''
    localStorage.removeItem('optiyoo_user')
    localStorage.removeItem(TOKEN_KEY)
  }

  watch(user, (newVal) => {
    if (newVal) {
      localStorage.setItem('optiyoo_user', JSON.stringify(newVal))
    }
  }, { deep: true })

  watch(token, (newVal) => {
    if (newVal) {
      localStorage.setItem(TOKEN_KEY, newVal)
    }
  })

  return {
    user,
    token,
    isAuthenticated,
    setUser,
    logout,
    authHeadersGet,
    authHeadersJson,
    authHeadersMultipart
  }
})
