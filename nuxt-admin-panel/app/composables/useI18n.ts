import { messages, localeList } from '~/i18n/messages'

// Yengil i18n — modulsiz (uz/ru/en). Til cookie'da saqlanadi.
export const useI18n = () => {
  const cookie = useCookie<string>('tf_lang', { maxAge: 60 * 60 * 24 * 365, default: () => 'uz' })
  const locale = useState<string>('locale', () => cookie.value || 'uz')

  const locales = localeList

  function setLocale(code: string) {
    locale.value = code
    cookie.value = code
  }

  function t(key: string): string {
    const lang = messages[locale.value] || messages.uz
    return lang[key] ?? messages.uz[key] ?? key
  }

  return { t, locale, locales, setLocale }
}
