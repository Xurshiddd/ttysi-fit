interface UserInfo {
  id: string
  full_name: string
  email: string
  role: string
  language?: string
}

const TOKEN_KEY = 'tf_token'
const REFRESH_KEY = 'tf_refresh'
const USER_KEY = 'tf_user'

export const useAuth = () => {
  const config = useRuntimeConfig()
  const token = useCookie<string | null>(TOKEN_KEY, { maxAge: 60 * 60 * 24 * 7, sameSite: 'lax' })
  const refresh = useCookie<string | null>(REFRESH_KEY, { maxAge: 60 * 60 * 24 * 7, sameSite: 'lax' })
  const user = useCookie<UserInfo | null>(USER_KEY, { maxAge: 60 * 60 * 24 * 7, sameSite: 'lax' })

  async function login(email: string, password: string) {
    const res = await $fetch<{ data: { access_token: string; refresh_token: string; user: UserInfo } }>(
      config.public.apiBase + '/auth/login',
      { method: 'POST', body: { email, password } }
    )
    token.value = res.data.access_token
    refresh.value = res.data.refresh_token
    user.value = res.data.user
    return res
  }

  function logout() {
    token.value = null
    refresh.value = null
    user.value = null
    return navigateTo('/login')
  }

  const isAuthenticated = computed(() => !!token.value)

  return { token, refresh, user, login, logout, isAuthenticated }
}
