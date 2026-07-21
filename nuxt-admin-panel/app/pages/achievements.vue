<script setup lang="ts">
// Yutuqlar va sertifikatlar CRUD (CLAUDE.md §16.3).
//
// MUHIM: turga xos maydonlar bu yerda QATTIQ YOZILMAGAN. Forma
// `GET /achievement-types` javobidan dinamik yasaladi — backendda yangi tur
// qo'shilsa, bu sahifa o'zgarishsiz uni ko'rsatadi.
//
// award_mode formada YO'Q: u turdan kelib chiqadi va backend o'zi to'ldiradi.
// Aks holda admin avtomatik yutuqni "qo'lda" qilib belgilab, mezonni chetlab
// o'tib, xohlagan odamga yutuq bera olardi.
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
interface TypeSpec {
  type: string
  label: string
  fields: Field[]
  award_mode: 'auto' | 'manual'
}

const types = ref<TypeSpec[]>([])
const rows = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
const statusFilter = ref('')
const typeFilter = ref('')

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'draft', label: t('ach.draft') },
  { value: 'active', label: t('ach.active') },
  { value: 'archived', label: t('ach.archived') }
])

function statusClass(s: string) {
  if (s === 'active') return 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
  if (s === 'draft') return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
  return 'bg-brand-100 text-brand-700 dark:bg-brand-900/40 dark:text-brand-200'
}
function statusLabel(s: string) {
  const v = t(`ach.${s}`)
  return v === `ach.${s}` ? s : v
}
function typeLabel(ty: string) {
  return types.value.find((x) => x.type === ty)?.label || ty
}
function isManual(r: any) {
  return r.award_mode === 'manual'
}

// ── Ro'yxat ────────────────────────────────────────────────
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
    const res = await api<any>(`/admin/achievements?${queryKey.value}`)
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
    const res = await api<any>('/achievement-types')
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
  type: '', title: '', description: '', status: 'draft',
  reward_coins: 0, icon_url: '', cover_url: '',
  criteria: {} as Record<string, any>
})

const activeSpec = computed<TypeSpec | undefined>(
  () => types.value.find((x) => x.type === form.type)
)
const activeFields = computed<Field[]>(() => activeSpec.value?.fields || [])

// Tur o'zgarsa mezonni tozalaymiz: eski turning kalitlari qolib ketsa
// backend "bu tur uchun noma'lum maydon" deb rad etadi.
watch(() => form.type, () => {
  form.criteria = {}
  fieldErrors.value = {}
})

function openCreate() {
  editId.value = null
  Object.assign(form, {
    type: types.value[0]?.type || '', title: '', description: '',
    status: 'draft', reward_coins: 0, icon_url: '', cover_url: '', criteria: {}
  })
  fieldErrors.value = {}
  open.value = true
}

function openEdit(r: any) {
  editId.value = r.id
  Object.assign(form, {
    type: r.type,
    title: r.title,
    description: r.description || '',
    status: r.status,
    reward_coins: r.reward_coins || 0,
    icon_url: r.icon_url || '',
    cover_url: r.cover_url || '',
    criteria: { ...(r.criteria || {}) }
  })
  fieldErrors.value = {}
  open.value = true
}

async function save() {
  saving.value = true
  fieldErrors.value = {}
  try {
    // Raqamli maydonlarni son sifatida yuboramiz: input qiymati satr bo'ladi,
    // backend esa mezonda raqam kutadi (ValidateAchievementCriteria).
    const criteria: Record<string, any> = {}
    for (const f of activeFields.value) {
      const v = form.criteria[f.key]
      if (v === undefined || v === '' || v === null) continue
      criteria[f.key] = f.type === 'number' ? Number(v) : v
    }

    const body: any = {
      type: form.type,
      title: form.title,
      description: form.description,
      status: form.status,
      reward_coins: Number(form.reward_coins) || 0,
      criteria
    }
    if (form.icon_url) body.icon_url = form.icon_url
    if (form.cover_url) body.cover_url = form.cover_url

    if (editId.value) {
      await api(`/admin/achievements/${editId.value}`, { method: 'PUT', body })
    } else {
      await api('/admin/achievements', { method: 'POST', body })
    }
    toast.add(t('common.saved'), 'success')
    open.value = false
    await load()
  } catch (e: any) {
    // Backend maydon-bo'yicha xato qaytaradi ({"fields":{"threshold":"..."}})
    // — uni aynan o'sha maydon tagida ko'rsatamiz.
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
  if (!confirm(t('ach.deleteConfirm').replace('{title}', r.title))) return
  try {
    await api(`/admin/achievements/${r.id}`, { method: 'DELETE' })
    toast.add(t('common.deleted'), 'success')
    await load()
  } catch {
    toast.add(t('common.saveError'), 'error')
  }
}

// ── Qo'lda berish ──────────────────────────────────────────
// Faqat award_mode='manual' yutuqlar uchun: musobaqa g'olibi, tadbir ishtiroki.
const awardOpen = ref(false)
const awarding = ref(false)
const awardTarget = ref<any>(null)
const awardNote = ref('')
const search = ref('')
const users = ref<any[]>([])
const searching = ref(false)
const selectedUser = ref<any>(null)

let searchTimer: any
watch(search, () => {
  clearTimeout(searchTimer)
  selectedUser.value = null
  if (!search.value.trim()) { users.value = []; return }
  searchTimer = setTimeout(searchUsers, 350)
})

async function searchUsers() {
  searching.value = true
  try {
    const params = new URLSearchParams({ page: '1', limit: '10', search: search.value })
    const res = await api<any>(`/admin/users?${params.toString()}`)
    users.value = res?.data || []
  } catch {
    users.value = []
  } finally {
    searching.value = false
  }
}

function openAward(r: any) {
  awardTarget.value = r
  awardNote.value = ''
  search.value = ''
  users.value = []
  selectedUser.value = null
  awardOpen.value = true
}

async function submitAward() {
  if (!selectedUser.value) return
  awarding.value = true
  try {
    await api(`/admin/achievements/${awardTarget.value.id}/award`, {
      method: 'POST',
      body: { user_id: selectedUser.value.id, note: awardNote.value }
    })
    toast.add(t('ach.awarded'), 'success')
    awardOpen.value = false
  } catch (e: any) {
    const data = e?.data || e?.response?._data
    toast.add(data?.details || data?.error || t('common.saveError'), 'error')
  } finally {
    awarding.value = false
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.achievements') }}</h1>
      <div class="flex flex-wrap gap-2">
        <select v-model="statusFilter" class="input w-36">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select v-model="typeFilter" class="input w-44">
          <option value="">{{ t('common.all') }}</option>
          <option v-for="ty in types" :key="ty.type" :value="ty.type">{{ ty.label }}</option>
        </select>
        <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
        <button class="btn-primary" @click="openCreate">+ {{ t('ach.new') }}</button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[720px]">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="table-th">{{ t('common.name') }}</th>
              <th class="table-th">{{ t('ach.type') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('ach.awardMode') }}</th>
              <th class="table-th">{{ t('ach.status') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('ach.reward') }}</th>
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
                <div class="font-medium truncate max-w-[220px]">{{ r.title }}</div>
                <div class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[220px]">{{ r.description || '—' }}</div>
              </td>
              <td class="table-td">{{ typeLabel(r.type) }}</td>
              <td class="table-td hidden sm:table-cell">
                <span class="badge" :class="isManual(r)
                  ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300'
                  : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'">
                  {{ isManual(r) ? t('ach.manual') : t('ach.auto') }}
                </span>
              </td>
              <td class="table-td"><span class="badge" :class="statusClass(r.status)">{{ statusLabel(r.status) }}</span></td>
              <td class="table-td hidden md:table-cell">{{ r.reward_coins || 0 }}</td>
              <td class="table-td text-right whitespace-nowrap">
                <button v-if="isManual(r)" class="btn-ghost px-2 text-accent-600 dark:text-accent-400" @click="openAward(r)">
                  {{ t('ach.award') }}
                </button>
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
    <div v-if="open" role="dialog" aria-modal="true" :aria-label="editId ? t('ach.edit') : t('ach.new')"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4" @click.self="open = false">
      <div class="card w-full sm:max-w-2xl max-h-[92vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-lg font-semibold">{{ editId ? t('ach.edit') : t('ach.new') }}</h2>
          <button class="btn-ghost px-2" @click="open = false"><Icon name="close" /></button>
        </div>

        <div class="p-4 space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label for="ach-type" class="block text-sm mb-1">{{ t('ach.type') }}</label>
              <select id="ach-type" v-model="form.type" class="input w-full" :disabled="!!editId">
                <option v-for="ty in types" :key="ty.type" :value="ty.type">{{ ty.label }}</option>
              </select>
              <p v-if="editId" class="text-xs text-slate-400 mt-1">{{ t('ach.typeLocked') }}</p>
              <p v-else-if="activeSpec" class="text-xs text-slate-400 mt-1">
                {{ activeSpec.award_mode === 'manual' ? t('ach.manualHint') : t('ach.autoHint') }}
              </p>
            </div>
            <div>
              <label for="ach-status" class="block text-sm mb-1">{{ t('ach.status') }}</label>
              <select id="ach-status" v-model="form.status" class="input w-full">
                <option value="draft">{{ t('ach.draft') }}</option>
                <option value="active">{{ t('ach.active') }}</option>
                <option value="archived">{{ t('ach.archived') }}</option>
              </select>
            </div>
          </div>

          <div>
            <label for="ach-title" class="block text-sm mb-1">{{ t('common.name') }}</label>
            <input id="ach-title" v-model="form.title" class="input w-full" :placeholder="t('ach.titlePlaceholder')" />
            <p v-if="fieldErrors.title" class="text-xs text-red-500 mt-1">{{ fieldErrors.title }}</p>
          </div>

          <div>
            <label for="ach-desc" class="block text-sm mb-1">{{ t('ach.description') }}</label>
            <textarea id="ach-desc" v-model="form.description" rows="2" class="input w-full"></textarea>
          </div>

          <!-- Turga xos mezon — /achievement-types dan dinamik -->
          <div v-if="activeFields.length" class="rounded-xl border border-slate-200 dark:border-slate-700 p-3 space-y-3">
            <p class="text-xs font-medium text-slate-500 dark:text-slate-400">{{ t('ach.criteria') }}</p>
            <div v-for="f in activeFields" :key="f.key">
              <label :for="`ach-crit-${f.key}`" class="block text-sm mb-1">
                {{ f.label }}
                <span v-if="f.unit" class="text-slate-400">({{ f.unit }})</span>
                <span v-if="f.required" class="text-red-500">*</span>
              </label>

              <input
                v-if="f.type === 'number'"
                :id="`ach-crit-${f.key}`"
                v-model="form.criteria[f.key]"
                type="number"
                :min="f.min"
                step="any"
                class="input w-full"
              />
              <select v-else-if="f.type === 'select'" :id="`ach-crit-${f.key}`" v-model="form.criteria[f.key]" class="input w-full">
                <option value="">—</option>
                <option v-for="o in f.options" :key="o" :value="o">{{ o }}</option>
              </select>
              <input v-else :id="`ach-crit-${f.key}`" v-model="form.criteria[f.key]" class="input w-full" />

              <p v-if="fieldErrors[f.key]" class="text-xs text-red-500 mt-1">{{ fieldErrors[f.key] }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label for="ach-reward" class="block text-sm mb-1">{{ t('ach.reward') }}</label>
              <input id="ach-reward" v-model="form.reward_coins" type="number" min="0" class="input w-full" />
            </div>
            <div>
              <label for="ach-icon" class="block text-sm mb-1">{{ t('ach.iconUrl') }}</label>
              <input id="ach-icon" v-model="form.icon_url" class="input w-full" placeholder="https://…" />
              <p v-if="fieldErrors.icon_url" class="text-xs text-red-500 mt-1">{{ fieldErrors.icon_url }}</p>
            </div>
            <div>
              <label for="ach-cover" class="block text-sm mb-1">{{ t('ach.coverUrl') }}</label>
              <input id="ach-cover" v-model="form.cover_url" class="input w-full" placeholder="https://…" />
              <p v-if="fieldErrors.cover_url" class="text-xs text-red-500 mt-1">{{ fieldErrors.cover_url }}</p>
            </div>
          </div>

          <p class="text-xs text-slate-400">{{ t('ach.certificateHint') }}</p>
        </div>

        <div class="flex justify-end gap-2 p-4 border-t border-slate-100 dark:border-slate-800">
          <button class="btn-ghost" @click="open = false">{{ t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="saving" @click="save">
            {{ saving ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Qo'lda berish modali -->
    <div v-if="awardOpen" role="dialog" aria-modal="true" :aria-label="t('ach.award')"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4" @click.self="awardOpen = false">
      <div class="card w-full sm:max-w-lg max-h-[92vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-lg font-semibold">{{ t('ach.award') }}</h2>
          <button class="btn-ghost px-2" @click="awardOpen = false"><Icon name="close" /></button>
        </div>

        <div class="p-4 space-y-4">
          <div class="rounded-xl bg-slate-50 dark:bg-slate-800/50 p-3">
            <div class="font-medium">{{ awardTarget?.title }}</div>
            <div class="text-xs text-slate-500 dark:text-slate-400">{{ awardTarget?.description || '—' }}</div>
          </div>

          <div>
            <label for="ach-user-search" class="block text-sm mb-1">{{ t('ach.selectUser') }}</label>
            <div class="relative">
              <Icon name="search" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
              <input id="ach-user-search" v-model="search" :placeholder="t('common.search')" class="input pl-9 w-full" />
            </div>
            <p v-if="searching" class="text-xs text-slate-400 mt-1">{{ t('common.loading') }}</p>

            <ul v-if="users.length" class="mt-2 max-h-52 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800">
              <li v-for="u in users" :key="u.id">
                <button
                  class="w-full text-left px-3 py-2 min-h-[40px] hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                  :class="selectedUser?.id === u.id ? 'bg-accent-50 dark:bg-accent-900/20' : ''"
                  @click="selectedUser = u">
                  <div class="text-sm font-medium">{{ u.full_name }}</div>
                  <div class="text-xs text-slate-500 dark:text-slate-400">{{ u.faculty_name || u.email || '—' }}</div>
                </button>
              </li>
            </ul>
          </div>

          <div>
            <label for="ach-award-note" class="block text-sm mb-1">{{ t('ach.note') }}</label>
            <input id="ach-award-note" v-model="awardNote" class="input w-full" :placeholder="t('ach.notePlaceholder')" />
          </div>
        </div>

        <div class="flex justify-end gap-2 p-4 border-t border-slate-100 dark:border-slate-800">
          <button class="btn-ghost" @click="awardOpen = false">{{ t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="awarding || !selectedUser" @click="submitAward">
            {{ awarding ? t('common.loading') : t('ach.award') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
