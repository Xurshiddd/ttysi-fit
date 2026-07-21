interface UserInfo {
  id: string
  full_name: string
  email: string
  role: string
  language?: string
}

export const useAuth = () => {
  const config = useRuntimeConfig()
  // secure: prod build'da cookie faqat HTTPS orqali yuboriladi (§17.3 #20/#35).
  // Dev'da (http://localhost) o'chirilgan — aks holda cookie umuman saqlanmaydi.
  const opts = { maxAge: 60 * 60 * 24 * 7, sameSite: 'lax' as const, secure: !import.meta.dev }
  const token = useCookie<string | null>('tf_token', opts)
  const refresh = useCookie<string | null>('tf_refresh', opts)
  const user = useCookie<UserInfo | null>('tf_user', opts)

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

  // refreshTokens — refresh token bilan yangi access+refresh juftini oladi
  // (backend rotatsiya qiladi: har safar yangi refresh ham qaytadi).
  // Muvaffaqiyatda true, aks holda false qaytaradi (logout chaqirmaydi —
  // qaror chaqiruvchida, useApi da).
  async function refreshTokens(): Promise<boolean> {
    if (!refresh.value) return false
    try {
      const res = await $fetch<{ data: { access_token: string; refresh_token: string; user?: UserInfo } }>(
        config.public.apiBase + '/auth/refresh',
        { method: 'POST', body: { refresh_token: refresh.value } }
      )
      token.value = res.data.access_token
      refresh.value = res.data.refresh_token
      if (res.data.user) user.value = res.data.user
      return true
    } catch {
      return false
    }
  }

  function logout() {
    token.value = null
    refresh.value = null
    user.value = null
    return navigateTo('/login')
  }

  const isAuthenticated = computed(() => !!token.value)

  return { token, refresh, user, login, logout, refreshTokens, isAuthenticated }
}
