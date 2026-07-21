<script setup lang="ts">
// Buyurtmalar — foydalanuvchilar coinga almashtirgan sovg'alar.
//
// Admin ikki amal qiladi: TOPSHIRISH (kod bo'yicha shaxsni tekshirib) va
// BEKOR QILISH. Bekor qilinganda coin foydalanuvchiga QAYTADI va sovg'a
// miqdori tiklanadi — buni backend bitta tranzaksiyada bajaradi.
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const rows = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
// Default "pending": adminni birinchi navbatda kutayotgan buyurtmalar qiziqtiradi.
const statusFilter = ref('pending')
const busyId = ref<string | null>(null)

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const queryKey = computed(() => {
  const p = new URLSearchParams({ page: String(page.value), limit: String(limit) })
  if (statusFilter.value) p.set('status', statusFilter.value)
  return p.toString()
})

let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/admin/redemptions?${queryKey.value}`)
    if (id !== reqId) return
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } catch {
    if (id === reqId) toast.add(t('common.loadError'), 'error')
  } finally {
    if (id === reqId) loading.value = false
  }
}

watch(statusFilter, () => { page.value = 1 })
watch(queryKey, load)
onMounted(load)

function statusClass(s: string) {
  if (s === 'issued') return 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
  if (s === 'cancelled') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
}

function fmtDate(iso: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '—' : d.toLocaleDateString('ru-RU')
}

async function issue(r: any) {
  if (!confirm(t('rd.issueConfirm').replace('{code}', r.code))) return
  busyId.value = r.id
  try {
    await api(`/admin/redemptions/${r.id}/issue`, { method: 'POST', body: {} })
    toast.add(t('rd.issued'), 'success')
    await load()
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    toast.add(data?.details || data?.error || t('common.saveError'), 'error')
  } finally {
    busyId.value = null
  }
}

async function cancel(r: any) {
  // Coin qaytishi haqida ogohlantiramiz — bu pul harakati.
  const note = prompt(t('rd.cancelPrompt').replace('{code}', r.code))
  if (note === null) return
  busyId.value = r.id
  try {
    await api(`/admin/redemptions/${r.id}/cancel`, { method: 'POST', body: { note } })
    toast.add(t('rd.cancelled'), 'success')
    await load()
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    toast.add(data?.details || data?.error || t('common.saveError'), 'error')
  } finally {
    busyId.value = null
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.redemptions') }}</h1>
      <select v-model="statusFilter" class="input w-44">
        <option value="">{{ t('common.all') }}</option>
        <option value="pending">{{ t('rd.pending') }}</option>
        <option value="issued">{{ t('rd.issuedStatus') }}</option>
        <option value="cancelled">{{ t('rd.cancelledStatus') }}</option>
      </select>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[720px]">
          <thead>
            <tr class="bg-slate-50 dark:bg-slate-800/50">
              <th class="table-th">{{ t('rd.code') }}</th>
              <th class="table-th">{{ t('rd.reward') }}</th>
              <th class="table-th">{{ t('rd.user') }}</th>
              <th class="table-th">{{ t('rw.cost') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('rd.date') }}</th>
              <th class="table-th">{{ t('rd.status') }}</th>
              <th class="table-th text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="7" class="table-td text-center text-slate-400">{{ t('common.loading') }}</td>
            </tr>
            <tr v-else-if="!rows.length">
              <td colspan="7" class="table-td text-center text-slate-400">{{ t('common.empty') }}</td>
            </tr>
            <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/40">
              <!-- Kod — topshirishda shaxsni tasdiqlaydi, ajratib ko'rsatiladi -->
              <td class="table-td font-mono font-semibold tracking-wider whitespace-nowrap">{{ r.code }}</td>
              <td class="table-td">
                <div class="flex items-center gap-2 min-w-0">
                  <img v-if="r.reward_image_url" :src="r.reward_image_url" alt="" class="h-8 w-8 rounded-lg object-cover shrink-0">
                  <span class="truncate">{{ r.reward_title }}</span>
                </div>
              </td>
              <td class="table-td truncate max-w-[200px]">{{ r.user_full_name }}</td>
              <td class="table-td tabular-nums font-semibold whitespace-nowrap">{{ r.cost_coins }}</td>
              <td class="table-td hidden md:table-cell whitespace-nowrap">{{ fmtDate(r.created_at) }}</td>
              <td class="table-td">
                <span class="badge" :class="statusClass(r.status)">{{ t(`rd.${r.status}Status`) }}</span>
              </td>
              <td class="table-td text-right whitespace-nowrap">
                <template v-if="r.status === 'pending'">
                  <button
                    type="button" class="btn-ghost text-sm text-accent-600 dark:text-accent-400"
                    :disabled="busyId === r.id" @click="issue(r)"
                  >
                    {{ t('rd.issue') }}
                  </button>
                  <button
                    type="button" class="btn-ghost text-sm text-red-600 dark:text-red-400"
                    :disabled="busyId === r.id" @click="cancel(r)"
                  >
                    {{ t('common.cancel') }}
                  </button>
                </template>
                <span v-else class="text-sm text-slate-400">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="pageCount > 1" class="flex items-center justify-between px-4 py-3 border-t border-slate-100 dark:border-slate-800">
        <button type="button" class="btn-ghost text-sm" :disabled="page <= 1" @click="page--">←</button>
        <span class="text-sm text-slate-500">{{ page }} / {{ pageCount }}</span>
        <button type="button" class="btn-ghost text-sm" :disabled="page >= pageCount" @click="page++">→</button>
      </div>
    </div>
  </div>
</template>
