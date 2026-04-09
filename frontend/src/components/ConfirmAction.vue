<template>
  <span v-if="!confirming">
    <button :class="buttonClass" :disabled="disabled" @click="confirming = true">
      {{ label }}
    </button>
  </span>
  <span v-else class="d-inline-flex align-items-center gap-2">
    <span class="text-danger fw-semibold" style="font-size: 0.85rem">{{ confirmText }}</span>
    <button class="btn btn-sm btn-danger" :disabled="running" @click="doConfirm">
      {{ running ? '...' : '确认' }}
    </button>
    <button class="btn btn-sm btn-outline-secondary" :disabled="running" @click="confirming = false">
      取消
    </button>
  </span>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  label: { type: String, default: '删除' },
  confirmText: { type: String, default: '确定要执行此操作？' },
  buttonClass: { type: String, default: 'btn btn-sm btn-outline-danger' },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['confirm'])
const confirming = ref(false)
const running = ref(false)

const doConfirm = async () => {
  running.value = true
  try {
    await new Promise((resolve, reject) => {
      emit('confirm', { resolve, reject })
    })
  } catch {
    // error handled by parent
  } finally {
    running.value = false
    confirming.value = false
  }
}
</script>
