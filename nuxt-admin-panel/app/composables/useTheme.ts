// Yengil light/dark mavzu boshqaruvi. Tanlov cookie'da saqlanadi.
export const useTheme = () => {
  const cookie = useCookie<string>('tf_theme', {
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
    default: () => 'light'
  })
  const theme = useState<string>('theme', () => cookie.value || 'light')

  const isDark = computed(() => theme.value === 'dark')

  function setTheme(value: 'light' | 'dark') {
    theme.value = value
    cookie.value = value
  }

  function toggle() {
    setTheme(isDark.value ? 'light' : 'dark')
  }

  // <html> ga class qo'shish (SSR + client). app.vue da useHead orqali ham bog'lanadi.
  if (import.meta.client) {
    watch(
      theme,
      (v) => document.documentElement.classList.toggle('dark', v === 'dark'),
      { immediate: true }
    )
  }

  return { theme, isDark, setTheme, toggle }
}
