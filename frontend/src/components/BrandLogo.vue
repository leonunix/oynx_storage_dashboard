<template>
  <div class="brand-lockup" :class="`brand-lockup--${layout}`">
    <img class="brand-lockup__icon" src="/brand/onyx-monolith.svg" alt="Onyx logo" />
    <div class="brand-lockup__text">
      <div v-if="badge" class="brand-lockup__badge">{{ badge }}</div>
      <div class="brand-lockup__title">{{ title }}</div>
      <p v-if="subtitle" class="brand-lockup__subtitle">{{ subtitle }}</p>
    </div>
  </div>
</template>

<script setup>
defineProps({
  badge: { type: String, default: '' },
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  layout: { type: String, default: 'sidebar' },
})
</script>
