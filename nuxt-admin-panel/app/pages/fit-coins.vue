<script setup lang="ts">
// FIT Coin — admin qo'lda berish/olish (CLAUDE.md §4.3 ledger modeli).
//
// Bu yerda "balansni tahrirlash" YO'Q va bo'lmasligi ham kerak: ledger
// append-only. Admin faqat yangi harakat qo'shadi (musbat — berish, manfiy —
// qaytarish), balans esa SUM(amount) sifatida hisoblanadi. Shu sababli har bir
// coin qayerdan kelgani doim ko'rinib turadi.
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

// Foydalanuvchi qidiruvi (grant uchun kimni tanlash).
const search = ref('')
const users = ref<any[]>([])
const searching = ref(false)
const selected = ref<any | null>(null)

const amount = ref<number | null>(null)
const note = ref('')
const saving = ref(false)

// Tanlangan foydalanuvchi balansi va tarixi.
const balance = ref<any | null>(null)
const history = ref<any[]>([])
const loadingUser = ref(false)

let timer: any
watch(search, () => {
  clearTimeout(timer)
  if (!search.value.trim()) { users.value = []; return }
  timer = setTimeout(searchUsers, 350)
})

async function searchUsers() {
  searching.value = true
  try {
    const params = new URLSearchParams({ page: '1', limit: '10', search: search.value })
    const res = await api<any>(`/admin/users?${params.toString()}`)
    users.value = res?.data || []
  } catch {
    toast.add(t('common.loadError'), 'error')
  } finally {
    searching.value = false
  }
}

async function select(u: any) {
  selected.value = u
  users.value = []
  search.value = ''
  await loadUserCoins()
}

// Tanlangan foydalanuvchi balansi va tarixi (admin endpoint).
// /fit-coins/balance token egasiniki — boshqaning balansi uchun alohida,
// faqat adminga ochiq endpoint ishlatiladi.
async function loadUserCoins() {
  if (!selected.value) return
  loadingUser.value = true
  try {
    const res = await api<any>(`/admin/fit-coins/${selected.value.id}`)
    balance.value = res?.data?.balance || null
    history.value = res?.data?.history || []
  } catch {
    toast.add(t('common.loadError'), 'error')
  } finally {
    loadingUser.value = false
  }
}

function reasonLabel(r: string) {
  const v = t(`coin.reason.${r}`)
  return v === `coin.reason.${r}` ? r : v
}
function fmtDate(iso: string) {
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '—' : d.toLocaleString()
}

async function grant(sign: 1 | -1) {
  if (!selected.value) { toast.add(t('coin.pickUser'), 'error'); return }
  const n = Number(amount.value)
  if (!n || n <= 0) { toast.add(t('coin.badAmount'), 'error'); return }

  saving.value = true
  try {
    await api('/admin/fit-coins/grant', {
      method: 'POST',
      body: { user_id: selected.value.id, amount: sign * n, note: note.value }
    })
    toast.add(t('common.saved'), 'success')
    amount.value = null
    note.value = ''
    await loadUserCoins() // balans va tarix darrov yangilansin
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    toast.add(data?.error || data?.details || t('common.saveError'), 'error')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-1">{{ t('nav.fitCoins') }}</h1>
    <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">{{ t('coin.adminDesc') }}</p>

    <div class="card p-4 space-y-4 max-w-2xl">
      <!-- Foydalanuvchi tanlash -->
      <div>
        <label class="block text-sm mb-1">{{ t('coin.user') }}</label>

        <div v-if="selected" class="flex items-center justify-between gap-2 rounded-xl border border-slate-200 dark:border-slate-700 p-2.5">
          <div class="flex items-center gap-2.5 min-w-0">
            <UserAvatar :src="selected.avatar_url" :name="selected.full_name" :size="36" />
            <div class="min-w-0">
              <div class="font-medium truncate">{{ selected.full_name }}</div>
              <div class="text-xs text-slate-500 dark:text-slate-400 truncate">
                {{ selected.email || selected.hemis_login || '—' }}
              </div>
            </div>
          </div>
          <button class="btn-ghost px-2" @click="selected = null">{{ t('common.cancel') }}</button>
        </div>

        <div v-else class="relative">
          <Icon name="search" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="search" :placeholder="t('common.search')" class="input pl-9 w-full" />

          <div v-if="users.length" class="absolute z-10 mt-1 w-full card max-h-72 overflow-y-auto">
            <button
              v-for="u in users" :key="u.id"
              class="flex items-center gap-2.5 w-full text-left px-3 py-2 hover:bg-slate-50 dark:hover:bg-slate-800/60"
              @click="select(u)"
            >
              <UserAvatar :src="u.avatar_url" :name="u.full_name" :size="30" />
              <div class="min-w-0">
                <div class="text-sm font-medium truncate">{{ u.full_name }}</div>
                <div class="text-xs text-slate-500 dark:text-slate-400 truncate">
                  {{ u.email || u.hemis_login || '—' }}
                </div>
              </div>
            </button>
          </div>
          <p v-if="searching" class="text-xs text-slate-400 mt-1">{{ t('common.loading') }}</p>
        </div>
      </div>

      <!-- Miqdor va izoh -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm mb-1">{{ t('coin.amount') }}</label>
          <input v-model="amount" type="number" min="1" class="input w-full" placeholder="50" />
        </div>
        <div class="sm:col-span-2">
          <label class="block text-sm mb-1">{{ t('coin.note') }}</label>
          <input v-model="note" class="input w-full" :placeholder="t('coin.notePlaceholder')" />
        </div>
      </div>

      <div class="flex flex-wrap gap-2 pt-1">
        <button class="btn-primary" :disabled="saving || !selected" @click="grant(1)">
          + {{ t('coin.grant') }}
        </button>
        <button
          class="btn-ghost text-red-600 dark:text-red-400 border border-red-200 dark:border-red-900"
          :disabled="saving || !selected"
          @click="grant(-1)"
        >
          − {{ t('coin.revoke') }}
        </button>
      </div>

      <p class="text-xs text-slate-400">{{ t('coin.ledgerNote') }}</p>
    </div>

    <!-- Tanlangan foydalanuvchi balansi va ledger tarixi -->
    <div v-if="selected" class="card mt-4 max-w-2xl overflow-hidden">
      <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
        <h2 class="font-semibold">{{ t('coin.history') }}</h2>
        <div v-if="balance" class="flex items-center gap-4 text-sm">
          <span class="text-slate-500 dark:text-slate-400">
            {{ t('coin.earned') }}: <b class="text-accent-600 dark:text-accent-400">{{ balance.earned }}</b>
          </span>
          <span class="text-slate-500 dark:text-slate-400">
            {{ t('coin.spent') }}: <b class="text-red-600 dark:text-red-400">{{ balance.spent }}</b>
          </span>
          <span class="badge bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">
            {{ t('coin.balance') }}: {{ balance.balance }}
          </span>
        </div>
      </div>

      <div v-if="loadingUser" class="p-6 text-center text-slate-400">{{ t('common.loading') }}</div>
      <div v-else-if="history.length === 0" class="p-6 text-center text-slate-400">{{ t('coin.empty') }}</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[520px]">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="table-th">{{ t('coin.reasonCol') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('coin.note') }}</th>
              <th class="table-th">{{ t('coin.date') }}</th>
              <th class="table-th text-right">{{ t('coin.amount') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="tx in history" :key="tx.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/50">
              <td class="table-td">{{ reasonLabel(tx.reason) }}</td>
              <td class="table-td hidden sm:table-cell text-slate-500 dark:text-slate-400 truncate max-w-[200px]">
                {{ tx.note || '—' }}
              </td>
              <td class="table-td text-sm text-slate-500 dark:text-slate-400">{{ fmtDate(tx.created_at) }}</td>
              <td class="table-td text-right font-semibold"
                  :class="tx.amount > 0 ? 'text-accent-600 dark:text-accent-400' : 'text-red-600 dark:text-red-400'">
                {{ tx.amount > 0 ? '+' : '' }}{{ tx.amount }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
