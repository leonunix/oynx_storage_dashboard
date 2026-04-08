<template>
  <article class="stat-card">
    <div class="tiny-label">{{ label }}</div>
    <div class="stat-value">{{ value }}</div>
    <div class="stat-note">{{ note }}</div>
  </article>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], required: true },
  note: { type: String, default: '' },
})
</script>
