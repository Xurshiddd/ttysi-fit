<script setup lang="ts">
const { t } = useI18n()
</script>

<template>
  <StructureList endpoint="/faculties" :title="t('nav.faculties')" />
</template>
