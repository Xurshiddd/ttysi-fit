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

// ── Faollikni tuzatish ─────────────────────────────────────
//
// NEGA KERAK: qadam yozuvi GREATEST bilan yangilanadi — bir marta yozilgan
// KATTA qiymat qayta sinxron bilan tuzalmaydi. Xato (yoki soxta) yozuv
// paydo bo'lsa uni o'chirishdan boshqa yo'l yo'q edi va buni faqat DB ga
// qo'l bilan kirib qilish mumkin edi.
//
// O'chirilgandan keyin telefon oxirgi 7 kunni qayta yuboradi va haqiqiy
// qiymatlar o'z-o'zidan tiklanadi.
const search = ref('')
const users = ref<any[]>([])
const selected = ref<any | null>(null)
const fixFrom = ref('')
const fixTo = ref('')
const fixing = ref(false)

let timer: any
watch(search, () => {
  clearTimeout(timer)
  if (!search.value.trim()) { users.value = []; return }
  timer = setTimeout(searchUsers, 350)
})

async function searchUsers() {
  try {
    const p = new URLSearchParams({ page: '1', limit: '10', search: search.value })
    const res = await api<any>(`/admin/users?${p}`)
    users.value = res?.data || []
  } catch {
    toast.add(t('common.loadError'), 'error')
  }
}

function selectUser(u: any) {
  selected.value = u
  users.value = []
  search.value = ''
}

async function fixActivity() {
  if (!selected.value || !fixFrom.value || !fixTo.value) return
  if (!confirm(t('fix.confirm')
    .replace('{name}', selected.value.full_name)
    .replace('{from}', fixFrom.value)
    .replace('{to}', fixTo.value))) return

  fixing.value = true
  try {
    const p = new URLSearchParams({ from: fixFrom.value, to: fixTo.value })
    const res = await api<any>(
      `/admin/users/${selected.value.id}/activities?${p}`, { method: 'DELETE' })
    toast.add(t('fix.done').replace('{n}', String(res?.data?.deleted ?? 0)), 'success')
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    toast.add(data?.details || data?.error || t('common.saveError'), 'error')
  } finally {
    fixing.value = false
  }
}

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

    <!-- ── Faollikni tuzatish ──────────────────────────────── -->
    <div class="card p-4 sm:p-6 max-w-3xl mt-6">
      <div class="flex items-start gap-3 mb-5">
        <div class="h-11 w-11 rounded-xl bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center shrink-0">
          <Icon name="refresh" class="h-5 w-5 text-amber-600 dark:text-amber-400" />
        </div>
        <div class="min-w-0">
          <div class="font-medium">{{ t('fix.title') }}</div>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">{{ t('fix.desc') }}</p>
        </div>
      </div>

      <!-- Foydalanuvchi tanlash -->
      <div class="mb-4">
        <label class="block text-sm font-medium mb-1.5">{{ t('coin.user') }}</label>
        <div v-if="selected" class="flex items-center gap-2">
          <div class="flex-1 min-w-0 rounded-xl bg-slate-50 dark:bg-slate-800/50 px-3.5 py-2.5">
            <div class="text-sm font-medium truncate">{{ selected.full_name }}</div>
            <div class="text-xs text-slate-500 truncate">{{ selected.email || selected.hemis_login }}</div>
          </div>
          <button type="button" class="btn-ghost text-sm" @click="selected = null">
            {{ t('common.cancel') }}
          </button>
        </div>
        <div v-else class="relative">
          <input v-model="search" class="input" :placeholder="t('common.search')">
          <ul
            v-if="users.length"
            class="absolute z-10 mt-1 w-full card max-h-60 overflow-y-auto py-1"
          >
            <li v-for="u in users" :key="u.id">
              <button
                type="button"
                class="w-full text-left px-3.5 py-2 hover:bg-slate-50 dark:hover:bg-slate-800"
                @click="selectUser(u)"
              >
                <div class="text-sm font-medium truncate">{{ u.full_name }}</div>
                <div class="text-xs text-slate-500 truncate">{{ u.email || u.hemis_login }}</div>
              </button>
            </li>
          </ul>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-5">
        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('fix.from') }}</label>
          <input v-model="fixFrom" type="date" class="input">
        </div>
        <div>
          <label class="block text-sm font-medium mb-1.5">{{ t('fix.to') }}</label>
          <input v-model="fixTo" type="date" class="input">
        </div>
      </div>

      <div class="rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 mb-5">
        <p class="text-xs text-amber-800 dark:text-amber-300 leading-relaxed">{{ t('fix.warn') }}</p>
      </div>

      <button
        type="button"
        class="btn-primary w-full sm:w-auto"
        :disabled="fixing || !selected || !fixFrom || !fixTo"
        @click="fixActivity"
      >
        {{ fixing ? t('common.loading') : t('fix.action') }}
      </button>
    </div>
  </div>
</template>
