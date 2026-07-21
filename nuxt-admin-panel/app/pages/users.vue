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

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

function roleClass(r: string) {
  if (r === 'admin') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (r === 'employee') return 'bg-brand-100 text-brand-700 dark:bg-brand-900/40 dark:text-brand-200'
  if (r === 'student') return 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
  return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
}
function roleLabel(r: string) {
  const k = `roles.${r}`
  const v = t(k)
  return v === k ? r : v
}

// searchDebounced — kechiktirilgan qidiruv. queryKey `search` ni to'g'ridan-to'g'ri
// kuzatsa, har bosilgan harfda so'rov ketardi.
const searchDebounced = ref('')

// Barcha filtr + sahifa bitta kalitga jamlanadi: bir tick'da bir necha ref
// o'zgarsa ham (filtr almashib, sahifa 1 ga qaytsa) so'rov BITTA ketadi.
const queryKey = computed(() => {
  const params = new URLSearchParams({ page: String(page.value), limit: String(limit) })
  if (role.value) params.set('role', role.value)
  if (searchDebounced.value) params.set('search', searchDebounced.value)
  return params.toString()
})

// reqId — kechikkan javob yangisining ustiga yozmasligi uchun.
let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/admin/users?${queryKey.value}`)
    if (id !== reqId) return
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } finally {
    if (id === reqId) loading.value = false
  }
}

let timer: any
watch(search, () => {
  clearTimeout(timer)
  timer = setTimeout(() => { searchDebounced.value = search.value }, 350)
})

watch([role, searchDebounced], () => { page.value = 1 })
watch(queryKey, load)

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.users') }}</h1>
      <div class="flex gap-2">
        <select v-model="role" class="input w-40">
          <option v-for="o in roleOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <div class="relative">
          <Icon name="search" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="search" :placeholder="t('common.search')" class="input pl-9 w-56" />
        </div>
        <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
      <table class="w-full min-w-[680px]">
        <thead class="bg-slate-50 dark:bg-slate-800/50">
          <tr>
            <th class="table-th">{{ t('common.name') }}</th>
            <th class="table-th">{{ t('common.role') }}</th>
            <th class="table-th hidden md:table-cell">{{ t('common.faculty') }}</th>
            <th class="table-th hidden sm:table-cell">{{ t('common.group') }}</th>
            <th class="table-th">{{ t('common.active') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td class="table-td text-center text-slate-400" colspan="5">{{ t('common.loading') }}</td>
          </tr>
          <tr v-else-if="rows.length === 0">
            <td class="table-td text-center text-slate-400" colspan="5">{{ t('common.empty') }}</td>
          </tr>
          <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
            <td class="table-td">
              <div class="flex items-center gap-2.5">
                <UserAvatar :src="r.avatar_url" :name="r.full_name" :size="36" />
                <div class="min-w-0">
                  <div class="font-medium truncate max-w-[180px]">{{ r.full_name }}</div>
                  <div class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[180px]">{{ r.email || r.hemis_login || '—' }}</div>
                </div>
              </div>
            </td>
            <td class="table-td">
              <span class="badge" :class="roleClass(r.role)">{{ roleLabel(r.role) }}</span>
            </td>
            <td class="table-td hidden md:table-cell">{{ r.faculty_name || '—' }}</td>
            <td class="table-td hidden sm:table-cell">
              {{ r.group_name || r.position || '—' }}
              <span v-if="r.course" class="text-xs text-slate-400">· {{ r.course }}-{{ t('common.course') }}</span>
            </td>
            <td class="table-td">
              <span class="badge" :class="r.is_active ? 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300' : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400'">
                {{ r.is_active ? t('common.active') : t('common.inactive') }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
      </div>

      <div class="flex items-center justify-between px-4 py-3 border-t border-slate-100 dark:border-slate-800">
        <span class="text-sm text-slate-500 dark:text-slate-400">{{ t('common.total') }}: {{ total }}</span>
        <div class="flex items-center gap-2">
          <button class="btn-ghost px-2" :disabled="page <= 1" @click="page--">‹</button>
          <span class="text-sm">{{ page }} / {{ pageCount }}</span>
          <button class="btn-ghost px-2" :disabled="page >= pageCount" @click="page++">›</button>
        </div>
      </div>
    </div>
  </div>
</template>
