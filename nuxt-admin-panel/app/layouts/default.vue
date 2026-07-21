<script setup lang="ts">
const { t, locale, locales, setLocale } = useI18n()
const { user, logout } = useAuth()
const { isDark, toggle } = useTheme()
const route = useRoute()

// Asosiy menyu
const mainNav = computed(() => [
  { to: '/', label: t('nav.dashboard'), icon: 'grid' },
  { to: '/ratings', label: t('nav.ratings'), icon: 'trophy' },
  { to: '/challenges', label: t('nav.challenges'), icon: 'cap' },
  { to: '/competitions', label: t('nav.competitions'), icon: 'trophy' },
  { to: '/news', label: t('nav.news'), icon: 'bell' },
  { to: '/trainings', label: t('nav.trainings'), icon: 'cap' },
  { to: '/achievements', label: t('nav.achievements'), icon: 'trophy' },
  { to: '/fit-coins', label: t('nav.fitCoins'), icon: 'database' },
  { to: '/rewards', label: t('nav.rewards'), icon: 'briefcase' },
  { to: '/redemptions', label: t('nav.redemptions'), icon: 'grid' },
  { to: '/reports', label: t('nav.reports'), icon: 'grid' },
  { to: '/hemis', label: t('nav.hemis'), icon: 'sync' }
])

// "Ma'lumotlar" bo'limi (dropdown)
const dataNav = computed(() => [
  { to: '/faculties', label: t('nav.faculties'), icon: 'faculty' },
  { to: '/departments', label: t('nav.departments'), icon: 'office' },
  { to: '/groups', label: t('nav.groups'), icon: 'group' },
  { to: '/users', label: t('nav.users'), icon: 'users' }
])

const dataPaths = ['/faculties', '/departments', '/groups', '/users']
const inDataSection = computed(() => dataPaths.includes(route.path))

// Dropdown ochiq holati — data sahifada bo'lsa avtomatik ochiladi
const dataOpen = ref(inDataSection.value)
watch(
  () => route.path,
  (p) => { if (dataPaths.includes(p)) dataOpen.value = true }
)

// Mobil sidebar
const mobileOpen = ref(false)
watch(() => route.path, () => { mobileOpen.value = false })

// Til dropdown
const langOpen = ref(false)

const initials = computed(() => {
  const n = user.value?.full_name?.trim()
  if (!n) return 'AD'
  return n.split(/\s+/).slice(0, 2).map((w) => w[0]?.toUpperCase()).join('')
})

const currentTitle = computed(() => {
  const all = [...mainNav.value, ...dataNav.value]
  return all.find((i) => i.to === route.path)?.label || t('nav.dashboard')
})
</script>

<template>
  <div class="min-h-screen flex bg-slate-50 dark:bg-slate-950">
    <!-- Mobil overlay -->
    <div
      v-if="mobileOpen"
      class="fixed inset-0 z-30 bg-slate-900/50 backdrop-blur-sm lg:hidden"
      @click="mobileOpen = false"
    />

    <!-- ====== SIDEBAR ====== -->
    <aside
      class="fixed lg:sticky top-0 z-40 h-screen w-72 shrink-0 flex flex-col
             border-r border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900
             transition-transform duration-300 lg:translate-x-0"
      :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center gap-3 px-5 border-b border-slate-200 dark:border-slate-800">
        <div class="h-9 w-9 rounded-xl bg-gradient-to-br from-brand-500 to-accent-500 flex items-center justify-center text-white font-bold text-sm shadow-lg shadow-brand-500/30">
          TF
        </div>
        <div class="leading-tight">
          <div class="font-bold tracking-tight">TTYSI_FIT</div>
          <div class="text-[11px] text-slate-400 dark:text-slate-500">Admin Panel</div>
        </div>
        <button class="ml-auto btn-ghost !p-2 lg:hidden" @click="mobileOpen = false">
          <Icon name="close" class="h-5 w-5" />
        </button>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 overflow-y-auto px-3 py-4 space-y-1">
        <p class="px-3 pb-1 text-[11px] font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-600">
          {{ t('nav.menu') }}
        </p>

        <NuxtLink
          v-for="item in mainNav" :key="item.to" :to="item.to"
          class="nav-link" active-class="nav-link-active"
        >
          <Icon :name="item.icon" class="h-5 w-5 shrink-0" />
          <span>{{ item.label }}</span>
        </NuxtLink>

        <!-- ====== MA'LUMOTLAR DROPDOWN ====== -->
        <p class="px-3 pt-4 pb-1 text-[11px] font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-600">
          {{ t('nav.section') }}
        </p>

        <button
          type="button"
          class="nav-link w-full"
          :class="inDataSection && !dataOpen ? 'text-brand-600 dark:text-accent-300' : ''"
          @click="dataOpen = !dataOpen"
        >
          <Icon name="database" class="h-5 w-5 shrink-0" />
          <span>{{ t('nav.data') }}</span>
          <span
            v-if="inDataSection"
            class="ml-1 h-1.5 w-1.5 rounded-full bg-accent-500"
          />
          <Icon
            name="chevron"
            class="h-4 w-4 ml-auto transition-transform duration-300"
            :class="dataOpen ? 'rotate-180' : ''"
          />
        </button>

        <Transition
          enter-active-class="transition-all duration-300 ease-out overflow-hidden"
          leave-active-class="transition-all duration-200 ease-in overflow-hidden"
          enter-from-class="opacity-0 max-h-0"
          enter-to-class="opacity-100 max-h-72"
          leave-from-class="opacity-100 max-h-72"
          leave-to-class="opacity-0 max-h-0"
        >
          <div v-show="dataOpen" class="overflow-hidden">
            <div class="ml-5 mt-1 pl-3 border-l border-slate-200 dark:border-slate-800 space-y-0.5">
              <NuxtLink
                v-for="item in dataNav" :key="item.to" :to="item.to"
                class="sub-link" active-class="sub-link-active"
              >
                <Icon :name="item.icon" class="h-4 w-4 shrink-0" />
                <span>{{ item.label }}</span>
              </NuxtLink>
            </div>
          </div>
        </Transition>
      </nav>

      <!-- Sidebar footer: profile -->
      <div class="border-t border-slate-200 dark:border-slate-800 p-3">
        <div class="flex items-center gap-3 rounded-xl px-2 py-2 bg-slate-50 dark:bg-slate-800/50">
          <div class="h-9 w-9 rounded-full bg-gradient-to-br from-brand-400 to-accent-500 flex items-center justify-center text-white text-xs font-bold">
            {{ initials }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-sm font-medium truncate">{{ user?.full_name || 'Admin' }}</div>
            <div class="text-xs text-slate-400 dark:text-slate-500 truncate">{{ user?.email || '—' }}</div>
          </div>
          <button class="btn-ghost !p-2" :title="t('auth.logout')" @click="logout">
            <Icon name="logout" class="h-5 w-5" />
          </button>
        </div>
      </div>
    </aside>

    <!-- ====== MAIN ====== -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Glass header -->
      <header class="sticky top-0 z-20 h-16 shrink-0 flex items-center gap-3 px-4 sm:px-6
                     border-b border-slate-200 dark:border-slate-800
                     bg-white/70 dark:bg-slate-900/70 backdrop-blur-xl">
        <button class="btn-ghost !p-2 lg:hidden" @click="mobileOpen = true">
          <Icon name="menu" class="h-5 w-5" />
        </button>

        <!-- h1 EMAS: har bir sahifada o'z <h1> i bor va u shu matnni takrorlaydi.
             Ikkita h1 hujjat tuzilishini buzadi (ekran o'quvchi "asosiy sarlavha
             qaysi?" degan savolga javob topolmaydi). Bu — navigatsiya konteksti. -->
        <div class="text-lg font-semibold tracking-tight truncate">{{ currentTitle }}</div>

        <div class="ml-auto flex items-center gap-1.5">
          <!-- Theme toggle -->
          <button
            class="btn-ghost !p-2.5"
            :title="isDark ? t('theme.light') : t('theme.dark')"
            @click="toggle"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" class="h-5 w-5" />
          </button>

          <!-- Notifications (dekorativ) -->
          <button class="btn-ghost !p-2.5 relative hidden sm:inline-flex" :title="t('common.notifications')">
            <Icon name="bell" class="h-5 w-5" />
            <span class="absolute top-2 right-2 h-2 w-2 rounded-full bg-accent-500 ring-2 ring-white dark:ring-slate-900" />
          </button>

          <!-- Language -->
          <div class="relative">
            <button class="btn-ghost !px-3" @click="langOpen = !langOpen">
              <Icon name="language" class="h-5 w-5" />
              <span class="hidden sm:inline">{{ locale.toUpperCase() }}</span>
            </button>
            <div
              v-if="langOpen"
              class="absolute right-0 mt-2 w-36 card p-1 shadow-lg z-30 animate-fade-slide"
              @mouseleave="langOpen = false"
            >
              <button
                v-for="l in (locales as any[])" :key="l.code"
                class="w-full text-left rounded-lg px-3 py-2 text-sm transition hover:bg-slate-100 dark:hover:bg-slate-800"
                :class="locale === l.code ? 'text-accent-600 dark:text-accent-400 font-semibold' : 'text-slate-600 dark:text-slate-300'"
                @click="setLocale(l.code); langOpen = false"
              >{{ l.name }}</button>
            </div>
          </div>
        </div>
      </header>

      <main class="flex-1 p-4 sm:p-6 overflow-auto">
        <div class="animate-fade-slide">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
