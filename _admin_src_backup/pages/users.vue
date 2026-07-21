<script setup lang="ts">
const { api } = useApi()
const { t } = useI18n()

const rows = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
const role = ref('')
const search = ref('')

const roleOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'student', label: t('roles.student') },
  { value: 'employee', label: t('roles.employee') },
  { value: 'admin', label: t('roles.admin') }
])

const columns = computed(() => [
  { key: 'full_name', label: t('common.name') },
  { key: 'role', label: t('common.role') },
  { key: 'faculty_name', label: t('common.faculty') },
  { key: 'group', label: t('common.group') },
  { key: 'is_active', label: t('common.active') }
])

function roleColor(r: string) {
  return r === 'admin' ? 'red' : r === 'employee' ? 'blue' : r === 'student' ? 'green' : 'gray'
}
function roleLabel(r: string) {
  return t(`roles.${r}`) !== `roles.${r}` ? t(`roles.${r}`) : r
}

async function load() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: String(page.value), limit: String(limit) })
    if (role.value) params.set('role', role.value)
    if (search.value) params.set('search', search.value)
    const res = await api<any>(`/admin/users?${params.toString()}`)
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } finally {
    loading.value = false
  }
}

watch(page, load)
watch(role, () => { page.value = 1; load() })

let timer: any
watch(search, () => {
  clearTimeout(timer)
  timer = setTimeout(() => { page.value = 1; load() }, 350)
})

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.users') }}</h1>
      <div class="flex gap-2">
        <USelectMenu
          v-model="role" :options="roleOptions"
          value-attribute="value" option-attribute="label"
          class="w-40" :placeholder="t('common.role')"
        />
        <UInput v-model="search" :placeholder="t('common.search')" icon="i-heroicons-magnifying-glass" />
        <UButton color="gray" variant="ghost" icon="i-heroicons-arrow-path" :loading="loading" @click="load" />
      </div>
    </div>

    <UCard :ui="{ body: { padding: '' } }">
      <UTable
        :rows="rows" :columns="columns" :loading="loading"
        :empty-state="{ icon: 'i-heroicons-inbox', label: t('common.empty') }"
      >
        <template #full_name-data="{ row }">
          <div class="flex items-center gap-2">
            <UAvatar :src="row.avatar_url || undefined" :alt="row.full_name" size="xs" />
            <div class="min-w-0">
              <div class="font-medium truncate">{{ row.full_name }}</div>
              <div class="text-xs text-slate-500 truncate">{{ row.email || row.hemis_login || '—' }}</div>
            </div>
          </div>
        </template>
        <template #role-data="{ row }">
          <UBadge :color="roleColor(row.role)" variant="subtle" size="xs">{{ roleLabel(row.role) }}</UBadge>
        </template>
        <template #faculty_name-data="{ row }">
          <span class="text-sm">{{ row.faculty_name || '—' }}</span>
        </template>
        <template #group-data="{ row }">
          <span class="text-sm">
            {{ row.group_name || (row.position ? row.position : '—') }}
            <span v-if="row.course" class="text-xs text-slate-400">· {{ row.course }}-{{ t('common.course') }}</span>
          </span>
        </template>
        <template #is_active-data="{ row }">
          <UBadge :color="row.is_active ? 'green' : 'gray'" variant="subtle" size="xs">
            {{ row.is_active ? t('common.active') : t('common.inactive') }}
          </UBadge>
        </template>
      </UTable>

      <div class="flex items-center justify-between px-4 py-3 border-t border-slate-100 dark:border-slate-800">
        <span class="text-sm text-slate-500">{{ t('common.total') }}: {{ total }}</span>
        <UPagination v-model="page" :page-count="limit" :total="total" :max="7" />
      </div>
    </UCard>
  </div>
</template>
