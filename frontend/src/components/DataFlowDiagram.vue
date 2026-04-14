<template>
  <section class="data-flow">
    <div class="section-header">
      <div>
        <h3>Data pipeline</h3>
        <p class="flow-note">Live throughput and IOPS across the write and read path.</p>
      </div>
      <span class="badge text-bg-dark">Mode {{ snapshot?.engineMode || 'unknown' }}</span>
    </div>

    <div class="pipeline-wrap">
      <div class="pipeline-layout" ref="layoutRef">

        <!-- Row 1: Client IO, centered -->
        <div class="top-row">
          <div class="stage stage-client" ref="clientRef">
            <div class="stage-head">
              <div class="stage-icon" style="color: #2563eb">
                <i class="bi bi-box-arrow-in-right"></i>
              </div>
              <div class="stage-title">Client IO</div>
            </div>
            <div class="stage-metrics">
              <div class="metric-pair">
                <span class="metric-dir metric-dir-w">W</span>
                <strong>{{ fmtRate(rates?.clientWriteBps) }}</strong>
                <span class="metric-iops">{{ fmtOps(rates?.clientWriteIops) }}</span>
              </div>
              <div class="metric-pair">
                <span class="metric-dir metric-dir-r">R</span>
                <strong>{{ fmtRate(rates?.clientReadBps) }}</strong>
                <span class="metric-iops">{{ fmtOps(rates?.clientReadIops) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: SVG diagonal connectors -->
        <div class="diag-row">
          <svg width="100%" :height="diagH" class="diag-svg">
            <!-- Write: Client → Buffer (blue, ↓) -->
            <line :x1="diagPts.clientL - bufNorm.nx" :y1="diagPad - bufNorm.ny"
              :x2="diagPts.buffer - bufNorm.nx" :y2="diagH - diagPad - bufNorm.ny"
              class="diag-line diag-write" :style="diagStroke(rates?.clientWriteBps)"
              vector-effect="non-scaling-stroke" />
            <!-- Read buffer: Buffer → Client (green, ↑) -->
            <line :x1="diagPts.buffer + bufNorm.nx" :y1="diagH - diagPad + bufNorm.ny"
              :x2="diagPts.clientL + bufNorm.nx" :y2="diagPad + bufNorm.ny"
              class="diag-line diag-read" :style="diagStroke(rates?.bufferReadBps)"
              vector-effect="non-scaling-stroke" />
            <!-- Read LV3: LV3 → Client (green, ↑) -->
            <line :x1="diagPts.lv3" :y1="diagH - diagPad" :x2="diagPts.clientR" :y2="diagPad"
              class="diag-line diag-read" :style="diagStroke(rates?.lv3ReadBps)"
              vector-effect="non-scaling-stroke" />
            <!-- Labels: write on right side of left lines, read on left side of right line -->
            <text :x="(diagPts.clientL + diagPts.buffer) / 2 + 16" :y="diagH / 2 + 4"
              class="diag-label diag-label-w" text-anchor="start">W {{ fmtRate(rates?.clientWriteBps) }}</text>
            <text :x="(diagPts.clientR + diagPts.lv3) / 2 - 16" :y="diagH / 2 + 4"
              class="diag-label diag-label-r" text-anchor="end">R {{ fmtRate(rates?.lv3ReadBps) }}</text>
          </svg>
        </div>

        <!-- Row 3: stages in a row -->
        <div class="stages-row">
          <!-- Buffer -->
          <div class="stage stage-buffer" ref="bufferRef">
            <div class="stage-head">
              <div class="buffer-gauge">
                <div class="buffer-gauge-fill" :class="bufferFillClass" :style="{ height: `${bufferFillPct}%` }"></div>
              </div>
              <div>
                <div class="stage-title">Buffer Pool</div>
                <div class="stage-sub">{{ bufferFillPct }}% full</div>
              </div>
            </div>
            <div class="stage-metrics">
              <div class="metric-pair">
                <span class="metric-dir metric-dir-w">W</span>
                <strong>{{ fmtRate(rates?.bufferWriteBps) }}</strong>
                <span class="metric-iops">{{ fmtOps(rates?.bufferWriteIops) }}</span>
              </div>
              <div class="metric-pair">
                <span class="metric-dir metric-dir-r">R</span>
                <strong>{{ fmtRate(rates?.bufferReadBps) }}</strong>
                <span class="metric-iops">{{ fmtOps(rates?.bufferReadIops) }}</span>
              </div>
            </div>
            <div class="buffer-detail">
              <span>{{ fmtBytes(snapshot?.bufferPayloadBytes) }} / {{ fmtBytes(snapshot?.bufferPayloadLimit) }}</span>
              <span>{{ snapshot?.bufferPendingEntries ?? 0 }} pending</span>
            </div>
          </div>

          <!-- → -->
          <div class="h-connector">
            <div class="arrow-shaft shaft-write" :style="shaftH(rates?.bufferWriteBps)"></div>
            <div class="arrow-tip tip-right tip-write" :style="tipSizeH(rates?.bufferWriteBps)"></div>
          </div>

          <!-- Dedup -->
          <div class="stage stage-compact">
            <div class="stage-head">
              <div class="stage-icon" style="color: #d97706">
                <i class="bi bi-intersect"></i>
              </div>
              <div class="stage-title">Dedup</div>
            </div>
            <div class="stage-metrics">
              <div class="metric-line">
                <span class="metric-label">Hit rate</span>
                <strong>{{ fmtPct(snapshot?.dedupHitRatePct) }}</strong>
              </div>
            </div>
            <div class="stage-badge badge-eliminated" v-if="dedupHitPct > 0">
              <i class="bi bi-dash-circle"></i> {{ dedupHitPct }}% eliminated
            </div>
          </div>

          <!-- → -->
          <div class="h-connector">
            <div class="arrow-shaft shaft-write" :style="shaftH(postDedupBps)"></div>
            <div class="arrow-tip tip-right tip-write" :style="tipSizeH(postDedupBps)"></div>
          </div>

          <!-- Compress -->
          <div class="stage stage-compact">
            <div class="stage-head">
              <div class="stage-icon" style="color: #0d9488">
                <i class="bi bi-arrows-collapse"></i>
              </div>
              <div class="stage-title">Compress</div>
            </div>
            <div class="stage-metrics">
              <div class="metric-line">
                <span class="metric-label">Ratio</span>
                <strong>{{ fmtRatio(snapshot?.compressionRatio) }}</strong>
              </div>
            </div>
          </div>

          <!-- → -->
          <div class="h-connector">
            <div class="arrow-shaft shaft-write" :style="shaftH(rates?.lv3WriteBps)"></div>
            <div class="arrow-tip tip-right tip-write" :style="tipSizeH(rates?.lv3WriteBps)"></div>
          </div>

          <!-- LV3 -->
          <div class="stage" ref="lv3Ref">
            <div class="stage-head">
              <div class="stage-icon" style="color: #0891b2">
                <i class="bi bi-hdd-network-fill"></i>
              </div>
              <div class="stage-title">LV3</div>
            </div>
            <div class="stage-metrics">
              <div class="metric-pair">
                <span class="metric-dir metric-dir-w">W</span>
                <strong>{{ fmtRate(rates?.lv3WriteBps) }}</strong>
                <span class="metric-iops">{{ fmtOps(rates?.lv3WriteIops) }}</span>
              </div>
              <div class="metric-pair">
                <span class="metric-dir metric-dir-r">R</span>
                <strong>{{ fmtRate(rates?.lv3ReadBps) }}</strong>
                <span class="metric-iops">{{ fmtOps(rates?.lv3ReadIops) }}</span>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>

    <!-- Reduction strip -->
    <div class="reduction-strip" v-if="dataReductionRatio > 0">
      <div class="reduction-bar-track">
        <div class="reduction-bar-saved" :style="{ width: `${savedPct}%` }"></div>
      </div>
      <div class="reduction-labels">
        <span>Data reduction <strong>{{ fmtRatio(snapshot?.dataReductionRatio) }}</strong></span>
        <span class="reduction-saved">{{ savedPct }}% space saved</span>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
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
})

const fmtRate = (v) => formatBytesPerSec(v)
const fmtOps = (v) => formatOpsPerSec(v)
const fmtBytes = (v) => formatBytes(v)
const fmtPct = (v) => formatPercent(v)
const fmtRatio = (v) => formatRatio(v)

/* ── element refs for measuring diagonal line positions ── */
const layoutRef = ref(null)
const clientRef = ref(null)
const bufferRef = ref(null)
const lv3Ref = ref(null)

const diagH = 56
const diagPad = 6

const diagPts = reactive({ clientL: 200, clientR: 300, buffer: 80, lv3: 520 })

const measure = () => {
  const layout = layoutRef.value
  const client = clientRef.value
  const buffer = bufferRef.value
  const lv3 = lv3Ref.value
  if (!layout || !client || !buffer || !lv3) return

  const base = layout.getBoundingClientRect().left
  const cr = client.getBoundingClientRect()
  const br = buffer.getBoundingClientRect()
  const lr = lv3.getBoundingClientRect()

  diagPts.clientL = cr.left - base
  diagPts.clientR = cr.left - base + cr.width
  diagPts.buffer = br.left - base + br.width
  diagPts.lv3 = lr.left - base
}

let ro = null
onMounted(() => {
  measure()
  ro = new ResizeObserver(measure)
  if (layoutRef.value) ro.observe(layoutRef.value)
})
onBeforeUnmount(() => { ro?.disconnect() })

/* ── computed metrics ── */

const bufferFillPct = computed(() => {
  const raw = props.snapshot?.bufferFillPercent ?? 0
  return Math.round(Math.max(0, Math.min(100, raw)))
})

const bufferFillClass = computed(() => {
  const p = bufferFillPct.value
  if (p >= 90) return 'fill-danger'
  if (p >= 70) return 'fill-warning'
  return 'fill-ok'
})

const dedupHitPct = computed(() => Math.round(props.snapshot?.dedupHitRatePct ?? 0))

const dataReductionRatio = computed(() => props.snapshot?.dataReductionRatio ?? 0)
const savedPct = computed(() => {
  const r = dataReductionRatio.value
  if (r <= 1) return 0
  return Math.round((1 - 1 / r) * 100)
})

const postDedupBps = computed(() => {
  const bw = props.rates?.bufferWriteBps ?? 0
  const hitRate = (props.snapshot?.dedupHitRatePct ?? 0) / 100
  return bw * (1 - hitRate)
})

const maxBps = computed(() => {
  return Math.max(
    props.rates?.clientWriteBps ?? 0,
    props.rates?.clientReadBps ?? 0,
    props.rates?.bufferWriteBps ?? 0,
    props.rates?.lv3WriteBps ?? 0,
    1,
  )
})

const shaftH = (bps) => {
  const r = Math.max(0.15, Math.min(1, (bps ?? 0) / maxBps.value))
  return { height: `${2 + r * 8}px` }
}

const tipSizeH = (bps) => {
  const r = Math.max(0.15, Math.min(1, (bps ?? 0) / maxBps.value))
  return { '--tip-half': `${4 + r * 6}px`, '--tip-len': `${6 + r * 8}px` }
}

/* perpendicular offset so write/read lines don't overlap */
const bufNorm = computed(() => {
  const dx = diagPts.buffer - diagPts.clientL
  const dy = diagH
  const len = Math.sqrt(dx * dx + dy * dy) || 1
  const sw = 2 + Math.max(0.15, Math.min(1, (props.rates?.clientWriteBps ?? 0) / maxBps.value)) * 6
  const sr = 2 + Math.max(0.15, Math.min(1, (props.rates?.bufferReadBps ?? 0) / maxBps.value)) * 6
  const gap = sw / 2 + 2 + sr / 2
  return { nx: (-dy / len) * gap, ny: (dx / len) * gap }
})

const diagStroke = (bps) => {
  const r = Math.max(0.15, Math.min(1, (bps ?? 0) / maxBps.value))
  return { strokeWidth: 2 + r * 6 }
}
</script>

<style scoped>
.data-flow {
  display: grid;
  gap: 1.25rem;
  padding: 1.5rem;
  border-radius: var(--onyx-radius);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
  box-shadow: var(--onyx-shadow-sm);
}

.flow-note {
  margin: 0.125rem 0 0;
  color: var(--onyx-muted);
  font-size: 0.8125rem;
}

/* ─── Layout ────────────────────────────────────── */

.pipeline-wrap { overflow-x: auto; }

.pipeline-layout {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  min-width: 560px;
}

.top-row {
  display: flex;
  justify-content: center;
}

.diag-row {
  position: relative;
}

.stages-row {
  display: flex;
  align-items: stretch;
  gap: 0;
}

/* ─── SVG diagonal connectors ───────────────────── */

.diag-svg {
  display: block;
  overflow: visible;
}

.diag-line {
  fill: none;
  stroke-linecap: round;
}

.diag-write {
  stroke: #3b82f6;
  stroke-dasharray: 8 5;
  animation: dash-flow 0.6s linear infinite;
}

.diag-read {
  stroke: #10b981;
  stroke-dasharray: 8 5;
  animation: dash-flow 0.6s linear infinite;
}

@keyframes dash-flow { to { stroke-dashoffset: -13; } }

.diag-label {
  font-size: 10px;
  font-weight: 700;
  fill: var(--onyx-muted);
}

.diag-label-w { fill: #3b82f6; }
.diag-label-r { fill: #10b981; }

/* ─── Stage nodes ───────────────────────────────── */

.stage {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  border-radius: var(--onyx-radius-sm);
  border: 1px solid var(--onyx-border);
  background: var(--onyx-surface);
  flex: 1 1 0;
  min-width: 0;
}

.stage-client {
  flex: 0 0 auto;
  min-width: 0;
}

.stage-compact {
  padding: 0.5rem 0.625rem;
  gap: 0.25rem;
  flex: 0 0 auto;
  align-self: center;
  min-width: 5.5rem;
  min-height: 5.5rem;
}

.stage-compact .stage-icon {
  width: 1.5rem;
  height: 1.5rem;
  font-size: 0.7rem;
}

.stage-compact .stage-title {
  font-size: 0.7rem;
}

.stage-compact .stage-head {
  gap: 0.375rem;
}

.stage-compact .metric-line strong {
  font-size: 0.9rem;
}

.stage-compact .stage-badge {
  font-size: 0.5625rem;
  padding: 0.0625rem 0.375rem;
}

.stage-buffer {
  border: 2px solid var(--onyx-accent);
  background: linear-gradient(135deg, rgba(13, 148, 136, 0.04), rgba(13, 148, 136, 0.01));
}

.stage-head { display: flex; align-items: center; gap: 0.625rem; }

.stage-icon {
  width: 2rem; height: 2rem;
  display: grid; place-items: center;
  border-radius: var(--onyx-radius-xs);
  font-size: 0.875rem;
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
  flex-shrink: 0;
}

.stage-title { font-size: 0.8125rem; font-weight: 700; }
.stage-sub { font-size: 0.6875rem; color: var(--onyx-muted); }

/* ─── Metrics ───────────────────────────────────── */

.stage-metrics { display: flex; flex-direction: column; gap: 0.25rem; }

.metric-pair {
  display: flex; align-items: baseline; gap: 0.375rem;
  font-size: 0.8125rem; line-height: 1.3;
}
.metric-pair strong { font-size: 0.875rem; font-weight: 700; }

.metric-dir {
  display: inline-flex; align-items: center; justify-content: center;
  width: 1.125rem; height: 1.125rem; border-radius: 3px;
  font-size: 0.5625rem; font-weight: 800; flex-shrink: 0;
}
.metric-dir-w { color: #1d4ed8; background: #dbeafe; }
.metric-dir-r { color: #065f46; background: #d1fae5; }

.metric-iops { color: var(--onyx-muted); font-size: 0.6875rem; white-space: nowrap; }

.metric-line { display: flex; align-items: baseline; gap: 0.375rem; font-size: 0.8125rem; }
.metric-line strong { font-size: 1.125rem; font-weight: 700; }
.metric-label { color: var(--onyx-muted); font-size: 0.6875rem; font-weight: 600; }

/* ─── Buffer ────────────────────────────────────── */

.buffer-detail {
  display: flex; flex-direction: column; gap: 0;
  font-size: 0.6875rem; color: var(--onyx-muted);
  border-top: 1px solid var(--onyx-border); padding-top: 0.375rem;
}

.buffer-gauge {
  width: 16px; height: 42px; border-radius: 3px;
  border: 1.5px solid var(--onyx-border);
  background: var(--onyx-surface-soft);
  position: relative; overflow: hidden; flex-shrink: 0;
}
.buffer-gauge-fill {
  position: absolute; bottom: 0; left: 0; right: 0;
  border-radius: 0 0 2px 2px; transition: height 0.6s ease;
}

.fill-ok { background: var(--onyx-accent); }
.fill-warning { background: var(--onyx-warm); }
.fill-danger { background: var(--onyx-danger); }

.stage-badge {
  display: inline-flex; align-items: center; gap: 0.25rem;
  padding: 0.125rem 0.5rem; border-radius: 999px;
  font-size: 0.625rem; font-weight: 600; white-space: nowrap; align-self: flex-start;
}
.badge-eliminated { color: #92400e; background: #fef3c7; }

/* ─── Horizontal arrows ─────────────────────────── */

.h-connector {
  display: flex; align-items: center;
  padding: 0 3px;
  min-width: 24px;
  flex: 0 0 auto;
  align-self: center;
}

.arrow-shaft {
  flex: 1; min-width: 16px;
  border-radius: 1px; transition: height 0.4s ease;
}

.shaft-write {
  background: repeating-linear-gradient(90deg, #3b82f6 0 6px, #93bbfd 6px 10px);
  background-size: 10px 100%;
  animation: flow-right 0.4s linear infinite;
}

@keyframes flow-right { to { background-position: 10px 0; } }

.arrow-tip { width: 0; height: 0; flex-shrink: 0; }

.tip-right {
  border-top: var(--tip-half, 6px) solid transparent;
  border-bottom: var(--tip-half, 6px) solid transparent;
  border-left: var(--tip-len, 9px) solid currentColor;
}
.tip-write { color: #3b82f6; }

/* ─── Reduction strip ───────────────────────────── */

.reduction-strip {
  display: grid; gap: 0.375rem;
  padding: 0.75rem 1rem;
  border-radius: var(--onyx-radius-sm);
  background: var(--onyx-surface-soft);
  border: 1px solid var(--onyx-border);
}
.reduction-bar-track {
  height: 6px; border-radius: 3px;
  background: var(--onyx-border); overflow: hidden;
}
.reduction-bar-saved {
  height: 100%; border-radius: 3px;
  background: linear-gradient(90deg, var(--onyx-accent), #0d9488cc);
  transition: width 0.6s ease;
}
.reduction-labels {
  display: flex; justify-content: space-between;
  font-size: 0.75rem; color: var(--onyx-muted);
}
.reduction-saved { font-weight: 600; color: var(--onyx-accent); }

/* ─── Responsive ────────────────────────────────── */

@media (max-width: 700px) {
  .pipeline-layout { min-width: 0; }
  .stages-row { flex-wrap: wrap; gap: 0.5rem; }
  .stage { flex: 1 1 120px; }
  .h-connector { display: none; }
  .diag-row { display: none; }
}
</style>
