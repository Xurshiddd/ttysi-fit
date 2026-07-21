<script setup lang="ts">
const { api } = useApi()
const { t } = useI18n()

const loading = ref(true)
const anLoading = ref(false)
const stats = ref({ users: 0, faculties: 0, departments: 0, groups: 0 })
const topStudents = ref<any[]>([])
const faculties = ref<any[]>([])

const period = ref<'week' | 'month' | 'all'>('week')
const facultyId = ref('')

const analytics = ref<any>({
  overview: { total_steps: 0, total_distance_km: 0, active_users: 0, total_users: 0, avg_steps_per_active: 0 },
  timeseries: [],
  faculties: []
})

const cards = computed(() => [
  { key: 'users', label: t('nav.users'), value: stats.value.users, icon: 'users', to: '/users', color: 'text-accent-500' },
  { key: 'faculties', label: t('nav.faculties'), value: stats.value.faculties, icon: 'faculty', to: '/faculties', color: 'text-brand-500' },
  { key: 'departments', label: t('nav.departments'), value: stats.value.departments, icon: 'office', to: '/departments', color: 'text-amber-500' },
  { key: 'groups', label: t('nav.groups'), value: stats.value.groups, icon: 'group', to: '/groups', color: 'text-violet-500' }
])

const fmt = (n: number) => Math.round(n).toLocaleString('ru-RU')

// Qamrov — faol / jami. Reyting adolatliligini bir qarashda ko'rsatadi:
// qamrov past bo'lsa reyting ham vakillik qilmaydi.
const coverage = computed(() => {
  const o = analytics.value.overview
  return o.total_users > 0 ? Math.round((o.active_users / o.total_users) * 100) : 0
})

const overviewCards = computed(() => {
  const o = analytics.value.overview
  return [
    { key: 'steps', label: t('an.totalSteps'), value: fmt(o.total_steps) },
    { key: 'dist', label: t('an.distance'), value: `${fmt(o.total_distance_km)} km` },
    { key: 'active', label: t('an.activeUsers'), value: `${fmt(o.active_users)}` , sub: `${coverage.value}% ${t('an.coverage').toLowerCase()}` },
    { key: 'avg', label: t('an.avgSteps'), value: fmt(o.avg_steps_per_active) }
  ]
})

async function loadAnalytics() {
  anLoading.value = true
  try {
    const q = new URLSearchParams({ period: period.value })
    if (facultyId.value) q.set('faculty_id', facultyId.value)
    const r = await api<any>(`/admin/analytics?${q}`)
    if (r?.data) analytics.value = r.data
  } catch {
    // Analitika yiqilsa ham qolgan dashboard ishlayveradi.
  } finally {
    anLoading.value = false
  }
}

// Davr yoki fakultet o'zgarganda qayta yuklaymiz.
watch([period, facultyId], loadAnalytics)

onMounted(async () => {
  try {
    const [u, f, d, g, r] = await Promise.all([
      api<any>('/admin/users?limit=1'),
      api<any>('/faculties'),
      api<any>('/departments'),
      api<any>('/groups'),
      api<any>('/ratings?type=student&period=week&limit=5')
    ])
    stats.value = {
      users: u?.meta?.total ?? 0,
      faculties: f?.data?.length ?? 0,
      departments: d?.data?.length ?? 0,
      groups: g?.data?.length ?? 0
    }
    topStudents.value = r?.data || []
    faculties.value = f?.data || []
  } finally {
    loading.value = false
  }
  await loadAnalytics()
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-6">{{ t('nav.dashboard') }}</h1>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <NuxtLink v-for="c in cards" :key="c.key" :to="c.to" class="card p-5 hover:ring-2 hover:ring-accent-500/40 hover:-translate-y-0.5 transition">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm text-slate-500 dark:text-slate-400">{{ c.label }}</div>
            <div class="text-3xl font-bold mt-1">
              <span v-if="loading" class="text-slate-300 dark:text-slate-600">—</span>
              <span v-else>{{ c.value.toLocaleString() }}</span>
            </div>
          </div>
          <div class="h-12 w-12 rounded-xl bg-slate-50 dark:bg-slate-800 flex items-center justify-center">
            <Icon :name="c.icon" :class="`h-6 w-6 ${c.color}`" />
          </div>
        </div>
      </NuxtLink>
    </div>

    <!-- ── Faollik tahlili ────────────────────────────────────────── -->
    <div class="card mt-6 p-4 sm:p-5">
      <!-- Sarlavha va filtrlar: mobilda ustma-ust, kattaroq ekranda bir qatorda -->
      <div class="flex flex-col sm:flex-row sm:items-center gap-3 mb-4">
        <h2 class="font-medium flex-1">{{ t('an.title') }}</h2>

        <select v-model="facultyId" class="input sm:w-56 py-2 text-sm">
          <option value="">{{ t('an.allFaculties') }}</option>
          <option v-for="f in faculties" :key="f.id" :value="f.id">{{ f.name }}</option>
        </select>

        <div class="flex rounded-xl bg-slate-100 dark:bg-slate-800 p-1 text-sm">
          <button
            v-for="p in (['week', 'month', 'all'] as const)" :key="p"
            type="button"
            class="flex-1 sm:flex-none px-3 py-1.5 rounded-lg font-medium transition min-h-[36px]"
            :class="period === p
              ? 'bg-white dark:bg-slate-700 shadow-sm'
              : 'text-slate-500 dark:text-slate-400'"
            @click="period = p"
          >
            {{ t(`rating.${p}`) }}
          </button>
        </div>
      </div>

      <!-- Umumiy raqamlar -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-5">
        <div
          v-for="o in overviewCards" :key="o.key"
          class="rounded-xl bg-slate-50 dark:bg-slate-800/50 px-3 py-3"
        >
          <div class="text-xs text-slate-500 dark:text-slate-400 truncate">{{ o.label }}</div>
          <div class="text-lg sm:text-xl font-bold mt-0.5 tabular-nums">
            <span v-if="anLoading" class="text-slate-300 dark:text-slate-600">—</span>
            <span v-else>{{ o.value }}</span>
          </div>
          <div v-if="o.sub && !anLoading" class="text-[11px] text-slate-400 dark:text-slate-500">{{ o.sub }}</div>
        </div>
      </div>

      <!-- Kunlik dinamika -->
      <div class="text-sm font-medium mb-2">{{ t('an.dynamics') }}</div>
      <ChartLine :points="analytics.timeseries" :loading="anLoading" />
    </div>

    <!-- ── Fakultetlar kesimi ─────────────────────────────────────── -->
    <div class="card mt-6 p-4 sm:p-5">
      <div class="flex items-center gap-3 mb-4">
        <Icon name="faculty" class="h-5 w-5 text-brand-500" />
        <h2 class="font-medium flex-1">{{ t('an.byFaculty') }}</h2>
        <NuxtLink to="/reports" class="btn-ghost text-sm">{{ t('nav.reports') }} →</NuxtLink>
      </div>
      <ChartBars :bars="analytics.faculties" :loading="anLoading" />
    </div>

    <!-- Haftalik top-5 talaba (reyting modulidan) -->
    <div class="card mt-6 overflow-hidden">
      <div class="flex items-center gap-3 px-5 py-4 border-b border-slate-100 dark:border-slate-800">
        <Icon name="trophy" class="h-5 w-5 text-amber-500" />
        <div class="font-medium flex-1">{{ t('nav.ratings') }} — {{ t('rating.week').toLowerCase() }}</div>
        <NuxtLink to="/ratings" class="btn-ghost text-sm">{{ t('common.all') }} →</NuxtLink>
      </div>
      <div v-if="loading" class="px-5 py-6 text-center text-slate-400">{{ t('common.loading') }}</div>
      <div v-else-if="topStudents.length === 0" class="px-5 py-6 text-center text-slate-400">{{ t('common.empty') }}</div>
      <div v-else>
        <div
          v-for="r in topStudents" :key="r.id"
          class="flex items-center gap-3 px-5 py-3 border-b border-slate-50 dark:border-slate-800/50 last:border-0"
        >
          <span class="w-6 text-center font-semibold text-sm" :class="r.rank <= 3 ? 'text-amber-500' : 'text-slate-400'">{{ r.rank }}</span>
          <UserAvatar :src="r.avatar_url" :name="r.name" :size="32" />
          <div class="min-w-0 flex-1">
            <div class="font-medium text-sm truncate">{{ r.name }}</div>
            <div class="text-xs text-slate-500 dark:text-slate-400 truncate">{{ r.faculty_name || '—' }}</div>
          </div>
          <div class="text-sm font-semibold">{{ (r.total_steps || 0).toLocaleString('ru-RU') }} <span class="text-xs font-normal text-slate-400">{{ t('rating.steps').toLowerCase() }}</span></div>
        </div>
      </div>
    </div>

    <div class="card p-5 mt-6 flex items-center gap-3">
      <Icon name="sync" class="h-6 w-6 text-accent-500" />
      <div class="flex-1">
        <div class="font-medium">{{ t('hemis.title') }}</div>
        <div class="text-sm text-slate-500 dark:text-slate-400">{{ t('hemis.desc') }}</div>
      </div>
      <NuxtLink to="/hemis" class="btn-primary">{{ t('nav.hemis') }}</NuxtLink>
    </div>
  </div>
</template>
