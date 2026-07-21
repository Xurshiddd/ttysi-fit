<script setup lang="ts">
// Reyting sahifasi — backend GET /ratings ga ulanadi.
// Kesim (talaba/xodim/guruh/fakultet), davr (hafta/oy/hammasi),
// fakultet/guruh filtri va server-side paginatsiya (CLAUDE.md §14.2).
const { api } = useApi()
const { t } = useI18n()

const rows = ref<any[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20

const type = ref('student')
const period = ref('week')
const facultyId = ref('')
const groupId = ref('')

const faculties = ref<any[]>([])
const groups = ref<any[]>([])

const typeOptions = computed(() => [
  { value: 'student', label: t('roles.student') },
  { value: 'employee', label: t('roles.employee') },
  { value: 'group', label: t('common.group') },
  { value: 'faculty', label: t('common.faculty') }
])
const periodOptions = computed(() => [
  { value: 'week', label: t('rating.week') },
  { value: 'month', label: t('rating.month') },
  { value: 'all', label: t('rating.all') }
])

const isIndividual = computed(() => type.value === 'student' || type.value === 'employee')
const showFaculty = computed(() => type.value !== 'faculty')
const showGroup = computed(() => type.value === 'student')

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / limit)))

// Top-3 uchun medal ranglari
function rankClass(rank: number) {
  if (rank === 1) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300'
  if (rank === 2) return 'bg-slate-200 text-slate-600 dark:bg-slate-700 dark:text-slate-200'
  if (rank === 3) return 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300'
  return 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400'
}
function km(m: number) {
  return ((m || 0) / 1000).toFixed(1)
}
function num(n: number) {
  return new Intl.NumberFormat('ru-RU').format(Math.round(n || 0))
}

// queryKey — barcha filtr + sahifa bitta satrga jamlanadi. Bir tick'da bir necha
// ref o'zgarsa ham (masalan type almashib, sahifa 1 ga qaytsa) kalit bir marta
// qayta hisoblanadi va so'rov BITTA ketadi. Ilgari har filtr o'z watcheriga ega
// bo'lib, ular bir-birini qo'zg'atgan va 2–3 parallel so'rov ketib, ortiqchasi
// uzilib qolgan (backend logida "rating: count: context canceled").
const queryKey = computed(() => {
  const params = new URLSearchParams({
    type: type.value, period: period.value,
    page: String(page.value), limit: String(limit)
  })
  if (facultyId.value && showFaculty.value) params.set('faculty_id', facultyId.value)
  if (groupId.value && showGroup.value) params.set('group_id', groupId.value)
  return params.toString()
})

// reqId — kechikkan javob yangisining ustiga yozmasligi uchun ketma-ketlik raqami.
let reqId = 0

async function load() {
  const id = ++reqId
  loading.value = true
  try {
    const res = await api<any>(`/ratings?${queryKey.value}`)
    if (id !== reqId) return // eskirgan javob — yangi so'rov allaqachon ketgan
    rows.value = res?.data || []
    total.value = res?.meta?.total || 0
  } finally {
    if (id === reqId) loading.value = false
  }
}

async function loadFaculties() {
  const res = await api<any>('/faculties')
  faculties.value = res?.data || []
}

async function loadGroups() {
  if (!facultyId.value) { groups.value = []; return }
  const res = await api<any>(`/groups?faculty_id=${facultyId.value}`)
  groups.value = res?.data || []
}

// Watcherlar tartibi muhim: avval sahifa/guruh holati to'g'rilanadi, keyin
// yagona queryKey watcheri yakuniy holat bilan bitta so'rov yuboradi.
watch([type, period, facultyId, groupId], () => { page.value = 1 })
watch(facultyId, () => { groupId.value = ''; loadGroups() })
watch(queryKey, load)

onMounted(() => { load(); loadFaculties() })
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
      <h1 class="text-2xl font-semibold">{{ t('nav.ratings') }}</h1>
      <button class="btn-ghost" :disabled="loading" @click="load"><Icon name="refresh" /></button>
    </div>

    <!-- Kesim tablari -->
    <div class="flex flex-wrap gap-2 mb-4">
      <button
        v-for="o in typeOptions" :key="o.value"
        class="px-4 py-2 rounded-xl text-sm font-medium transition-colors min-h-[40px]"
        :class="type === o.value
          ? 'bg-brand-600 text-white shadow-lg shadow-brand-500/30'
          : 'bg-white text-slate-600 hover:bg-slate-100 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-800'"
        @click="type = o.value"
      >
        {{ o.label }}
      </button>
    </div>

    <!-- Filtrlar -->
    <div class="flex flex-wrap gap-2 mb-4">
      <select v-model="period" class="input w-40">
        <option v-for="o in periodOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
      </select>
      <select v-if="showFaculty" v-model="facultyId" class="input w-64">
        <option value="">{{ t('common.faculty') }}: {{ t('common.all') }}</option>
        <option v-for="f in faculties" :key="f.id" :value="f.id">{{ f.name }}</option>
      </select>
      <select v-if="showGroup && facultyId" v-model="groupId" class="input w-48">
        <option value="">{{ t('common.group') }}: {{ t('common.all') }}</option>
        <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
      </select>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
      <table class="w-full min-w-[640px]">
        <thead class="bg-slate-50 dark:bg-slate-800/50">
          <tr>
            <th class="table-th w-16">{{ t('rating.rank') }}</th>
            <th class="table-th">{{ t('common.name') }}</th>
            <template v-if="isIndividual">
              <th class="table-th hidden md:table-cell">{{ t('common.faculty') }}</th>
              <th class="table-th">{{ t('rating.steps') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('rating.distance') }}</th>
              <th class="table-th hidden lg:table-cell">{{ t('rating.calories') }}</th>
              <th class="table-th hidden lg:table-cell">{{ t('rating.activeDays') }}</th>
            </template>
            <template v-else>
              <th class="table-th">{{ t('rating.members') }}</th>
              <th class="table-th">{{ t('rating.avgSteps') }}</th>
              <th class="table-th hidden sm:table-cell">{{ t('rating.steps') }}</th>
              <th class="table-th hidden md:table-cell">{{ t('rating.distance') }}</th>
            </template>
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
            <td class="table-td">
              <span class="badge font-semibold" :class="rankClass(r.rank)">{{ r.rank }}</span>
            </td>
            <td class="table-td">
              <div v-if="isIndividual" class="flex items-center gap-2.5">
                <UserAvatar :src="r.avatar_url" :name="r.name" :size="36" />
                <div class="min-w-0">
                  <div class="font-medium truncate max-w-[200px]">{{ r.name }}</div>
                  <div class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[200px]">{{ r.group_name || '—' }}</div>
                </div>
              </div>
              <div v-else class="min-w-0">
                <div class="font-medium truncate max-w-[240px]">{{ r.name }}</div>
                <div v-if="r.faculty_name" class="text-xs text-slate-500 dark:text-slate-400 truncate max-w-[240px]">{{ r.faculty_name }}</div>
              </div>
            </td>
            <template v-if="isIndividual">
              <td class="table-td hidden md:table-cell">{{ r.faculty_name || '—' }}</td>
              <td class="table-td font-semibold">{{ num(r.total_steps) }}</td>
              <td class="table-td hidden sm:table-cell">{{ km(r.total_distance_m) }} km</td>
              <td class="table-td hidden lg:table-cell">{{ num(r.total_calories) }} kkal</td>
              <td class="table-td hidden lg:table-cell">{{ r.active_days || 0 }}</td>
            </template>
            <template v-else>
              <td class="table-td">{{ r.member_count || 0 }}</td>
              <td class="table-td font-semibold">{{ num(r.avg_steps) }}</td>
              <td class="table-td hidden sm:table-cell">{{ num(r.total_steps) }}</td>
              <td class="table-td hidden md:table-cell">{{ km(r.total_distance_m) }} km</td>
            </template>
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
  </div>
</template>
