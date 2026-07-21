<script setup lang="ts">
// ChartBars — fakultetlar kesimi (gorizontal ustunlar).
//
// Nega gorizontal: fakultet nomlari uzun ("To'qimachilik muhandisligi
// fakulteti"). Vertikal ustunlarda nom 90° burilib o'qib bo'lmas edi,
// mobilda esa umuman sig'masdi. Gorizontalda nom bir qatorda turadi va
// ro'yxat pastga cho'ziladi — telefonda tabiiy.
//
// SVG kerak emas: ustun — kengligi foizda berilgan oddiy div.

type Bar = {
  faculty_id: string
  name: string
  avg_steps: number
  total_steps: number
  user_count: number
  active_users: number
}

const props = defineProps<{ bars: Bar[]; loading?: boolean }>()

// Miqyos eng katta qiymatga nisbatan. 0 bo'lsa ham bo'linish xatosi bo'lmasin.
const max = computed(() => Math.max(1, ...props.bars.map(b => b.avg_steps)))

const fmt = (n: number) => n.toLocaleString('ru-RU')

// Faol foydalanuvchilar ulushi — "qatnashuv" ko'rsatkichi.
const share = (b: Bar) =>
  b.user_count > 0 ? Math.round((b.active_users / b.user_count) * 100) : 0
</script>

<template>
  <div>
    <div v-if="loading" class="py-8 text-center text-slate-400 text-sm">…</div>

    <div v-else-if="!bars.length" class="py-8 text-center text-slate-400 text-sm">
      Ma'lumot yo'q
    </div>

    <ul v-else class="space-y-3">
      <li v-for="b in bars" :key="b.faculty_id">
        <div class="flex items-baseline gap-2 mb-1">
          <span class="text-sm font-medium truncate flex-1 min-w-0" :title="b.name">
            {{ b.name }}
          </span>
          <span class="text-sm font-semibold tabular-nums shrink-0">
            {{ fmt(b.avg_steps) }}
          </span>
        </div>

        <div class="flex items-center gap-2">
          <div class="h-2 flex-1 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
            <div
              class="h-full rounded-full bg-gradient-to-r from-accent-500 to-accent-400 transition-[width] duration-500"
              :style="{ width: Math.max(2, (b.avg_steps / max) * 100) + '%' }"
            />
          </div>
          <!-- Qatnashuv: jami qadam ko'p bo'lsa ham, kam odam qatnashayotgan
               bo'lishi mumkin — hisobot uchun aynan shu muhim. -->
          <span
            class="text-[11px] tabular-nums shrink-0 w-24 text-right text-slate-500 dark:text-slate-400"
          >
            {{ b.active_users }}/{{ b.user_count }} · {{ share(b) }}%
          </span>
        </div>
      </li>
    </ul>

    <p class="mt-3 text-[11px] text-slate-400 dark:text-slate-500">
      Ustun — jon boshiga o'rtacha qadam (katta fakultet shunchaki hajmi bilan
      yutib chiqmasligi uchun). O'ngda — faol / jami foydalanuvchi.
    </p>
  </div>
</template>
