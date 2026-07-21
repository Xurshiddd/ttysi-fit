<script setup lang="ts">
const { api } = useApi()
const { t } = useI18n()

const loading = ref(true)
const stats = ref({ users: 0, faculties: 0, departments: 0, groups: 0 })

const cards = computed(() => [
  { key: 'users', label: t('nav.users'), value: stats.value.users, icon: 'i-heroicons-users', to: '/users', color: 'text-primary-500' },
  { key: 'faculties', label: t('nav.faculties'), value: stats.value.faculties, icon: 'i-heroicons-building-library', to: '/faculties', color: 'text-blue-500' },
  { key: 'departments', label: t('nav.departments'), value: stats.value.departments, icon: 'i-heroicons-building-office-2', to: '/departments', color: 'text-amber-500' },
  { key: 'groups', label: t('nav.groups'), value: stats.value.groups, icon: 'i-heroicons-user-group', to: '/groups', color: 'text-violet-500' }
])

onMounted(async () => {
  try {
    const [u, f, d, g] = await Promise.all([
      api<any>('/admin/users?limit=1'),
      api<any>('/faculties'),
      api<any>('/departments'),
      api<any>('/groups')
    ])
    stats.value = {
      users: u?.meta?.total ?? 0,
      faculties: f?.data?.length ?? 0,
      departments: d?.data?.length ?? 0,
      groups: g?.data?.length ?? 0
    }
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-6">{{ t('nav.dashboard') }}</h1>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <NuxtLink v-for="c in cards" :key="c.key" :to="c.to">
        <UCard class="hover:ring-2 hover:ring-primary-500/40 transition">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm text-slate-500">{{ c.label }}</div>
              <div class="text-3xl font-bold mt-1">
                <span v-if="loading" class="text-slate-300">—</span>
                <span v-else>{{ c.value.toLocaleString() }}</span>
              </div>
            </div>
            <UIcon :name="c.icon" class="h-10 w-10" :class="c.color" />
          </div>
        </UCard>
      </NuxtLink>
    </div>

    <UCard class="mt-6">
      <div class="flex items-center gap-3">
        <UIcon name="i-heroicons-arrow-path" class="h-6 w-6 text-primary-500" />
        <div class="flex-1">
          <div class="font-medium">{{ t('hemis.title') }}</div>
          <div class="text-sm text-slate-500">{{ t('hemis.desc') }}</div>
        </div>
        <UButton :to="'/hemis'" :label="t('nav.hemis')" trailing-icon="i-heroicons-arrow-right" />
      </div>
    </UCard>
  </div>
</template>
