<script setup lang="ts">
// Musobaqalar CRUD (CLAUDE.md §16.3).
//
// challenges.vue bilan bir xil andoza: turga xos maydonlar QATTIQ YOZILMAGAN,
// forma `GET /competition-types` javobidan dinamik yasaladi. Backendda yangi
// tur qo'shilsa bu sahifa o'zgarishsiz uni ko'rsatadi.
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

interface Field {
  key: string
  label: string
  type: 'number' | 'text' | 'select'
  required: boolean
  min?: number
  options?: string[]
  unit?: string
}
interface TypeSpec { type: string; label: string; fields: Field[] }

const types = ref<TypeSpec[]>([])
const rows = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
const statusFilter = ref('')
const typeFilter = ref('')

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const statuses = ['draft', 'registration', 'ongoing', 'finished']
const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  ...statuses.map((s) => ({ value: s, label: t(`comp.${s}`) }))
])

function statusClass(s: string) {
  if (s === 'registration') return 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
  if (s === 'ongoing') return 'bg-brand-100 text-brand-700 dark:bg-brand-900/40 dark:text-brand-200'
  if (s === 'finished') return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
  return 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400'
}
function statusLabel(s: string) {
  const v = t(`comp.${s}`)
  return v === `comp.${s}` ? s : v
}
function typeLabel(ty: string) {
  return types.value.find((x) => x.type === ty)?.label || ty
}

// Barcha filtr + sahifa bitta kalitga jamlanadi: bir tick'da bir necha ref
// o'zgarsa ham (filtr almashib, sahifa 1 ga qaytsa) so'rov BITTA ketadi.
const queryKey = computed(() => {
  const params = new URLSearchParams({ page: String(page.value), limit: String(limit) })
  if (statusFilter.value) params.set('status', statusFilter.value)
  if (typeFilter.value) params.set('type', typeFilter.value)
  return params.toString()
})

// reqId — kechikkan javob yangisining ustiga yozmasligi uchun.
let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/admin/competitions?${queryKey.value}`)
    if (id !== reqId) return
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } catch {
    if (id === reqId) toast.add(t('common.loadError'), 'error')
  } finally {
    if (id === reqId) loading.value = false
  }
}

async function loadTypes() {
  try {
    const res = await api<any>('/competition-types')
    types.value = res?.data || []
  } catch {
    toast.add(t('common.loadError'), 'error')
  }
}

watch([statusFilter, typeFilter], () => { page.value = 1 })
watch(queryKey, load)
onMounted(async () => { await loadTypes(); await load() })

// ── Forma ──────────────────────────────────────────────────
const open = ref(false)
const saving = ref(false)
const editId = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})

const form = reactive<any>({
  type: '', title: '', description: '', scope: 'university', status: 'draft',
  location: '', max_participants: null, reward_coins: 0,
  starts_at: '', ends_at: '', reg_ends_at: '', config: {} as Record<string, any>
})

const activeFields = computed<Field[]>(
  () => types.value.find((x) => x.type === form.type)?.fields || []
)

// Tur o'zgarsa config tozalanadi: eski turning kalitlari qolsa backend
// "bu tur uchun noma'lum maydon" deb rad etadi.
watch(() => form.type, () => { form.config = {}; fieldErrors.value = {} })

function openCreate() {
  editId.value = null
  Object.assign(form, {
    type: types.value[0]?.type || '', title: '', description: '',
    scope: 'university', status: 'draft', location: '',
    max_participants: null, reward_coins: 0,
    starts_at: '', ends_at: '', reg_ends_at: '', config: {}
  })
  fieldErrors.value = {}
  open.value = true
}

function openEdit(r: any) {
  editId.value = r.id
  Object.assign(form, {
    type: r.type, title: r.title, description: r.description || '',
    scope: r.scope || 'university', status: r.status,
    location: r.location || '',
    max_participants: r.max_participants ?? null,
    reward_coins: r.reward_coins || 0,
    starts_at: toLocalInput(r.starts_at),
    ends_at: toLocalInput(r.ends_at),
    reg_ends_at: toLocalInput(r.reg_ends_at),
    config: { ...(r.config || {}) }
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
    // Raqamli maydonlar son bo'lib ketishi kerak — input satr beradi.
    const config: Record<string, any> = {}
    for (const f of activeFields.value) {
      const v = form.config[f.key]
      if (v === undefined || v === '' || v === null) continue
      config[f.key] = f.type === 'number' ? Number(v) : v
    }

    const body: any = {
      type: form.type,
      title: form.title,
      description: form.description,
      scope: form.scope,
      status: form.status,
      location: form.location,
      reward_coins: Number(form.reward_coins) || 0,
      config
    }
    // max_participants: bo'sh qoldirilsa yubormaymiz (cheklovsiz).
    if (form.max_participants !== null && form.max_participants !== '') {
      body.max_participants = Number(form.max_participants)
    }
    const sa = toRfc3339(form.starts_at)
    const ea = toRfc3339(form.ends_at)
    const ra = toRfc3339(form.reg_ends_at)
    if (sa) body.starts_at = sa
    if (ea) body.ends_at = ea
    if (ra) body.reg_ends_at = ra

    if (editId.value) {
      await api(`/admin/competitions/${editId.value}`, { method: 'PUT', body })
    } else {
      await api('/admin/competitions', { method: 'POST', body })
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
  if (!confirm(t('comp.deleteConfirm').replace('{title}', r.title))) return
  try {
    await api(`/admin/competitions/${r.id}`, { method: 'DELETE' })
    toast.add(t('common.deleted'), 'success')
    await load()
  } catch {
    toast.add(t('common.saveError'), 'error')
  }
}

// ── Ishtirokchilar ─────────────────────────────────────────
const partsOpen = ref(false)
const parts = ref<any[]>([])
const partsTitle = ref('')
const partsLoading = ref(false)

async function showParticipants(r: any) {
  partsTitle.value = r.title
  partsOpen.value = true
  partsLoading.value = true
  parts.value = []
  try {
    const res = await api<any>(`/admin/competitions/${r.id}/participants?limit=100`)
    parts.value = res?.data || []
  } catch {
    toast.add(t('common.loadError'), 'error')
  } finally {
    partsLoading.value = false
  }
}

function fmtDate(iso: string | null) {
  if (!iso) return '—'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '—' : d.toLocaleDateString()
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.competitions') }}</h1>
      <div class="flex flex-wrap gap-2">
        <select v-model="statusFilter" class="input w-40">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select v-model="typeFilter" class="input w-40">
          <option value="">{{ t('common.all') }}</option>
          <option v-for="ty in types" :key="ty.type" :value="ty.type">{{ ty.label }}</option>
        </select>
        <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
        <button class="btn-primary" @click="openCreate">+ {{ t('comp.new') }}</button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[760px]">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="table-th">{{ t('common.name') }}</th>
              <th class="table-th">{{ t('ch.type') }}</th>
              <th class="table-th">{{ t('ch.status') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('comp.slots') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('comp.startsAt') }}</th>
              <th class="table-th text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td class="table-td text-center text-slate-400" colspan="6">{{ t('common.loading') }}</td>
            </tr>
            <tr v-else-if="rows.length === 0">
              <td class="table-td text-center text-slate-400" colspan="6">{{ t('common.empty') }}</td>
            </tr>
            <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
              <td class="table-td">
                <div class="font-medium truncate max-w-[200px]">{{ r.title }}</div>
                <div class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[200px]">
                  {{ r.location || r.description || '—' }}
                </div>
              </td>
              <td class="table-td">{{ typeLabel(r.type) }}</td>
              <td class="table-td"><span class="badge" :class="statusClass(r.status)">{{ statusLabel(r.status) }}</span></td>
              <td class="table-td hidden sm:table-cell">
                {{ r.max_participants ? r.max_participants : t('comp.unlimited') }}
              </td>
              <td class="table-td hidden md:table-cell text-sm text-slate-500 dark:text-slate-400">
                {{ fmtDate(r.starts_at) }}
              </td>
              <td class="table-td text-right whitespace-nowrap">
                <button class="btn-ghost px-2" @click="showParticipants(r)">{{ t('comp.participants') }}</button>
                <button class="btn-ghost px-2" @click="openEdit(r)">{{ t('common.edit') }}</button>
                <button class="btn-ghost px-2 text-red-600 dark:text-red-400" @click="remove(r)">{{ t('common.delete') }}</button>
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

    <!-- Forma modali -->
    <div v-if="open" role="dialog" aria-modal="true" :aria-label="editId ? t('comp.edit') : t('comp.new')"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4" @click.self="open = false">
      <div class="card w-full sm:max-w-2xl max-h-[92vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-lg font-semibold">{{ editId ? t('comp.edit') : t('comp.new') }}</h2>
          <button class="btn-ghost px-2" @click="open = false"><Icon name="close" /></button>
        </div>

        <div class="p-4 space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm mb-1">{{ t('ch.type') }}</label>
              <select v-model="form.type" class="input w-full" :disabled="!!editId">
                <option v-for="ty in types" :key="ty.type" :value="ty.type">{{ ty.label }}</option>
              </select>
              <p v-if="editId" class="text-xs text-slate-400 mt-1">{{ t('ch.typeLocked') }}</p>
            </div>
            <div>
              <label class="block text-sm mb-1">{{ t('ch.status') }}</label>
              <select v-model="form.status" class="input w-full">
                <option v-for="s in statuses" :key="s" :value="s">{{ t(`comp.${s}`) }}</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-sm mb-1">{{ t('common.name') }}</label>
            <input v-model="form.title" class="input w-full" :placeholder="t('comp.titlePlaceholder')" />
            <p v-if="fieldErrors.title" class="text-xs text-red-500 mt-1">{{ fieldErrors.title }}</p>
          </div>

          <div>
            <label class="block text-sm mb-1">{{ t('ch.description') }}</label>
            <textarea v-model="form.description" rows="2" class="input w-full"></textarea>
          </div>

          <!-- Turga xos maydonlar — /competition-types dan dinamik -->
          <div v-if="activeFields.length" class="rounded-xl border border-slate-200 dark:border-slate-700 p-3 space-y-3">
            <p class="text-xs font-medium text-slate-500 dark:text-slate-400">{{ t('ch.typeParams') }}</p>
            <div v-for="f in activeFields" :key="f.key">
              <label class="block text-sm mb-1">
                {{ f.label }}
                <span v-if="f.unit" class="text-slate-400">({{ f.unit }})</span>
                <span v-if="f.required" class="text-red-500">*</span>
              </label>
              <input v-if="f.type === 'number'" v-model="form.config[f.key]" type="number" :min="f.min" step="any" class="input w-full" />
              <select v-else-if="f.type === 'select'" v-model="form.config[f.key]" class="input w-full">
                <option value="">—</option>
                <option v-for="o in f.options" :key="o" :value="o">{{ o }}</option>
              </select>
              <input v-else v-model="form.config[f.key]" class="input w-full" />
              <p v-if="fieldErrors[f.key]" class="text-xs text-red-500 mt-1">{{ fieldErrors[f.key] }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-sm mb-1">{{ t('comp.location') }}</label>
              <input v-model="form.location" class="input w-full" />
            </div>
            <div>
              <label class="block text-sm mb-1">{{ t('comp.slots') }}</label>
              <input v-model="form.max_participants" type="number" min="0" class="input w-full" :placeholder="t('comp.unlimited')" />
            </div>
            <div>
              <label class="block text-sm mb-1">{{ t('ch.reward') }}</label>
              <input v-model="form.reward_coins" type="number" min="0" class="input w-full" />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-sm mb-1">{{ t('comp.regEndsAt') }}</label>
              <input v-model="form.reg_ends_at" type="datetime-local" class="input w-full" />
            </div>
            <div>
              <label class="block text-sm mb-1">{{ t('comp.startsAt') }}</label>
              <input v-model="form.starts_at" type="datetime-local" class="input w-full" />
            </div>
            <div>
              <label class="block text-sm mb-1">{{ t('ch.endsAt') }}</label>
              <input v-model="form.ends_at" type="datetime-local" class="input w-full" />
            </div>
          </div>
          <p class="text-xs text-slate-400">{{ t('comp.regHint') }}</p>
        </div>

        <div class="flex justify-end gap-2 p-4 border-t border-slate-100 dark:border-slate-800">
          <button class="btn-ghost" @click="open = false">{{ t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="saving" @click="save">
            {{ saving ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Ishtirokchilar modali -->
    <div v-if="partsOpen" role="dialog" aria-modal="true" :aria-label="t('comp.participants')"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4" @click.self="partsOpen = false">
      <div class="card w-full sm:max-w-lg max-h-[80vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="font-semibold truncate">{{ partsTitle }}</h2>
          <button class="btn-ghost px-2" @click="partsOpen = false"><Icon name="close" /></button>
        </div>
        <div v-if="partsLoading" class="p-6 text-center text-slate-400">{{ t('common.loading') }}</div>
        <div v-else-if="parts.length === 0" class="p-6 text-center text-slate-400">{{ t('comp.noParticipants') }}</div>
        <ul v-else class="divide-y divide-slate-100 dark:divide-slate-800">
          <li v-for="p in parts" :key="p.id" class="flex items-center justify-between px-4 py-2.5 text-sm">
            <span class="font-mono text-xs text-slate-500 dark:text-slate-400 truncate">{{ p.user_id }}</span>
            <span v-if="p.place" class="badge bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">
              {{ p.place }}
            </span>
            <span v-else class="text-xs text-slate-400">{{ fmtDate(p.registered_at) }}</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
