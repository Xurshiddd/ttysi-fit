// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // SSR yoqiq (default) — Nuxt 3.21 + Vite 7 da ssr:false build xatosi beradi.
  // Auth cookie-asosli, shuning uchun SSR bilan ham to'g'ri ishlaydi.
  compatibilityDate: '2025-01-01',
  devtools: { enabled: false },

  modules: ['@nuxt/ui', '@nuxtjs/i18n'],

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1'
    }
  },

  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'uz',
    locales: [
      { code: 'uz', name: "O‘zbek" },
      { code: 'ru', name: 'Русский' },
      { code: 'en', name: 'English' }
    ],
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'tf_lang',
      redirectOn: 'root'
    },
    vueI18n: './i18n.config.ts'
  },

  app: {
    head: {
      title: 'TTYSI_FIT Admin',
      htmlAttrs: { lang: 'uz' },
      