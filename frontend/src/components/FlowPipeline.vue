<template>
  <section class="flow-panel">
    <div class="flow-header">
      <div>
        <p class="eyebrow">Live engine flow</p>
        <h3 class="flow-title">From client IO to durable placement</h3>
      </div>
      <div class="flow-badges">
        <span class="badge text-bg-dark">{{ windowLabel }}</span>
        <span class="badge text-bg-dark">Mode {{ snapshot?.engineMode || 'unknown' }}</span>
      </div>
    </div>

    <div class="flow-throughput">
      <div class="ft-corner"></div>
      <div class="ft-col-head">Client</div>
      <div class="ft-col-head">Buffer</div>
      <div class="ft-col-head">LV3</div>

      <div class="ft-row-head"><i class="bi bi-arrow-down"></i> Write</div>
      <div class="ft-cell"><strong>{{ formatBytesPerSec(rates?.clientWriteBps) }}</strong></div>
      <div class="ft-cell"><strong>{{ formatBytesPerSec(rates?.bufferWriteBps) }}</strong></div>
      <div class="ft-cell"><strong>{{ formatBytesPerSec(rates?.lv3WriteBps) }}</strong></div>

      <div class="ft-row-head"><i class="bi bi-arrow-up"></i> Read</div>
      <div class="ft-cell"><strong>{{ formatBytesPerSec(rates?.clientReadBps) }}</strong></div>
      <div class="ft-cell"><strong>{{ formatBytesPerSec(rates?.bufferReadBps) }}</strong></div>
      <div class="ft-cell"><strong>{{ formatBytesPerSec(rates?.lv3ReadBps) }}</strong></div>
    </div>

    <div class="pipeline-strip">
      <template v-for="(stage, index) in stages" :key="stage.title">
        <article class="flow-node">
          <div class="flow-node-top">
            <div class="flow-node-icon" :style="{ color: stage.color }">
              <i :class="stage.icon"></i>
            </div>
            <div>
              <div class="tiny-label">{{ stage.kicker }}</div>
              <h4>{{ stage.title }}</h4>
            </div>
          </div>
          <div class="flow-node-value">{{ stage.value }}</div>
          <p>{{ stage.note }}</p>
        </article>

        <div v-if="links[index]" class="flow-link">
          <i class="bi bi-chevron-right"></i>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import {
  formatBytes,
  formatBytesPerSec,
  formatOpsPerSec,
  formatPercent,
  formatRatio,
} from '../lib/telemetry'

const props = defineProps({
  snapshot: { type: Object, default: null },
  rates: { type: Object, default: null },
  windowLabel: { type: String, default: '1 minute cadence' },
})

const stages = computed(() => [
  {
    kicker: 'Ingress',
    title: 'Client IO',
    icon: 'bi bi-arrow-down-up',
    color: '#2563eb',
    value: `${formatBytesPerSec(props.rates?.clientWriteBps)} / ${formatBytesPerSec(props.rates?.clientReadBps)}`,
    note: `${formatOpsPerSec(props.rates?.clientWriteIops)} write, ${formatOpsPerSec(props.rates?.clientReadIops)} read`,
  },
  {
    kicker: 'Routing',
    title: 'Zone Workers',
    icon: 'bi bi-diagram-3',
    color: '#3b82f6',
    value: `${props.snapshot?.zoneCount ?? 0} workers`,
    note: `${props.snapshot?.liveHandleCount ?? 0} live handles`,
  },
  {
    kicker: 'Hot path',
    title: 'Buffer Pool',
    icon: 'bi bi-layers',
    color: '#0d9488',
    value: formatPercent(props.snapshot?.bufferFillPercent, 0),
    note: `${formatBytes(props.snapshot?.bufferPayloadBytes)} resident, ${props.snapshot?.bufferPendingEntries ?? 0} pending`,
  },
  {
    kicker: 'Reduction',
    title: 'Dedup + Compress',
    icon: 'bi bi-magic',
    color: '#d97706',
    value: `${formatPercent(props.snapshot?.dedupHitRatePct)} hit / ${formatRatio(props.snapshot?.compressionRatio)}`,
    note: `${formatRatio(props.snapshot?.dataReductionRatio)} total reduction`,
  },
  {
    kicker: 'Durability',
    title: 'Pack + LV3',
    icon: 'bi bi-hdd-network',
    color: '#0891b2',
    value: `${formatBytesPerSec(props.rates?.lv3WriteBps)} write`,
    note: `${formatBytesPerSec(props.rates?.lv3ReadBps)} read, ${formatOpsPerSec(props.rates?.lv3WriteIops)} write IOPS`,
  },
  {
    kicker: 'Placement',
    title: 'Allocator',
    icon: 'bi bi-bounding-box-circles',
    color: '#6366f1',
    value: formatPercent(props.snapshot?.allocatorUsagePercent),
    note: `${props.snapshot?.volumeCount ?? 0} volumes`,
  },
])

const links = computed(() => [
  { label: 'pressure' },
  { label: 'fan-out' },
  { label: 'absorb' },
  { label: 'optimize' },
  { label: 'emit' },
])
</script>

<style scoped>
.flow-panel {
  display: grid;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: var(--onyx-radius);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
  box-shadow: var(--onyx-shadow-sm);
}

.flow-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.flow-title {
  margin: 0.125rem 0 0;
  font-size: 1.125rem;
  font-weight: 700;
}

.flow-badges {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.flow-throughput {
  display: grid;
  grid-template-columns: auto 1fr 1fr 1fr;
  gap: 0;
  border: 1px solid var(--onyx-border);
  border-radius: var(--onyx-radius-sm);
  overflow: hidden;
}

.ft-corner {
  background: var(--onyx-surface-soft);
  border-bottom: 1px solid var(--onyx-border);
  border-right: 1px solid var(--onyx-border);
}

.ft-col-head {
  padding: 0.5rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--onyx-muted);
  background: var(--onyx-surface-soft);
  border-bottom: 1px solid var(--onyx-border);
  border-right: 1px solid var(--onyx-border);
  text-align: center;
}

.ft-col-head:last-child {
  border-right: none;
}

.ft-row-head {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--onyx-text-secondary);
  background: var(--onyx-surface-soft);
  border-right: 1px solid var(--onyx-border);
  border-bottom: 1px solid var(--onyx-border);
  white-space: nowrap;
}

.ft-row-head i {
  font-size: 0.75rem;
  color: var(--onyx-muted);
}

.flow-throughput > .ft-row-head:nth-last-child(-n+4) {
  border-bottom: none;
}

.ft-cell {
  padding: 0.5rem 0.75rem;
  text-align: center;
  border-right: 1px solid var(--onyx-border);
  border-bottom: 1px solid var(--onyx-border);
}

.ft-cell:nth-child(4n) {
  border-right: none;
}

.flow-throughput > :nth-last-child(-n+4).ft-cell {
  border-bottom: none;
}

.ft-cell strong {
  font-size: 0.9375rem;
  font-weight: 700;
}

.pipeline-strip {
  display: flex;
  align-items: stretch;
  gap: 0;
  flex-wrap: wrap;
}

.flow-node {
  flex: 1 1 160px;
  min-width: 0;
  display: grid;
  gap: 0.5rem;
  padding: 0.875rem;
  border-radius: var(--onyx-radius-sm);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
}

.flow-node-top {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.flow-node-icon {
  width: 2.25rem;
  height: 2.25rem;
  display: grid;
  place-items: center;
  border-radius: var(--onyx-radius-xs);
  font-size: 1rem;
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
}

.flow-node h4 {
  margin: 0.125rem 0 0;
  font-size: 0.875rem;
  font-weight: 600;
}

.flow-node-value {
  font-size: 1.125rem;
  font-weight: 700;
  line-height: 1.2;
}

.flow-node p {
  margin: 0;
  color: var(--onyx-muted);
  font-size: 0.75rem;
  line-height: 1.4;
}

.flow-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  flex-shrink: 0;
  color: var(--onyx-muted);
  font-size: 0.75rem;
}

@media (max-width: 980px) {
  .flow-link {
    width: 100%;
    height: 1.25rem;
    flex-basis: 100%;
  }

  .flow-link i::before {
    content: '\F282';
  }
}
</style>
