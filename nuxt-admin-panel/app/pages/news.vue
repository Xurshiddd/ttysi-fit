<script setup lang="ts">
// Yangiliklar CRUD (CLAUDE.md §16.3: kontent admin panel orqali kiritiladi).
//
// challenges/competitions dan farqli: bu yerda DINAMIK FORMA yo'q va kerak
// emas — yangilikning turga qarab o'zgaradigan maydonlari yo'q. Forma statik.
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const rows = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
const statusFilter = ref('')
const search = ref('')

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'draft', label: t('news.draft') },
  { value: 'published', label: t('news.published') }
])

function statusClass(s: string) {
  return s === 'published'
    ? 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
    : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
}

// searchDebounced — kechiktirilgan qidiruv. queryKey `search` ni to'g'ridan-to'g'ri
// kuzatsa, har bosilgan harfda so'rov ketardi.
const searchDebounced = ref('')

// Barcha filtr + sahifa bitta kalitga jamlanadi: bir tick'da bir necha ref
// o'zgarsa ham (filtr almashib, sahifa 1 ga qaytsa) so'rov BITTA ketadi.
const queryKey = computed(() => {
  const params = new URLSearchParams({ page: String(page.value), limit: String(limit) })
  if (statusFilter.value) params.set('status', statusFilter.value)
  if (searchDebounced.value) params.set('search', searchDebounced.value)
  return params.toString()
})

// reqId — kechikkan javob yangisining ustiga yozmasligi uchun.
let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/admin/news?${queryKey.value}`)
    if (id !== reqId) return
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } catch {
    if (id === reqId) toast.add(t('common.loadError'), 'error')
  } finally {
    if (id === reqId) loading.value = false
  }
}

let timer: any
watch(search, () => {
  clearTimeout(timer)
  timer = setTimeout(() => { searchDebounced.value = search.value }, 350)
})

watch([statusFilter, searchDebounced], () => { page.value = 1 })
watch(queryKey, load)

onMounted(load)

// ── Forma ──────────────────────────────────────────────────
const open = ref(false)
const saving = ref(false)
const editId = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})

const form = reactive({
  title: '', excerpt: '', body: '', cover_url: '',
  status: 'draft', published_at: '', pinned: false
})

function openCreate() {
  editId.value = null
  Object.assign(form, {
    title: '', excerpt: '', body: '', cover_url: '',
    status: 'draft', published_at: '', pinned: false
  })
  fieldErrors.value = {}
  open.value = true
}

// Tahrirlashda to'liq yozuv kerak: ro'yxatda `body` yo'q (backend uni
// ro'yxatga qo'shmaydi — uzun matn), shuning uchun alohida so'raymiz.
async function openEdit(r: any) {
  editId.value = r.id
  fieldErrors.value = {}
  open.value = true
  try {
    const res = await api<any>(`/news/${r.id}`)
    const n = res?.data || {}
    Object.assign(form, {
      title: n.title || '',
      excerpt: n.excerpt || '',
      body: n.body || '',
      cover_url: n.cover_url || '',
      status: n.status || 'draft',
      published_at: toLocalInput(n.published_at),
      pinned: !!n.pinned
    })
  } catch {
    toast.add(t('common.loadError'), 'error')
    open.value = false
  }
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
      excerpt: form.excerpt,
      body: form.body,
      status: form.status,
      pinned: form.pinned
    }
    // Bo'sh URL yubormaymiz: backend `url` validatsiyasi bo'sh satrni rad etadi.
    if (form.cover_url) body.cover_url = form.cover_url
    const pa = toRfc3339(form.published_at)
    if (pa) body.published_at = pa

    if (editId.value) {
      await api(`/admin/news/${editId.value}`, { method: 'PUT', body })
    } else {
      await api('/admin/news', { method: 'POST', body })
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
  if (!confirm(t('news.deleteConfirm').replace('{title}', r.title))) return
  try {
    await api(`/admin/news/${r.id}`, { method: 'DELETE' })
    toast.add(t('common.deleted'), 'success')
    await load()
  } catch {
    toast.add(t('common.saveError'), 'error')
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
      <h1 class="text-2xl font-semibold">{{ t('nav.news') }}</h1>
      <div class="flex flex-wrap gap-2">
        <select v-model="statusFilter" class="input w-36">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <div class="relative">
          <Icon name="search" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="search" :placeholder="t('common.search')" class="input pl-9 w-48" />
        </div>
        <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
        <button class="btn-primary" @click="openCreate">+ {{ t('news.new') }}</button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[680px]">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="table-th">{{ t('common.name') }}</th>
              <th class="table-th">{{ t('ch.status') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('news.publishedAt') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('news.views') }}</th>
              <th class="table-th text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td class="table-td text-center text-slate-400" colspan="5">{{ t('common.loading') }}</td>
            </tr>
            <tr v-else-if="rows.length === 0">
              <td class="table-td text-center text-slate-400" colspan="5">{{ t('common.empty') }}</td>
            </tr>
            <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
              <td class="table-td">
                <div class="flex items-start gap-2">
                  <Icon v-if="r.pinned" name="bell" class="h-4 w-4 mt-0.5 text-accent-500 shrink-0" />
                  <div class="min-w-0">
                    <div class="font-medium truncate max-w-[260px]">{{ r.title }}</div>
                    <div class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[260px]">
                      {{ r.excerpt || '—' }}
                    </div>
                  </div>
                </div>
              </td>
              <td class="table-td">
                <span class="badge" :class="statusClass(r.status)">
                  {{ r.status === 'published' ? t('news.published') : t('news.draft') }}
                </span>
              </td>
              <td class="table-td hidden sm:table-cell text-sm text-slate-500 dark:text-slate-400">
                {{ fmtDate(r.published_at) }}
              </td>
              <td class="table-td hidden md:table-cell">{{ r.views }}</td>
              <td class="table-td text-right whitespace-nowrap">
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
    <div v-if="open" role="dialog" aria-modal="true" :aria-label="editId ? t('news.edit') : t('news.new')"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4" @click.self="open = false">
      <div class="card w-full sm:max-w-2xl max-h-[92vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-lg font-semibold">{{ editId ? t('news.edit') : t('news.new') }}</h2>
          <button class="btn-ghost px-2" @click="open = false"><Icon name="close" /></button>
        </div>

        <div class="p-4 space-y-4">
          <div>
            <label for="news-title" class="block text-sm mb-1">{{ t('common.name') }}</label>
            <input id="news-title" v-model="form.title" class="input w-full" :placeholder="t('news.titlePlaceholder')" />
            <p v-if="fieldErrors.title" class="text-xs text-red-500 mt-1">{{ fieldErrors.title }}</p>
          </div>

          <div>
            <label for="news-excerpt" class="block text-sm mb-1">{{ t('news.excerpt') }}</label>
            <input id="news-excerpt" v-model="form.excerpt" class="input w-full" :placeholder="t('news.excerptHint')" />
            <p v-if="fieldErrors.excerpt" class="text-xs text-red-500 mt-1">{{ fieldErrors.excerpt }}</p>
          </div>

          <div>
            <label for="news-body" class="block text-sm mb-1">{{ t('news.body') }}</label>
            <textarea id="news-body" v-model="form.body" rows="8" class="input w-full"></textarea>
            <p v-if="fieldErrors.body" class="text-xs text-red-500 mt-1">{{ fieldErrors.body }}</p>
          </div>

          <div>
            <label for="news-cover" class="block text-sm mb-1">{{ t('news.cover') }}</label>
            <input id="news-cover" v-model="form.cover_url" class="input w-full" placeholder="https://..." />
            <p v-if="fieldErrors.cover_url" class="text-xs text-red-500 mt-1">{{ fieldErrors.cover_url }}</p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label for="news-status" class="block text-sm mb-1">{{ t('ch.status') }}</label>
              <select id="news-status" v-model="form.status" class="input w-full">
                <option value="draft">{{ t('news.draft') }}</option>
                <option value="published">{{ t('news.published') }}</option>
              </select>
            </div>
            <div>
              <label for="news-published-at" class="block text-sm mb-1">{{ t('news.publishedAt') }}</label>
              <input id="news-published-at" v-model="form.published_at" type="datetime-local" class="input w-full" />
            </div>
            <div class="flex items-end">
              <label class="flex items-center gap-2 text-sm cursor-pointer pb-2">
                <input v-model="form.pinned" type="checkbox" class="h-4 w-4 rounded" />
                {{ t('news.pinned') }}
              </label>
            </div>
          </div>
          <p class="text-xs text-slate-400">{{ t('news.hint') }}</p>
        </div>

        <div class="flex justify-end gap-2 p-4 border-t border-slate-100 dark:border-slate-800">
          <button class="btn-ghost" @click="open = false">{{ t('common.cancel') }}</button>
          <button class="btn-primary" :disabled="saving" @click="save">
            {{ saving ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
