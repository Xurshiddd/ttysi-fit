<script setup lang="ts">
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const items = computed(() => [
  { key: 'structures', label: t('hemis.structures'), path: '/admin/hemis/sync/structures', icon: 'faculty' },
  { key: 'groups', label: t('hemis.groups'), path: '/admin/hemis/sync/groups', icon: 'group' },
  { key: 'students', label: t('hemis.students'), path: '/admin/hemis/sync/students', icon: 'cap' },
  { key: 'employees', label: t('hemis.employees'), path: '/admin/hemis/sync/employees', icon: 'briefcase' }
])

const loading = reactive<Record<string, boolean>>({})
const result = reactive<Record<string, { total: number; created: number; updated: number }>>({})

async function sync(item: { key: string; label: string; path: string }) {
  loading[item.key] = true
  try {
    const res = await api<any>(item.path, { method: 'POST' })
    result[item.key] = res.data
    toast.add(`${item.label}: ${t('hemis.success')}`, 'success')
  } catch (e: any) {
    toast.add(`${item.label}: ${e?.data?.error || 'Error'}`, 'error')
  } finally {
    loading[item.key] = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold">{{ t('hemis.title') }}</h1>
    <p class="text-sm text-slate-500 dark:text-slate-400 mb-6">{{ t('hemis.desc') }}</p>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div v-for="item in items" :key="item.key" class="card p-5">
        <div class="flex items-start gap-4">
          <Icon :name="item.icon" class="h-8 w-8 text-accent-500 shrink-0" />
          <div class="flex-1 min-w-0">
            <div class="font-medium">{{ item.label }}</div>
            <div v-if="result[item.key]" class="text-sm text-slate-500 dark:text-slate-400 mt-1">
              {{ t('common.total') }}: <b>{{ result[item.key].total }}</b> ·
              {{ t('hemis.created') }}: <b class="text-accent-600 dark:text-accent-400">{{ result[item.key].created }}</b> ·
              {{ t('hemis.updated') }}: <b class="text-brand-600 dark:text-brand-300">{{ result[item.key].updated }}</b>
            </div>
            <div v-else class="text-sm text-slate-400 mt-1">—</div>
          </div>
          <button class="btn-primary" :disabled="loading[item.key]" @click="sync(item)">
            {{ loading[item.key] ? t('hemis.syncing') : t('hemis.sync') }}
          </button>
        </div>
      </div>
    </div>

    <div class="card !border-l-4 !border-l-amber-400 bg-amber-50 dark:bg-amber-900/20 p-4 mt-6 text-sm text-amber-800 dark:text-amber-200">
      Strukturalar → Guruhlar → Talabalar → Xodimlar tartibida sinxronlang.
    </div>
  </div>
</template>
