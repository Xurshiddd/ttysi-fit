<script setup lang="ts">
definePageMeta({ layout: false })

const { login } = useAuth()
const { t, locale, locales, setLocale } = useI18n()
const { isDark, toggle } = useTheme()
const toast = useToast()

const email = ref('')
const password = ref('')
const loading = ref(false)

async function onSubmit() {
  loading.value = true
  try {
    await login(email.value, password.value)
    await navigateTo('/')
  } catch (e) {
    toast.add(t('auth.error'), 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen relative flex items-center justify-center p-4 overflow-hidden
              bg-slate-50 dark:bg-slate-950">
    <!-- Dekorativ gradient orqa fon -->
    <div class="pointer-events-none absolute -top-32 -left-32 h-96 w-96 rounded-full bg-brand-500/20 blur-3xl" />
    <div class="pointer-events-none absolute -bottom-32 -right-32 h-96 w-96 rounded-full bg-accent-500/20 blur-3xl" />

    <!-- Theme toggle (yuqori o'ng) -->
    <button class="absolute top-5 right-5 btn-ghost !p-2.5" @click="toggle">
      <Icon :name="isDark ? 'sun' : 'moon'" class="h-5 w-5" />
    </button>

    <div class="card w-full max-w-sm p-8 relative animate-fade-slide">
      <div class="text-center mb-7">
        <div class="mx-auto mb-4 h-14 w-14 rounded-2xl bg-gradient-to-br from-brand-500 to-accent-500 flex items-center justify-center text-white font-bold text-xl shadow-lg shadow-brand-500/30">
          TF
        </div>
        <h1 class="text-xl font-bold tracking-tight">TTYSI_FIT Admin</h1>
        <p class="text-sm text-slate-400 dark:text-slate-500 mt-1">{{ t('auth.signIn') }}</p>
      </div>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <!-- for/id bog'lanishi majburiy: usiz ekran o'quvchi maydonni e'lon
             qila olmaydi va yorliqni bosish inputga fokus bermaydi. -->
        <div>
          <label for="email" class="block text-sm font-medium mb-1.5">{{ t('auth.email') }}</label>
          <input id="email" v-model="email" type="email" class="input" autocomplete="username" required />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium mb-1.5">{{ t('auth.password') }}</label>
          <input id="password" v-model="password" type="password" class="input" autocomplete="current-password" required />
        </div>
        <button type="submit" class="btn-primary w-full" :disabled="loading">
          {{ loading ? '...' : t('auth.signIn') }}
        </button>
      </form>

      <div class="flex justify-center gap-1 mt-6">
        <button
          v-for="l in (locales as any[])" :key="l.code"
          class="px-2.5 py-1 text-xs rounded-lg transition"
          :class="locale === l.code
            ? 'text-accent-600 dark:text-accent-400 font-bold bg-accent-50 dark:bg-slate-800'
            : 'text-slate-400 hover:text-slate-600 dark:hover:text-slate-300'"
          @click="setLocale(l.code)"
        >{{ l.code.toUpperCase() }}</button>
      </div>
    </div>
  </div>
</template>
