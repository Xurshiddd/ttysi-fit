<script setup lang="ts">
// Mavzuni global qo'llash — SSR'da ham <html class="dark"> bo'ladi (flashsiz)
const { theme } = useTheme()
useHead({
  htmlAttrs: {
    class: computed(() => (theme.value === 'dark' ? 'dark' : ''))
  }
})
</script>

<template>
  <div>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <ToastContainer />
  </div>
</template>
