<script setup lang="ts">
const props = withDefaults(defineProps<{
  endpoint: string
  title: string
  perPage?: number
}>(), {
  perPage: 20
})

const { api } = useApi()
const { t } = useI18n()

const rows = ref<any[]>([])
const loading = ref(true)
const q = ref('')
const page = ref(1)

const filtered = computed(() => {
  if (!q.value) return rows.value
  const s = q.value.toLowerCase()
  return rows.value.filter(
    (r) => (r.name || '').toLowerCase().includes(s) || (r.code || '').toLowerCase().includes(s)
  )
})

// Client-side paginatsiya (≤ 1000 yozuv uchun — CLAUDE.md §14).
const pageCount = computed(() => Math.max(1, Math.ceil(filtered.value.length / props.perPage)))
const paged = computed(() =>
  filtered.value.slice((page.value - 1) * props.perPage, page.value * props.perPage)
)

// Qidiruv o'zgarsa 1-sahifaga qaytish; sahifa chegaradan oshmasin.
watch(q, () => { page.value = 1 })
watch(pageCount, (c) => { if (page.value > c) page.value = c })

async function load() {
  loading.value = true
  try {
    const res = await api<any>(props.endpoint)
    rows.value = res?.data || []
    page.value = 1
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between mb-4 gap-2">
      <h1 class="text-2xl font-semibold">{{ title }}</h1>
      <div class="flex gap-2">
        <div class="relative">
          <Icon name="search" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="q" :placeholder="t('common.search')" class="input pl-9 w-full sm:w-56" />
        </div>
        <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[420px]">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="table-th">{{ t('common.name') }}</th>
              <th class="table-th">{{ t('common.code') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td class="table-td text-center text-slate-400" colspan="2">{{ t('common.loading') }}</td>
            </tr>
            <tr v-else-if="filtered.length === 0">
              <td class="table-td text-center text-slate-400" colspan="2">{{ t('common.empty') }}</td>
            </tr>
            <tr
              v-for="r in paged" :key="r.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
            >
              <td class="table-td font-medium">{{ r.name }}</td>
              <td class="table-td text-slate-500 dark:text-slate-400">{{ r.code || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 border-t border-slate-100 dark:border-slate-800">
        <span class="text-sm text-slate-500 dark:text-slate-400">
          {{ t('common.total') }}: {{ filtered.length }}
        </span>
        <div v-if="pageCount > 1" class="flex items-center gap-2">
          <button class="btn-ghost px-2" :disabled="page <= 1" @click="page--">‹</button>
          <span class="text-sm">{{ page }} / {{ pageCount }}</span>
          <button class="btn-ghost px-2" :disabled="page >= pageCount" @click="page++">›</button>
        </div>
      </div>
    </div>
  </div>
</template>
