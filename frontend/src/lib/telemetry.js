export const telemetryWindows = [
  { key: '1h', label: '1H' },
  { key: '6h', label: '6H' },
  { key: '24h', label: '24H' },
  { key: '7d', label: '7D' },
  { key: '14d', label: '14D' },
]

export function seriesForKey(telemetry, key) {
  return telemetry?.series?.[key] || []
}

export function buildSeries(telemetry, definitions) {
  return definitions.map((definition) => ({
    ...definition,
    points: seriesForKey(telemetry, definition.key),
  }))
}

export function downsampleSeries(points, maxPoints = 240) {
  if (!Array.isArray(points) || points.length <= maxPoints) {
    return points || []
  }

  const bucketSize = Math.ceil(points.length / maxPoints)
  const result = []

  for (let index = 0; index < points.length; index += bucketSize) {
    const bucket = points.slice(index, index + bucketSize)
    if (!bucket.length) continue

    const first = bucket[0]
    const last = bucket[bucket.length - 1]
    let sum = 0
    let maxPoint = first
    let minPoint = first

    for (const point of bucket) {
      sum += point.value
      if (point.value > maxPoint.value) maxPoint = point
      if (point.value < minPoint.value) minPoint = point
    }

    result.push({
      timestamp: first.timestamp,
      value: sum / bucket.length,
    })

    if (maxPoint.timestamp !== first.timestamp && maxPoint.timestamp !== last.timestamp) {
      result.push(maxPoint)
    }
    if (
      minPoint.timestamp !== first.timestamp &&
      minPoint.timestamp !== last.timestamp &&
      minPoint.timestamp !== maxPoint.timestamp
    ) {
      result.push(minPoint)
    }

    if (last.timestamp !== first.timestamp) {
      result.push(last)
    }
  }

  return result.sort((left, right) => left.timestamp - right.timestamp)
}

export function formatBytes(bytes) {
  const value = Number(bytes || 0)
  if (value >= 1024 ** 4) return `${(value / 1024 ** 4).toFixed(2)} TiB`
  if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(2)} GiB`
  if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MiB`
  if (value >= 1024) return `${(value / 1024).toFixed(0)} KiB`
  return `${value.toFixed(0)} B`
}

export function formatBytesPerSec(value) {
  return `${formatBytes(value)}/s`
}

export function formatOpsPerSec(value) {
  const numeric = Number(value || 0)
  if (numeric >= 1000) return `${numeric.toFixed(0)}/s`
  if (numeric >= 100) return `${numeric.toFixed(1)}/s`
  return `${numeric.toFixed(2)}/s`
}

export function formatPercent(value, digits = 1) {
  const numeric = Number(value || 0)
  return `${numeric.toFixed(digits)}%`
}

export function formatRatio(value) {
  const numeric = Number(value || 0)
  if (!numeric) return '1.00x'
  return `${numeric.toFixed(2)}x`
}

export function formatNumber(value) {
  const numeric = Number(value || 0)
  if (numeric >= 1000000) return `${(numeric / 1000000).toFixed(1)}M`
  if (numeric >= 1000) return `${(numeric / 1000).toFixed(1)}K`
  return `${numeric.toFixed(0)}`
}

export function formatTimestamp(timestamp) {
  if (!timestamp) return '--'
  return new Date(timestamp * 1000).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatDateTime(value) {
  if (!value) return '--'
  return new Date(value).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatWindowLabel(windowKey) {
  switch (windowKey) {
    case '1h':
      return 'Last hour'
    case '6h':
      return 'Last 6 hours'
    case '24h':
      return 'Last 24 hours'
    case '7d':
      return 'Last 7 days'
    case '14d':
      return 'Last 14 days'
    default:
      return 'Last 24 hours'
  }
}

export function formatDurationNs(value) {
  if (value == null || !Number.isFinite(value) || value < 0) {
    return '-'
  }
  const ms = value / 1_000_000
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)} s`
  }
  if (ms >= 10) {
    return `${ms.toFixed(1)} ms`
  }
  if (ms >= 0.01) {
    return `${ms.toFixed(2)} ms`
  }
  // Sub-10µs: still show in ms for consistency (3 decimals) to satisfy the
  // dashboard's ms-only convention.
  return `${ms.toFixed(3)} ms`
}

export function formatDurationUs(value) {
  if (value == null || !Number.isFinite(value) || value < 0) {
    return '-'
  }
  return formatDurationNs(value * 1000)
}

export function formatByKind(kind, value) {
  switch (kind) {
    case 'bytes':
      return formatBytes(value)
    case 'bytesRate':
      return formatBytesPerSec(value)
    case 'opsRate':
      return formatOpsPerSec(value)
    case 'percent':
      return formatPercent(value)
    case 'ratio':
      return formatRatio(value)
    case 'duration':
      return formatDurationNs(value)
    case 'durationUs':
      return formatDurationUs(value)
    case 'number':
    default:
      return formatNumber(value)
  }
}
