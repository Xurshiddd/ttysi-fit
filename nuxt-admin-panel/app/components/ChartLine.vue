<script setup lang="ts">
// ChartLine — kunlik faollik dinamikasi (maydon + chiziq).
//
// Nega tashqi kutubxona emas: ma'lumot sodda (bir qator son), loyihada
// umuman UI bog'liqligi yo'q, canvas esa Tailwind dark mode'ni avtomatik
// olmaydi. Bu yerda SVG faqat chiziqni chizadi.
//
// O'lchamga moslashish hiylasi: preserveAspectRatio="none" bilan SVG
// konteynerni to'liq to'ldiradi, chiziq qalinligi esa
// vector-effect="non-scaling-stroke" tufayli cho'zilmaydi. Matn SVG ichida
// EMAS — HTML'da, shuning uchun 375px telefonda ham tiniq va o'qiladigan.

type Point = { date: string; steps: number; active_users: number }

const props = defineProps<{ points: Point[]; loading?: boolean }>()

const active = ref<number | null>(null)

const maxSteps = computed(() =>
  Math.max(1, ...props.points.map(p => p.steps))
)

// Nuqtalarning foizdagi joylashuvi — SVG ham, HTML nuqtalar ham shundan.
const coords = computed(() =>
  props.points.map((p, i) => ({
    x: props.points.length <= 1 ? 50 : (i / (props.points.length - 1)) * 100,
    y: 100 - (p.steps / maxSteps.value) * 100
  }))
)

const linePath = computed(() =>
  coords.value.map((c, i) => `${i === 0 ? 'M' : 'L'}${c.x},${c.y}`).join(' ')
)

// Maydon uchun yopiq kontur (pastga tushirib yopamiz).
const areaPath = computed(() => {
  if (!coords.value.length) return ''
  const first = coords.value[0]!
  const last = coords.value[coords.value.length - 1]!
  return `M${first.x},100 L${linePath.value.slice(1)} L${last.x},100 Z`
})

// X o'qi belgilari — mobilda tiqilib qolmasin: ko'pi bilan 7 ta ko'rsatamiz.
const xLabels = computed(() => {
  const n = props.points.length
  if (!n) return []
  const step = Math.max(1, Math.ceil(n / 7))
  return props.points.map((p, i) => ({
    i,
    text: (i % step === 0 || i === n - 1) ? p.date.slice(5).replace('-', '.') : ''
  }))
})

const fmt = (n: number) => n.toLocaleString('ru-RU')

const shown = computed(() =>
  active.value !== null ? props.points[active.value] : null
)
</script>

<template>
  <div>
    <!-- Tanlangan nuqta ma'lumoti. Balandligi doim band: qiymat almashganda
         grafik sakramasin. -->
    <div class="h-6 mb-1 text-sm">
      <template v-if="shown">
        <span class="font-semibold">{{ fmt(shown.steps) }}</span>
        <span class="text-slate-500 dark:text-slate-400"> qadam · {{ shown.date }} · {{ shown.active_users }} faol</span>
      </template>
      <span v-else class="text-slate-400 dark:text-slate-500 text-xs">
        Kunni tanlang
      </span>
    </div>

    <div class="relative">
      <!-- Y o'qi eng katta qiymati -->
      <div class="absolute -top-1 left-0 text-[11px] text-slate-400 dark:text-slate-500 pointer-events-none">
        {{ fmt(maxSteps) }}
      </div>

      <svg
        viewBox="0 0 100 100" preserveAspectRatio="none"
        class="w-full h-40 sm:h-48 overflow-visible"
        role="img" aria-label="Kunlik qadam dinamikasi"
      >
        <defs>
          <linearGradient id="chartLineFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="currentColor" stop-opacity="0.28" />
            <stop offset="100%" stop-color="currentColor" stop-opacity="0" />
          </linearGradient>
        </defs>

        <!-- To'r chiziqlari -->
        <line
          v-for="y in [0, 50, 100]" :key="y"
          x1="0" :y1="y" x2="100" :y2="y"
          class="stroke-slate-200 dark:stroke-slate-800"
          stroke-width="1" vector-effect="non-scaling-stroke"
        />

        <template v-if="!loading && points.length">
          <path :d="areaPath" fill="url(#chartLineFill)" class="text-accent-500" />
          <path
            :d="linePath" fill="none"
            class="stroke-accent-500"
            stroke-width="2" stroke-linejoin="round" stroke-linecap="round"
            vector-effect="non-scaling-stroke"
          />
        </template>
      </svg>

      <!-- Nuqtalar: HTML orqali — SVG cho'zilganda aylana ellipsga
           aylanib ketardi. -->
      <div
        v-for="(c, i) in coords" :key="i"
        class="absolute h-2 w-2 -ml-1 -mt-1 rounded-full bg-accent-500 transition-transform pointer-events-none"
        :class="active === i ? 'scale-150 ring-2 ring-accent-500/30' : 'opacity-0 sm:opacity-100'"
        :style="{ left: c.x + '%', top: (c.y / 100 * 100) + '%' }"
      />

      <!-- Sezgir zona: har kunga bitta ustun. Bosish/hover bilan tanlanadi
           (telefonda hover yo'q — shuning uchun click ham bor). Balandligi
           to'liq: barmoq uchun qulay nishon (§7.3). -->
      <div class="absolute inset-0 flex" @mouseleave="active = null">
        <button
          v-for="(p, i) in points" :key="p.date"
          type="button"
          class="flex-1 cursor-pointer focus:outline-none focus:bg-accent-500/5"
          :aria-label="`${p.date}: ${fmt(p.steps)} qadam`"
          @mouseenter="active = i"
          @focus="active = i"
          @click="active = i"
        />
      </div>
    </div>

    <!-- X o'qi belgilari -->
    <div class="flex mt-2">
      <div
        v-for="l in xLabels" :key="l.i"
        class="flex-1 text-center text-[10px] sm:text-[11px] text-slate-400 dark:text-slate-500 truncate"
      >
        {{ l.text }}
      </div>
    </div>
  </div>
</template>
