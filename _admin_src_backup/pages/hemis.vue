<script setup lang="ts">
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const items = computed(() => [
  { key: 'structures', label: t('hemis.structures'), path: '/admin/hemis/sync/structures', icon: 'i-heroicons-building-library' },
  { key: 'groups', label: t('hemis.groups'), path: '/admin/hemis/sync/groups', icon: 'i-heroicons-user-group' },
  { key: 'students', label: t('hemis.students'), path: '/admin/hemis/sync/students', icon: 'i-heroicons-academic-cap' },
  { key: 'employees', label: t('hemis.employees'), path: '/admin/hemis/sync/employees', icon: 'i-heroicons-briefcase' }
])

const loading = reactive<Record<string, boolean>>({})
const result = reactive<Record<string, { total: number; created: number; updated: number }>>({})

async function sync(item: { key: string; label: string; path: string }) {
  loading[item.key] = true
  try {
    const res = await api<any>(item.path, { method: 'POST' })
    result[item.key] = res.data
    toast.add({ title: `${item.label}: ${t('hemis.success')}`, color: 'green', icon: 'i-heroicons-check-circle' })
  } catch (e: any) {
    toast.add({ title: `${item.label}: ${e?.data?.error || 'Error'}`, color: 'red', icon: 'i-heroicons-x-circle' })
  } finally {
    loading[item.key] = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold">{{ t('hemis.title') }}</h1>
    <p class="text-sm text-slate-500 mb-6">{{ t('hemis.desc') }}</p>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <UCard v-for="item in items" :key="item.key">
        <div class="flex items-start gap-4">
          <UIcon :name="item.icon" class="h-8 w-8 text-primary-500 shrink-0" />
          <div class="flex-1 min-w-0">
            <div class="font-medium">{{ item.label }}</div>
            <div v-if="result[item.key]" class="text-sm text-slate-500 mt-1">
              {{ t('common.total') }}: <b>{{ result[item.key].total }}</b> ·
              {{ t('hemis.created') }}: <b class="text-green-600">{{ result[item.key].created }}</b> ·
              {{ t('hemis.updated') }}: <b class="text-blue-600">{{ result[item.key].updated }}</b>
            </div>
            <div v-else class="text-sm text-slate-400 mt-1">—</div>
          </div>
          <UButton
            :loading="loading[item.key]"
            :label="loading[item.key] ? t('hemis.syncing') : t('hemis.sync')"
            icon="i-heroicons-arrow-path"
            @click="sync(item)"
          />
        </div>
      </UCard>
    </div>

    <UAlert
      class="mt-6"
      icon="i-heroicons-information-circle"
      color="amber"
      variant="soft"
      :title="t('hemis.desc')"
      description="Strukturalar → Guruhlar → Talabalar → Xodimlar tartibida sinxronlang."
    />
  </div>
</template>
