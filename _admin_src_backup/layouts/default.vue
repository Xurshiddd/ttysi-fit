<script setup lang="ts">
const { t, locale, locales, setLocale } = useI18n()
const { user, logout } = useAuth()

const nav = computed(() => [
  { to: '/', label: t('nav.dashboard'), icon: 'i-heroicons-home' },
  { to: '/hemis', label: t('nav.hemis'), icon: 'i-heroicons-arrow-path' },
  { to: '/faculties', label: t('nav.faculties'), icon: 'i-heroicons-building-library' },
  { to: '/departments', label: t('nav.departments'), icon: 'i-heroicons-building-office-2' },
  { to: '/groups', label: t('nav.groups'), icon: 'i-heroicons-user-group' },
  { to: '/users', label: t('nav.users'), icon: 'i-heroicons-users' }
])

const langItems = computed(() => (locales.value as any[]).map((l) => [{
  label: l.name, click: () => setLocale(l.code)
}]))
</script>

<template>
  <div class="min-h-screen flex bg-slate-50 dark:bg-slate-950">
    <!-- Sidebar -->
    <aside class="w-60 shrink-0 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-col">
      <div class="h-16 flex items-center gap-2 px-4 border-b border-slate-200 dark:border-slate-800">
        <div class="h-8 w-8 rounded-lg bg-primary-500 flex items-center justify-center text-white font-bold text-sm">TF</div>
        <span class="font-semibold">TTYSI_FIT</span>
      </div>
      <nav class="flex-1 p-3 space-y-1">
        <NuxtLink
          v-for="item in nav" :key="item.to" :to="item.to"
          class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          active-class="!bg-primary-50 dark:!bg-primary-950 !text-primary-600 dark:!text-primary-400 font-medium"
        >
          <UIcon :name="item.icon" class="h-5 w-5" />
          {{ item.label }}
        </NuxtLink>
      </nav>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-16 shrink-0 flex items-center justify-end gap-2 px-6 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
        <UDropdown :items="langItems">
          <UButton color="gray" variant="ghost" icon="i-heroicons-language" :label="locale.toUpperCase()" />
        </UDropdown>
        <div class="text-sm text-right hidden sm:block">
          <div class="font-medium leading-tight">{{ user?.full_name }}</div>
          <div class="text-xs text-slate-500">{{ user?.email }}</div>
        </div>
        <UButton color="gray" variant="ghost" icon="i-heroicons-arrow-right-on-rectangle" @click="logout" />
      </header>

      <main class="flex-1 p-6 overflow-auto">
        <slot />
      </main>
    </div>
  </div>
</template>
