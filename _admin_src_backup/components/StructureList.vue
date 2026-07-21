<script setup lang="ts">
const props = defineProps<{ endpoint: string; title: string }>()

const { api } = useApi()
const { t } = useI18n()

const rows = ref<any[]>([])
const loading = ref(true)
const q = ref('')

const columns = computed(() => [
  { key: 'name', label: t('common.name'), sortable: true },
  { key: 'code', label: t('common.code') }
])

const filtered = computed(() => {
  if (!q.value) return rows.value
  const s = q.value.toLowerCase()
  return rows.value.filter(
    (r) => (r.name || '').toLowerCase().includes(s) || (r.code || '').toLowerCase().includes(s)
  )
})

async function load() {
  loading.value = true
  try {
    const res = await api<any>(props.endpoint)
    rows.value = res?.data || []
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4 gap-2">
      <h1 class="text-2xl font-semibold">{{ title }}</h1>
      <div class="flex gap-2">
        <UInput v-model="q" :placeholder="t('common.search')" icon="i-heroicons-magnifying-glass" />
        <UButton color="gray" variant="ghost" icon="i-heroicons-arrow-path" :loading="loading" @click="load" />
      </div>
    </div>

    <UCard :ui="{ body: { padding: '' } }">
      <UTable
        :rows="filtered"
        :columns="columns"
        :loading="loading"
        :empty-state="{ icon: 'i-heroicons-inbox', label: t('common.empty') }"
      >
        <template #code-data="{ row }">
          <span class="text-slate-500">{{ row.code || '—' }}</span>
        </template>
      </UTable>
      <div class="px-4 py-2 text-sm text-slate-500 border-t border-slate-100 dark:border-slate-800">
        {{ t('common.total') }}: {{ filtered.length }}
      </div>
    </UCard>
  </div>
</template>
