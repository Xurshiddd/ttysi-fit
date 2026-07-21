// useApi — Authorization va Accept-Language header bilan API chaqiruvchi.
export const useApi = () => {
  const config = useRuntimeConfig()
  const { token, logout } = useAuth()
  const { locale } = useI18n()

  async function api<T>(path: string, opts: Record<string, any> = {}): Promise<T> {
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
      // Token muddati o'tgan / yaroqsiz — login sahifasiga
      if (e?.response?.status === 401) {
        await logout()
      }
      throw e
    }
  }

  return { api }
}
