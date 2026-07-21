// useApi — Authorization va Accept-Language header bilan API chaqiruvchi.
// Access token (15 min) tugaganda (401) avtomatik refresh token bilan yangilanadi
// va so'rov bir marta qayta yuboriladi. Faqat refresh ham muvaffaqiyatsiz
// bo'lsa — logout (login sahifaga).

// Single-flight: bir vaqtda bir nechta so'rov 401 olsa, faqat BITTA refresh
// bajariladi; qolganlari o'sha natijani kutadi. (Modul darajasida — SPA singleton.)
let refreshInFlight: Promise<boolean> | null = null

export const useApi = () => {
  const config = useRuntimeConfig()
  const { token, logout, refreshTokens } = useAuth()
  const { locale } = useI18n()

  function sharedRefresh(): Promise<boolean> {
    if (!refreshInFlight) {
      refreshInFlight = refreshTokens().finally(() => { refreshInFlight = null })
    }
    return refreshInFlight
  }

  async function api<T>(path: string, opts: Record<string, any> = {}, retried = false): Promise<T> {
    try {
      return await $fetch<T>(config.public.apiBase + path, {
        ...opts,
        headers: {
          ...(opts.headers || {}),
          ...(token.value ? { Authorization: `Bearer ${token.value}` } : {}),
          'Accept-Language': locale.value
        }
      })
    } catch (e: any) {
      const status = e?.response?.status || e?.status

      // 401 — token eskirgan bo'lishi mumkin. Auth endpoint'lari uchun refresh qilmaymiz.
      const isAuthCall = path.startsWith('/auth/')
      if (status === 401 && !retried && !isAuthCall) {
        const ok = await sharedRefresh()
        if (ok) {
          // Yangi token bilan bir marta qayta urinamiz.
          return api<T>(path, opts, true)
        }
        // Refresh ham ishlamadi — sessiya tugadi.
        await logout()
      }
      throw e
    }
  }

  return { api }
}
