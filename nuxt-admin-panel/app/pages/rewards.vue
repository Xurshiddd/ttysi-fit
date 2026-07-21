<script setup lang="ts">
// FIT Coin do'koni — sovg'alar CRUD (CLAUDE.md §16.3: kontent kodda emas,
// admin panel orqali kiritiladi).
//
// KATEGORIYA enum: chegaralangan ro'yxat, backenddan `/reward-categories`
// orqali keladi (kodda takrorlanmaydi — yangi kategoriya qo'shilsa forma
// o'zi moslashadi).
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const rows = ref<any[]>([])
const categories = ref<string[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
const categoryFilter = ref('')

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const queryKey = computed(() => {
  const p = new URLSearchParams({ page: String(page.value), limit: String(limit) })
  if (categoryFilter.value) p.set('category', categoryFilter.value)
  return p.toString()
})

let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/admin/rewards?${queryKey.value}`)
    if (id !== reqId) return
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } catch {
    if (id === reqId) toast.add(t('common.loadError'), 'error')
  } finally {
    if (id === reqId) loading.value = false
  }
}

async function loadCategories() {
  try {
    const res = await api<any>('/reward-categories')
    categories.value = res?.data || []
  } catch {
    // Kategoriya ro'yxati yiqilsa ham sahifa ishlaydi.
  }
}

function catLabel(c: string) {
  const v = t(`rw.cat.${c}`)
  return v === `rw.cat.${c}` ? c : v
}

/// stockLabel — miqdor: null cheksiz degani, 0 tugagan.
function stockLabel(s: number | null) {
  if (s === null || s === undefined) return '∞'
  return String(s)
}

function statusClass(r: any) {
  if (!r.is_active) return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
  if (r.stock === 0) return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  return 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
}
function statusLabel(r: any) {
  if (!r.is_active) return t('rw.inactive')
  if (r.stock === 0) return t('rw.soldOut')
  return t('common.active')
}

watch(categoryFilter, () => { page.value = 1 })
watch(queryKey, load)

onMounted(async () => { await loadCategories(); await load() })

// ── Forma ──────────────────────────────────────────────────
const open = ref(false)
const saving = ref(false)
const editId = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})

const empty = () => ({
  title: '', description: '', image_url: '', category: 'merch',
  cost_coins: 50, stock: '', per_user_limit: '', is_active: true,
  starts_at: '', ends_at: ''
})

const form = reactive<any>(empty())

function openCreate() {
  editId.value = null
  Object.assign(form, empty())
  fieldErrors.value = {}
  open.value = true
}

function openEdit(r: any) {
  editId.value = r.id
  Object.assign(form, {
    title: r.title || '',
    description: r.description || '',
    image_url: r.image_url || '',
    category: r.category || 'merch',
    cost_coins: r.cost_coins ?? 0,
    // null (cheksiz) bo'sh maydonga aylanadi va aksincha.
    stock: r.stock ?? '',
    per_user_limit: r.per_user_limit ?? '',
    is_active: !!r.is_active,
    starts_at: toLocalInput(r.starts_at),
    ends_at: toLocalInput(r.ends_at)
  })
  fieldErrors.value = {}
  open.value = true
}

function toLocalInput(iso: string | null): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}
function toRfc3339(local: string): string | undefined {
  if (!local) return undefined
  const d = new Date(local)
  return isNaN(d.getTime()) ? undefined : d.toISOString()
}

async function save() {
  saving.value = true
  fieldErrors.value = {}
  try {
    const body: any = {
      title: form.title,
      description: form.description,
      category: form.category,
      cost_coins: Number(form.cost_coins) || 0,
      is_active: !!form.is_active
    }
    // Bo'sh URL yubormaymiz: backend `url` validatsiyasi bo'sh satrni rad etadi.
    if (form.image_url) body.image_url = form.image_url
    // Bo'sh maydon = cheksiz (null): backend `omitempty` bilan qabul qiladi.
    if (form.stock !== '' && form.stock !== null) body.stock = Number(form.stock)
    if (form.per_user_limit !== '' && form.per_user_limit !== null) {
      body.per_user_limit = Number(form.per_user_limit)
    }
    const sa = toRfc3339(form.starts_at)
    const ea = toRfc3339(form.ends_at)
    if (sa) body.starts_at = sa
    if (ea) body.ends_at = ea

    if (editId.value) {
      await api(`/admin/rewards/${editId.value}`, { method: 'PUT', body })
    } else {
      await api('/admin/rewards', { method: 'POST', body })
    }
    toast.add(t('common.saved'), 'success')
    open.value = false
    await load()
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    if (data?.fields) {
      fieldErrors.value = data.fields
      toast.add(data.error || t('common.saveError'), 'error')
    } else {
      toast.add(data?.details || data?.error || t('common.saveError'), 'error')
    }
  } finally {
    saving.value = false
  }
}

async function remove(r: any) {
  if (!confirm(t('rw.deleteConfirm').replace('{title}', r.title))) return
  try {
    await api(`/admin/rewards/${r.id}`, { method: 'DELETE' })
    toast.add(t('common.deleted'), 'success')
    await load()
  } catch {
    toast.add(t('common.saveError'), 'error')
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.rewards') }}</h1>
      <div class="flex flex-wrap gap-2">
        <select v-model="categoryFilter" class="input w-40">
          <option value="">{{ t('common.all') }}</option>
          <option v-for="c in categories" :key="c" :value="c">{{ catLabel(c) }}</option>
        </select>
        <button type="button" class="btn-primary" @click="openCreate">
          {{ t('rw.new') }}
        </button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <!-- Jadval mobilda gorizontal scroll ichida (§7.3) -->
      <div class="overflow-x-auto">
        <table class="w-full min-w-[640px]">
          <thead>
            <tr class="bg-slate-50 dark:bg-slate-800/50">
              <th class="table-th">{{ t('common.name') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('rw.category') }}</th>
              <th class="table-th">{{ t('rw.cost') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('rw.stock') }}</th>
              <th class="table-th hidden lg:table-cell">{{ t('rw.perUser') }}</th>
              <th class="table-th">{{ t('common.active') }}</th>
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
              <td class="table-td">
                <div class="flex items-center gap-2 min-w-0">
                  <img
                    v-if="r.image_url" :src="r.image_url" alt=""
                    class="h-8 w-8 rounded-lg object-cover shrink-0"
                  >
                  <div class="min-w-0">
                    <div class="font-medium truncate">{{ r.title }}</div>
                    <div v-if="r.description" class="text-xs text-slate-500 dark:text-slate-400 truncate">
                      {{ r.description }}
                    </div>
                  </div>
                </div>
              </td>
              <td class="table-td hidden md:table-cell">{{ catLabel(r.category) }}</td>
              <td class="table-td font-semibold tabular-nums whitespace-nowrap">{{ r.cost_coins }}</td>
              <td class="table-td hidden sm:table-cell tabular-nums">{{ stockLabel(r.stock) }}</td>
              <td class="table-td hidden lg:table-cell tabular-nums">{{ stockLabel(r.per_user_limit) }}</td>
              <td class="table-td">
                <span class="badge" :class="statusClass(r)">{{ statusLabel(r) }}</span>
              </td>
              <td class="table-td text-right whitespace-nowrap">
                <button type="button" class="btn-ghost text-sm" @click="openEdit(r)">{{ t('common.edit') }}</button>
                <button type="button" class="btn-ghost text-sm text-red-600 dark:text-red-400" @click="remove(r)">
                  {{ t('common.delete') }}
                </button>
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

    <!-- ── Forma ────────────────────────────────────────────── -->
    <div
      v-if="open" role="dialog" aria-modal="true"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/40 p-0 sm:p-4"
      @click.self="open = false"
    >
      <div class="card w-full sm:max-w-lg max-h-[90vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="px-5 py-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="font-semibold">{{ editId ? t('rw.edit') : t('rw.new') }}</h2>
        </div>

        <div class="p-5 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1.5">{{ t('common.name') }}</label>
            <input v-model="form.title" class="input" :placeholder="t('rw.titlePh')">
            <p v-if="fieldErrors.title" class="text-xs text-red-600 mt-1">{{ fieldErrors.title }}</p>
          </div>

          <div>
            <label class="block text-sm font-medium mb-1.5">{{ t('rw.description') }}</label>
            <textarea v-model="form.description" rows="2" class="input" />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1.5">{{ t('rw.category') }}</label>
              <select v-model="form.category" class="input">
                <option v-for="c in categories" :key="c" :value="c">{{ catLabel(c) }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1.5">{{ t('rw.cost') }}</label>
              <input v-model="form.cost_coins" type="number" min="1" class="input">
              <p v-if="fieldErrors.cost_coins" class="text-xs text-red-600 mt-1">{{ fieldErrors.cost_coins }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1.5">{{ t('rw.stock') }}</label>
              <input v-model="form.stock" type="number" min="0" class="input" :placeholder="t('rw.unlimited')">
              <p class="text-xs text-slate-400 mt-1">{{ t('rw.stockHint') }}</p>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1.5">{{ t('rw.perUser') }}</label>
              <input v-model="form.per_user_limit" type="number" min="1" class="input" :placeholder="t('rw.unlimited')">
              <p class="text-xs text-slate-400 mt-1">{{ t('rw.perUserHint') }}</p>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium mb-1.5">{{ t('rw.image') }}</label>
            <input v-model="form.image_url" class="input" placeholder="https://...">
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1.5">{{ t('rw.startsAt') }}</label>
              <input v-model="form.starts_at" type="datetime-local" class="input">
            </div>
            <div>
              <label class="block text-sm font-medium mb-1.5">{{ t('rw.endsAt') }}</label>
              <input v-model="form.ends_at" type="datetime-local" class="input">
            </div>
          </div>

          <label class="flex items-center gap-2 cursor-pointer">
            <input v-model="form.is_active" type="checkbox" class="h-4 w-4 rounded">
            <span class="text-sm">{{ t('rw.isActive') }}</span>
          </label>
        </div>

        <div class="flex justify-end gap-2 px-5 py-4 border-t border-slate-100 dark:border-slate-800">
          <button type="button" class="btn-ghost" @click="open = false">{{ t('common.cancel') }}</button>
          <button type="button" class="btn-primary" :disabled="saving" @click="save">
            {{ saving ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
