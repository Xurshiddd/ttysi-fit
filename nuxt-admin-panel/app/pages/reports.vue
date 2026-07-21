<script setup lang="ts">
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const period = ref<'week' | 'month' | 'all'>('month')
const facultyId = ref('')
const faculties = ref<any[]>([])
const busy = ref(false)

onMounted(async () => {
  try {
    const f = await api<any>('/faculties')
    faculties.value = f?.data || []
  } catch {
    // Fakultet ro'yxati yuklanmasa ham eksport butun universitet bo'yicha ishlaydi.
  }
})

/**
 * download — hisobotni yuklab oladi.
 *
 * Oddiy <a href="..."> ishlamaydi: endpoint himoyalangan va Authorization
 * header talab qiladi, brauzer esa havolaga header qo'sha olmaydi. Shuning
 * uchun fayl blob sifatida olinadi va vaqtinchalik obyekt-URL orqali
 * saqlanadi.
 */
async function download() {
  busy.value = true
  try {
    const q = new URLSearchParams({ period: period.value })
    if (facultyId.value) q.set('faculty_id', facultyId.value)

    const blob = await api<Blob>(`/admin/reports/users.csv?${q}`, { responseType: 'blob' })

    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ttysi_fit_hisobot_${period.value}.csv`
    document.body.appendChild(a)
    a.click()
    a.remove()
    // Obyekt-URL bo'shatilmasa sahifa yopilguncha xotirada qoladi.
    URL.revokeObjectURL(url)
  } catch {
    toast.add(t('rep.error'), 'error')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-6">{{ t('nav.reports') }}</h1>

    <div class="card p-4 sm:p-6 max-w-3xl">
      <div class="flex items-start gap-3 mb-5">
        <div class="h-11 w-11 rounded-xl bg-accent-50 dark:bg-accent-900/30 flex items-center justify-center shrink-0">
          <Icon name="database" class="h-5 w-5 text-accent-600 dark:text-accent-400" />
        </div>
        <div class="min-w-0">
          <div class="font-medium">{{ t('rep.title') }}</div>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">{{ t('rep.desc') }}</p>
        </div>
      </div>

      <!-- Filtrlar: mobilda ustma-ust, md dan boshlab yonma-yon -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-5">
        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('an.period') }}</label>
          <select v-model="period" class="input">
            <option value="week">{{ t('rating.week') }}</option>
            <option value="month">{{ t('rating.month') }}</option>
            <option value="all">{{ t('rating.all') }}</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('common.faculty') }}</label>
          <select v-model="facultyId" class="input">
            <option value="">{{ t('an.allFaculties') }}</option>
            <option v-for="f in faculties" :key="f.id" :value="f.id">{{ f.name }}</option>
          </select>
        </div>
      </div>

      <div class="rounded-xl bg-slate-50 dark:bg-slate-800/50 px-4 py-3 mb-5">
        <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">{{ t('rep.columns') }}</p>
      </div>

      <button type="button" class="btn-accent w-full sm:w-auto" :disabled="busy" @click="download">
        <Icon name="database" class="h-4 w-4" />
        {{ busy ? t('rep.preparing') : t('rep.download') }}
      </button>
    </div>
  </div>
</template>
