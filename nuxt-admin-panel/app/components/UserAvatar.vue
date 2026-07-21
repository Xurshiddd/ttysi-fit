<script setup lang="ts">
// UserAvatar — rasm bo'lsa ko'rsatadi; rasm yo'q yoki yuklanmasa (404/xato)
// default profil rasmga tushadi: ism bo'lsa initsiallar, bo'lmasa person ikoni.
const props = withDefaults(defineProps<{
  src?: string | null
  name?: string | null
  size?: number // px
}>(), {
  src: '',
  name: '',
  size: 36
}) // eslint-disable-line

const failed = ref(false)
watch(() => props.src, () => { failed.value = false })

const showImage = computed(() => !!props.src && !failed.value)

const initials = computed(() => {
  const n = (props.name || '').trim()
  if (!n) return ''
  return n.split(/\s+/).slice(0, 2).map((w) => w[0]?.toUpperCase() || '').join('')
})

const dim = computed(() => ({ width: `${props.size}px`, height: `${props.size}px` }))
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full
           bg-slate-100 dark:bg-slate-800 ring-1 ring-slate-200 dark:ring-slate-700"
    :style="dim"
  >
    <!-- Haqiqiy rasm -->
    <img
      v-if="showImage"
      :src="src as string"
      :alt="name || 'avatar'"
      class="h-full w-full object-cover"
      loading="lazy"
      referrerpolicy="no-referrer"
      @error="failed = true"
    />

    <!-- Default: initsiallar (ism bor bo'lsa) -->
    <span
      v-else-if="initials"
      class="h-full w-full flex items-center justify-center font-semibold text-white
             bg-gradient-to-br from-brand-400 to-accent-500"
      :style="{ fontSize: `${Math.round(size * 0.38)}px` }"
    >{{ initials }}</span>

    <!-- Default: person ikoni (ism ham yo'q bo'lsa) -->
    <svg
      v-else
      viewBox="0 0 24 24" fill="none"
      class="text-slate-400 dark:text-slate-500"
      :style="{ width: `${Math.round(size * 0.62)}px`, height: `${Math.round(size * 0.62)}px` }"
    >
      <path
        fill="currentColor"
        d="M12 12a5 5 0 1 0 0-10 5 5 0 0 0 0 10Zm0 2c-4.42 0-8 2.69-8 6v1a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-1c0-3.31-3.58-6-8-6Z"
      />
    </svg>
  </span>
</template>
