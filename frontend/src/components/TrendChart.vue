<template>
  <div ref="chartRef" class="trend-chart" :class="{ compact }" :style="{ height: `${height}px` }">
    <div v-if="!compact && hasData" class="trend-scale">
      <span>{{ maxLabel }}</span>
      <span>{{ minLabel }}</span>
    </div>

    <svg
      v-if="hasData"
      class="trend-svg"
      :viewBox="`0 0 ${svgWidth} ${height}`"
      @mousemove="onMouseMove"
      @mouseleave="onMouseLeave"
    >
      <g v-if="!compact" class="trend-grid">
        <line
          v-for="tick in ticks"
          :key="tick.key"
          x1="0"
          :x2="svgWidth"
          :y1="tick.y"
          :y2="tick.y"
        />
      </g>

      <template v-for="serie in normalizedSeries" :key="serie.key">
        <path
          v-if="!compact && serie.areaPath"
          class="trend-area"
          :d="serie.areaPath"
          :fill="serie.fill || serie.color"
          :opacity="serie.opacity ?? 0.12"
        />
        <path
          class="trend-line"
          :d="serie.path"
          :stroke="serie.color"
          :stroke-width="compact ? 1.8 : 2"
        />
        <circle
          v-if="!hover && serie.lastPoint"
          class="trend-dot"
          :cx="serie.lastPoint.x"
          :cy="serie.lastPoint.y"
          :fill="serie.color"
          :r="compact ? 3 : 4"
        />
      </template>

      <!-- Hover crosshair + dots -->
      <template v-if="hover && !compact">
        <line
          class="trend-crosshair"
          :x1="hover.x"
          :x2="hover.x"
          :y1="chartPadY"
          :y2="height - chartPadY"
        />
        <circle
          v-for="snap in hover.snaps"
          :key="snap.key"
          class="trend-snap-dot"
          :cx="snap.cx"
          :cy="snap.cy"
          :fill="snap.color"
          r="4"
        />
      </template>
    </svg>

    <!-- Tooltip (HTML overlay, outside SVG) -->
    <div
      v-if="hover && !compact"
      class="trend-tooltip"
      :style="tooltipStyle"
    >
      <div class="tt-time">{{ hover.timeLabel }}</div>
      <div v-for="snap in hover.snaps" :key="snap.key" class="tt-row">
        <i class="tt-dot" :style="{ background: snap.color }"></i>
        <span class="tt-label">{{ snap.label }}</span>
        <strong class="tt-value">{{ snap.formatted }}</strong>
      </div>
    </div>

    <div v-if="!hasData" class="trend-empty">Waiting for samples</div>

    <div v-if="!compact && hasData" class="trend-footer">
      <span>{{ startLabel }}</span>
      <span>{{ endLabel }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { downsampleSeries, formatByKind, formatTimestamp } from '../lib/telemetry'

const props = defineProps({
  series: { type: Array, default: () => [] },
  height: { type: Number, default: 220 },
  compact: { type: Boolean, default: false },
  baseline: { type: String, default: 'zero' },
  format: { type: String, default: 'number' },
  maxPoints: { type: Number, default: 240 },
})

const chartRef = ref(null)
const svgWidth = ref(300)
const hover = ref(null)

let resizeObserver = null

onMounted(() => {
  if (chartRef.value) {
    svgWidth.value = chartRef.value.clientWidth || 300
    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        svgWidth.value = entry.contentRect.width || 300
      }
    })
    resizeObserver.observe(chartRef.value)
  }
})

onBeforeUnmount(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
})

const chartPadY = computed(() => (props.compact ? 4 : 8))
const chartPadX = computed(() => (props.compact ? 4 : 8))

const filteredSeries = computed(() =>
  props.series
    .map((serie) => ({
      ...serie,
      points: downsampleSeries(serie.points, props.compact ? Math.min(props.maxPoints, 72) : props.maxPoints),
    }))
    .filter((serie) => Array.isArray(serie.points) && serie.points.length > 0),
)

const allPoints = computed(() => filteredSeries.value.flatMap((serie) => serie.points))

const bounds = computed(() => {
  if (!allPoints.value.length) {
    return { minTs: 0, maxTs: 1, minValue: 0, maxValue: 1 }
  }

  const timestamps = allPoints.value.map((point) => point.timestamp)
  const values = allPoints.value.map((point) => point.value)
  const minTs = Math.min(...timestamps)
  const maxTs = Math.max(...timestamps)
  const rawMinValue = Math.min(...values)
  const minValue = props.baseline === 'fit' ? rawMinValue : Math.min(0, rawMinValue)
  const maxValue = Math.max(...values)

  return {
    minTs,
    maxTs: maxTs === minTs ? minTs + 1 : maxTs,
    minValue,
    maxValue: maxValue === minValue ? maxValue + 1 : maxValue,
  }
})

function toX(timestamp) {
  const usableWidth = svgWidth.value - chartPadX.value * 2
  return chartPadX.value + ((timestamp - bounds.value.minTs) / (bounds.value.maxTs - bounds.value.minTs)) * usableWidth
}

function toY(value) {
  const usableHeight = props.height - chartPadY.value * 2
  return props.height - chartPadY.value - ((value - bounds.value.minValue) / (bounds.value.maxValue - bounds.value.minValue)) * usableHeight
}

function fromX(pixelX) {
  const usableWidth = svgWidth.value - chartPadX.value * 2
  return bounds.value.minTs + ((pixelX - chartPadX.value) / usableWidth) * (bounds.value.maxTs - bounds.value.minTs)
}

function findNearest(points, targetTs) {
  if (!points.length) return null
  let best = points[0]
  let bestDist = Math.abs(best.timestamp - targetTs)
  for (let i = 1; i < points.length; i++) {
    const dist = Math.abs(points[i].timestamp - targetTs)
    if (dist < bestDist) {
      best = points[i]
      bestDist = dist
    } else {
      break
    }
  }
  return best
}

function onMouseMove(event) {
  if (props.compact) return
  const svg = event.currentTarget
  const rect = svg.getBoundingClientRect()
  const ratioX = (event.clientX - rect.left) / rect.width
  const svgX = ratioX * svgWidth.value
  const targetTs = fromX(svgX)

  const snaps = filteredSeries.value.map((serie) => {
    const pt = findNearest(serie.points, targetTs)
    return pt
      ? {
          key: serie.key,
          label: serie.label || serie.key,
          color: serie.color,
          value: pt.value,
          formatted: formatByKind(props.format, pt.value),
          cx: toX(pt.timestamp).toFixed(2),
          cy: toY(pt.value).toFixed(2),
          timestamp: pt.timestamp,
        }
      : null
  }).filter(Boolean)

  if (!snaps.length) {
    hover.value = null
    return
  }

  const snapX = Number(snaps[0].cx)
  hover.value = {
    x: snapX.toFixed(2),
    pixelRatio: snapX / svgWidth.value,
    timeLabel: formatTimestamp(snaps[0].timestamp),
    snaps,
  }
}

function onMouseLeave() {
  hover.value = null
}

const tooltipStyle = computed(() => {
  if (!hover.value) return { display: 'none' }
  const ratio = hover.value.pixelRatio
  const flipRight = ratio > 0.7
  return {
    top: `${chartPadY.value + 4}px`,
    ...(flipRight
      ? { right: `${Math.round((1 - ratio) * 100)}%` }
      : { left: `${Math.round(ratio * 100)}%` }),
  }
})

function buildLinePath(points) {
  return points
    .map((point, index) => `${index === 0 ? 'M' : 'L'} ${toX(point.timestamp).toFixed(2)} ${toY(point.value).toFixed(2)}`)
    .join(' ')
}

function buildAreaPath(points) {
  const firstX = toX(points[0].timestamp).toFixed(2)
  const lastX = toX(points[points.length - 1].timestamp).toFixed(2)
  const baseY = toY(bounds.value.minValue).toFixed(2)
  return `${buildLinePath(points)} L ${lastX} ${baseY} L ${firstX} ${baseY} Z`
}

const normalizedSeries = computed(() =>
  filteredSeries.value.map((serie) => {
    const path = buildLinePath(serie.points)
    const last = serie.points[serie.points.length - 1]
    return {
      ...serie,
      path,
      areaPath: serie.points.length > 1 ? buildAreaPath(serie.points) : '',
      lastPoint: last
        ? {
            x: toX(last.timestamp).toFixed(2),
            y: toY(last.value).toFixed(2),
          }
        : null,
    }
  }),
)

const hasData = computed(() => normalizedSeries.value.length > 0)

const ticks = computed(() => {
  if (props.compact) return []
  return Array.from({ length: 5 }, (_, index) => {
    const ratio = index / 4
    return {
      key: `tick-${index}`,
      y: chartPadY.value + (props.height - chartPadY.value * 2) * ratio,
    }
  })
})

const startLabel = computed(() => formatTimestamp(bounds.value.minTs))
const endLabel = computed(() => formatTimestamp(bounds.value.maxTs))
const maxLabel = computed(() => `High ${formatByKind(props.format, bounds.value.maxValue)}`)
const minLabel = computed(() => `Low ${formatByKind(props.format, bounds.value.minValue)}`)
</script>

<style scoped>
.trend-chart {
  position: relative;
  display: grid;
  gap: 0.55rem;
}

.trend-scale {
  display: flex;
  justify-content: space-between;
  color: var(--onyx-muted);
  font-size: 0.76rem;
}

.trend-svg {
  width: 100%;
  height: 100%;
  overflow: visible;
  cursor: crosshair;
}

.compact .trend-svg {
  cursor: default;
}

.trend-grid line {
  stroke: rgba(21, 55, 79, 0.1);
  stroke-dasharray: 4 5;
}

.trend-line {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.trend-area {
  transition: opacity 0.2s ease;
}

.trend-dot {
  filter: drop-shadow(0 0 6px rgba(23, 72, 120, 0.16));
}

.trend-crosshair {
  stroke: var(--onyx-muted, #64748b);
  stroke-width: 1;
  stroke-dasharray: 3 3;
  opacity: 0.5;
}

.trend-snap-dot {
  stroke: #fff;
  stroke-width: 2;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.12));
}

/* ── Tooltip ───────────────────────────────────────── */

.trend-tooltip {
  position: absolute;
  z-index: 10;
  pointer-events: none;
  padding: 0.5rem 0.625rem;
  border-radius: var(--onyx-radius-xs, 0.375rem);
  background: var(--onyx-surface, #fff);
  border: 1px solid var(--onyx-border, #e2e8f0);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  font-size: 0.75rem;
  line-height: 1.4;
  white-space: nowrap;
  min-width: 120px;
}

.tt-time {
  font-weight: 600;
  color: var(--onyx-text, #0f172a);
  margin-bottom: 0.25rem;
  padding-bottom: 0.25rem;
  border-bottom: 1px solid var(--onyx-border-light, #f1f5f9);
}

.tt-row {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.1rem 0;
}

.tt-dot {
  width: 0.4375rem;
  height: 0.4375rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.tt-label {
  color: var(--onyx-muted, #64748b);
  flex: 1;
}

.tt-value {
  font-weight: 600;
  color: var(--onyx-text, #0f172a);
}

.trend-footer {
  display: flex;
  justify-content: space-between;
  color: var(--onyx-muted);
  font-size: 0.78rem;
}

.trend-empty {
  display: grid;
  place-items: center;
  height: 100%;
  min-height: 120px;
  color: var(--onyx-muted);
  border-radius: var(--onyx-radius, 0.75rem);
  border: 1px dashed var(--onyx-border, #e2e8f0);
  background: var(--onyx-surface-soft, #f1f5f9);
}

.compact .trend-empty {
  min-height: 54px;
  font-size: 0.78rem;
}
</style>
