<script setup lang="ts">
definePageMeta({ layout: false })

const { login } = useAuth()
const { t, locale, locales, setLocale } = useI18n()
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
    toast.add({ title: t('auth.error'), color: 'red', icon: 'i-heroicons-x-circle' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950 p-4">
    <UCard class="w-full max-w-sm">
      <template #header>
        <div class="text-center">
          <div class="mx-auto mb-2 h-12 w-12 rounded-xl bg-primary-500 flex items-center justify-center text-white font-bold text-lg">TF</div>
          <h1 class="text-lg font-semibold">TTYSI_FIT Admin</h1>
        </div>
      </template>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <UFormGroup :label="t('auth.email')" required>
          <UInput v-model="email" type="email" autocomplete="username" icon="i-heroicons-envelope" required />
        </UFormGroup>
        <UFormGroup :label="t('auth.password')" required>
          <UInput v-model="password" type="password" autocomplete="current-password" icon="i-heroicons-lock-closed" required />
        </UFormGroup>
        <UButton type="submit" block :loading="loading" :label="t('auth.signIn')" />
      </form>

      <template #footer>
        <div class="flex justify-center gap-1">
          <UButton
            v-for="l in (locales as any[])" :key="l.code"
            size="2xs" variant="ghost"
            :color="locale === l.code ? 'primary' : 'gray'"
            :label="l.code.toUpperCase()"
            @click="setLocale(l.code)"
          />
        </div>
      </template>
    </UCard>
  </div>
</template>
