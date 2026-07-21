<script setup lang="ts">
const { t } = useI18n()
</script>

<template>
  <StructureList endpoint="/groups" :title="t('nav.groups')" />
</template>
