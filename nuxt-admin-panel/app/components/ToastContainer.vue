<script setup lang="ts">
const toasts = useToasts()

const styles: Record<string, string> = {
  success: 'border-accent-500 bg-accent-50 text-accent-800 dark:bg-accent-900/30 dark:text-accent-200',
  error: 'border-red-500 bg-red-50 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  info: 'border-brand-500 bg-brand-50 text-brand-700 dark:bg-brand-900/40 dark:text-brand-200'
}
</script>

<template>
  <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 w-80">
    <div
      v-for="t in toasts" :key="t.id"
      class="card border-l-4 px-4 py-3 text-sm shadow-sm"
      :class="styles[t.type]"
    >
      {{ t.title }}
    </div>
  </div>
</template>
