<template>
  <article class="kpi-card">
    <div class="kpi-head">
      <div class="kpi-icon" :style="{ color: color }">
        <i :class="icon"></i>
      </div>
      <div>
        <div class="tiny-label">{{ label }}</div>
        <div class="kpi-value">{{ value }}</div>
      </div>
    </div>
    <p class="kpi-note">{{ note }}</p>
    <TrendChart compact :height="76" :series="normalizedSeries" :baseline="baseline" />
  </article>
</template>

<script setup>
import { computed } from 'vue'
import TrendChart from './TrendChart.vue'

const props = defineProps({
  icon: { type: String, required: true },
  label: { type: String, required: true },
  value: { type: String, required: true },
  note: { type: String, default: '' },
  series: { type: Array, default: () => [] },
  color: { type: String, default: '#2563eb' },
  baseline: { type: String, default: 'zero' },
})

const normalizedSeries = computed(() =>
  props.series.map((serie, index) => ({
    fill: props.color,
    color: props.color,
    opacity: index === 0 ? 0.15 : 0.06,
    ...serie,
  })),
)
</script>

<style scoped>
.kpi-card {
  display: grid;
  gap: 0.625rem;
  padding: 1rem;
  border-radius: var(--onyx-radius);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
  box-shadow: var(--onyx-shadow-sm);
}

.kpi-head {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.kpi-icon {
  display: inline-grid;
  place-items: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--onyx-radius-xs);
  font-size: 1rem;
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
}

.kpi-value {
  margin-top: 0.125rem;
  font-size: 1.375rem;
  font-weight: 700;
  line-height: 1.1;
}

.kpi-note {
  margin: 0;
  color: var(--onyx-muted);
  font-size: 0.75rem;
  line-height: 1.4;
}
</style>
