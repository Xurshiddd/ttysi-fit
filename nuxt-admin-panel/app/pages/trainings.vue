<script setup lang="ts">
// Mashg'ulotlar CRUD (CLAUDE.md §16.3).
//
// KATEGORIYA: kodda ro'yxat YO'Q. `GET /training-categories` mavjudlarini
// qaytaradi, admin ulardan tanlaydi yoki yangisini yozadi (datalist). Yangi
// kategoriya qo'shish uchun redeploy shart emas — §16 talabi shu.
//
// DARAJA esa enum: chegaralangan shkala (boshlang'ich/o'rta/yuqori).
const { api } = useApi()
const { t } = useI18n()
const toast = useToast()

const rows = ref<any[]>([])
const categories = ref<string[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20
const statusFilter = ref('')
const categoryFilter = ref('')
const levelFilter = ref('')
const search = ref('')

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const levels = ['beginner', 'intermediate', 'advanced']

function levelLabel(l: string) {
  const v = t(`tr.level.${l}`)
  return v === `tr.level.${l}` ? l : v
}
function levelClass(l: string) {
  if (l === 'advanced') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (l === 'intermediate') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-accent-100 text-accent-700 dark:bg-accent-900/30 dark:text-accent-300'
}
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
  if (categoryFilter.value) params.set('category', categoryFilter.value)
  if (levelFilter.value) params.set('level', levelFilter.value)
  if (searchDebounced.value) params.set('search', searchDebounced.value)
  return params.toString()
})

// reqId — kechikkan javob yangisining ustiga yozmasligi uchun.
let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/admin/trainings?${queryKey.value}`)
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
    const res = await api<any>('/training-categories')
    categories.value = res?.data || []
  } catch {
    // Kategoriya ro'yxati yiqilsa sahifa ishlashda davom etsin — u faqat
    // qulaylik (datalist), majburiy emas.
  }
}

let timer: any
watch(search, () => {
  clearTimeout(timer)
  timer = setTimeout(() => { searchDebounced.value = search.value }, 350)
})

watch([statusFilter, categoryFilter, levelFilter, searchDebounced], () => { page.value = 1 })
watch(queryKey, load)

onMounted(async () => { await loadCategories(); await load() })

// ── Forma ──────────────────────────────────────────────────
const open = ref(false)
const saving = ref(false)
const editId = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})

const form = reactive<any>({
  title: '', description: '', category: '', level: 'beginner',
  video_url: '', thumbnail_url: '', duration_min: null,
  status: 'draft', published_at: '', sort_order: 0
})

function openCreate() {
  editId.value = null
  Object.assign(form, {
    title: '', description: '', category: '', level: 'beginner',
    video_url: '', thumbnail_url: '', duration_min: null,
    status: 'draft', published_at: '', sort_order: 0
  })
  fieldErrors.value = {}
  open.value = true
}

function openEdit(r: any) {
  editId.value = r.id
  Object.assign(form, {
    title: r.title || '',
    description: r.description || '',
    category: r.category || '',
    level: r.level || 'beginner',
    video_url: r.video_url || '',
    thumbnail_url: r.thumbnail_url || '',
    duration_min: r.duration_min ?? null,
    status: r.status || 'draft',
    published_at: toLocalInput(r.published_at),
    sort_order: r.sort_order ?? 0
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
      level: form.level,
      video_url: form.video_url,
      status: form.status,
      sort_order: Number(form.sort_order) || 0
    }
    // Bo'sh URL yubormaymiz: backend `url` validatsiyasi bo'sh satrni rad etadi.
    if (form.thumbnail_url) body.thumbnail_url = form.thumbnail_url
    if (form.duration_min !== null && form.duration_min !== '') {
      body.duration_min = Number(form.duration_min)
    }
    const pa = toRfc3339(form.published_at)
    if (pa) body.published_at = pa

    if (editId.value) {
      await api(`/admin/trainings/${editId.value}`, { method: 'PUT', body })
    } else {
      await api('/admin/trainings', { method: 'POST', body })
    }
    toast.add(t('common.saved'), 'success')
    open.value = false
    // Yangi kategoriya kiritilgan bo'lsa datalist yangilansin.
    await loadCategories()
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
  if (!confirm(t('tr.deleteConfirm').replace('{title}', r.title))) return
  try {
    await api(`/admin/trainings/${r.id}`, { method: 'DELETE' })
    toast.add(t('common.deleted'), 'success')
    await loadCategories()
    await load()
  } catch {
    toast.add(t('common.saveError'), 'error')
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.trainings') }}</h1>
      <div class="flex flex-wrap gap-2">
        <select v-model="statusFilter" class="input w-32">
          <option value="">{{ t('common.all') }}</option>
          <option value="draft">{{ t('news.draft') }}</option>
          <option value="published">{{ t('news.published') }}</option>
        </select>
        <select v-model="categoryFilter" class="input w-36">
          <option value="">{{ t('tr.allCategories') }}</option>
          <option v-for="c in categories" :key="c" :value="c">{{ c }}</option>
        </select>
        <select v-model="levelFilter" class="input w-36">
          <option value="">{{ t('tr.allLevels') }}</option>
          <option v-for="l in levels" :key="l" :value="l">{{ levelLabel(l) }}</option>
        </select>
        <div class="relative">
          <Icon name="search" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="search" :placeholder="t('common.search')" class="input pl-9 w-40" />
        </div>
        <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
        <button class="btn-primary" @click="openCreate">+ {{ t('tr.new') }}</button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[720px]">
          <thead class="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th class="table-th">{{ t('tr.order') }}</th>
              <th class="table-th">{{ t('common.name') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('tr.category') }}</th>
              <th class="table-th">{{ t('tr.level') }}</th>
              <th class="table-th">{{ t('ch.status') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('news.views') }}</th>
              <th class="table-th text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td class="table-td text-center text-slate-400" colspan="7">{{ t('common.loading') }}</td>
            </tr>
            <tr v-else-if="rows.length === 0">
              <td class="table-td text-center text-slate-400" colspan="7">{{ t('common.empty') }}</td>
            </tr>
            <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
              <td class="table-td text-slate-400 text-sm">{{ r.sort_order }}</td>
              <td class="table-td">
                <div class="font-medium truncate max-w-[220px]">{{ r.title }}</div>
                <div class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[220px]">
                  {{ r.duration_min ? r.duration_min + ' ' + t('tr.min') : '—' }}
                </div>
              </td>
              <td class="table-td hidden sm:table-cell">{{ r.category || '—' }}</td>
              <td class="table-td"><span class="badge" :class="levelClass(r.level)">{{ levelLabel(r.level) }}</span></td>
              <td class="table-td">
                <span class="badge" :class="statusClass(r.status)">
                  {{ r.status === 'published' ? t('news.published') : t('news.draft') }}
                </span>
              </td>
              <td class="table-td hidden md:table-cell">{{ r.views }}</td>
              <td class="table-td text-right whitespace-nowrap">
                <a :href="r.video_url" target="_blank" rel="noopener noreferrer" class="btn-ghost px-2 inline-block">
                  {{ t('tr.watch') }}
                </a>
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
    <div v-if="open" role="dialog" aria-modal="true" :aria-label="editId ? t('tr.edit') : t('tr.new')"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-0 sm:p-4" @click.self="open = false">
      <div class="card w-full sm:max-w-2xl max-h-[92vh] overflow-y-auto rounded-b-none sm:rounded-2xl">
        <div class="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-lg font-semibold">{{ editId ? t('tr.edit') : t('tr.new') }}</h2>
          <button class="btn-ghost px-2" @click="open = false"><Icon name="close" /></button>
        </div>

        <div class="p-4 space-y-4">
          <div>
            <label for="tr-title" class="block text-sm mb-1">{{ t('common.name') }}</label>
            <input id="tr-title" v-model="form.title" class="input w-full" :placeholder="t('tr.titlePlaceholder')" />
            <p v-if="fieldErrors.title" class="text-xs text-red-500 mt-1">{{ fieldErrors.title }}</p>
          </div>

          <div>
            <label for="tr-desc" class="block text-sm mb-1">{{ t('ch.description') }}</label>
            <textarea id="tr-desc" v-model="form.description" rows="3" class="input w-full"></textarea>
          </div>

          <div>
            <label for="tr-video" class="block text-sm mb-1">{{ t('tr.videoUrl') }}</label>
            <input id="tr-video" v-model="form.video_url" class="input w-full" placeholder="https://www.youtube.com/watch?v=..." />
            <p v-if="fieldErrors.video_url" class="text-xs text-red-500 mt-1">{{ fieldErrors.video_url }}</p>
          </div>

          <div>
            <label for="tr-thumb" class="block text-sm mb-1">{{ t('tr.thumbnail') }}</label>
            <input id="tr-thumb" v-model="form.thumbnail_url" class="input w-full" placeholder="https://..." />
            <p v-if="fieldErrors.thumbnail_url" class="text-xs text-red-500 mt-1">{{ fieldErrors.thumbnail_url }}</p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label for="tr-category" class="block text-sm mb-1">{{ t('tr.category') }}</label>
              <!-- datalist: mavjudlaridan tanlash YOKI yangisini yozish. Kodda
                   ro'yxat yo'q — u backenddagi ma'lumotdan kelib chiqadi (§16). -->
              <input id="tr-category" v-model="form.category" list="tr-categories" class="input w-full" />
              <datalist id="tr-categories">
                <option v-for="c in categories" :key="c" :value="c" />
              </datalist>
              <p class="text-xs text-slate-400 mt-1">{{ t('tr.categoryHint') }}</p>
            </div>
            <div>
              <label for="tr-level" class="block text-sm mb-1">{{ t('tr.level') }}</label>
              <select id="tr-level" v-model="form.level" class="input w-full">
                <option v-for="l in levels" :key="l" :value="l">{{ levelLabel(l) }}</option>
              </select>
            </div>
            <div>
              <label for="tr-duration" class="block text-sm mb-1">{{ t('tr.duration') }}</label>
              <input id="tr-duration" v-model="form.duration_min" type="number" min="1" class="input w-full" />
              <p v-if="fieldErrors.duration_min" class="text-xs text-red-500 mt-1">{{ fieldErrors.duration_min }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label for="tr-status" class="block text-sm mb-1">{{ t('ch.status') }}</label>
              <select id="tr-status" v-model="form.status" class="input w-full">
                <option value="draft">{{ t('news.draft') }}</option>
                <option value="published">{{ t('news.published') }}</option>
              </select>
            </div>
            <div>
              <label for="tr-published" class="block text-sm mb-1">{{ t('news.publishedAt') }}</label>
              <input id="tr-published" v-model="form.published_at" type="datetime-local" class="input w-full" />
            </div>
            <div>
              <label for="tr-order" class="block text-sm mb-1">{{ t('tr.order') }}</label>
              <input id="tr-order" v-model="form.sort_order" type="number" class="input w-full" />
              <p class="text-xs text-slate-400 mt-1">{{ t('tr.orderHint') }}</p>
            </div>
          </div>
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
