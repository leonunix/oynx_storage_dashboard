<template>
  <div v-if="!allowed" class="alert alert-warning d-flex align-items-center gap-2 mb-3">
    <i class="bi bi-shield-lock"></i>
    <span>
      当前环境禁止执行存储变更操作。如需启用，请设置
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
