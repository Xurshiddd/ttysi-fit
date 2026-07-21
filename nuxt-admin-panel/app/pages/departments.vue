<script setup lang="ts">
const { t } = useI18n()
</script>

<template>
  <StructureList endpoint="/departments" :title="t('nav.departments')" />
</template>
