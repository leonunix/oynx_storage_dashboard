<template>
  <div v-if="!allowed" class="alert alert-warning d-flex align-items-center gap-2 mb-3">
    <i class="bi bi-shield-lock"></i>
    <span>
      {{ $t('mutationGuard.warning') }}
      <code>ONYX_DASHBOARD_ALLOW_DM_MUTATIONS=true</code>
    </span>
  </div>
  <slot />
</template>

<script setup>
defineProps({
  allowed: { type: Boolean, default: false },
})
</script>
