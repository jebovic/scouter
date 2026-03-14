import { describe, it, expect } from 'vitest'
import type { PoolHealth } from '../api/health'

// Import the internal helper by re-exporting it from the module under test.
// Since deriveOverallStatus is not exported, we test it via the hook's
// overallStatus output. We also directly test the logic by extracting it.
// The hook file exports deriveOverallStatus for testability.
import { deriveOverallStatus } from './useLLMHealth'

const makePool = (models: Array<{ healthy: boolean }>): PoolHealth => ({
  models: models.map((m, i) => ({
    name: `model-${i}`,
    healthy: m.healthy,
    circuit_state: m.healthy ? 'closed' : 'open',
    last_latency_ms: 100,
    capabilities: [],
  })),
})

describe('deriveOverallStatus', () => {
  it('returns "loading" when isLoading is true', () => {
    expect(deriveOverallStatus(undefined, true, false)).toBe('loading')
  })

  it('returns "error" when isError is true', () => {
    expect(deriveOverallStatus(undefined, false, true)).toBe('error')
  })

  it('returns "error" when data is undefined and not loading', () => {
    expect(deriveOverallStatus(undefined, false, false)).toBe('error')
  })

  it('returns "error" when models array is empty', () => {
    expect(deriveOverallStatus(makePool([]), false, false)).toBe('error')
  })

  it('returns "healthy" when all models are healthy', () => {
    const pool = makePool([{ healthy: true }, { healthy: true }])
    expect(deriveOverallStatus(pool, false, false)).toBe('healthy')
  })

  it('returns "healthy" when single model is healthy', () => {
    const pool = makePool([{ healthy: true }])
    expect(deriveOverallStatus(pool, false, false)).toBe('healthy')
  })

  it('returns "down" when no models are healthy', () => {
    const pool = makePool([{ healthy: false }, { healthy: false }])
    expect(deriveOverallStatus(pool, false, false)).toBe('down')
  })

  it('returns "down" when single model is unhealthy', () => {
    const pool = makePool([{ healthy: false }])
    expect(deriveOverallStatus(pool, false, false)).toBe('down')
  })

  it('returns "degraded" when some but not all models are healthy', () => {
    const pool = makePool([{ healthy: true }, { healthy: false }])
    expect(deriveOverallStatus(pool, false, false)).toBe('degraded')
  })

  it('returns "degraded" when one of three models is healthy', () => {
    const pool = makePool([{ healthy: true }, { healthy: false }, { healthy: false }])
    expect(deriveOverallStatus(pool, false, false)).toBe('degraded')
  })

  it('loading takes precedence over error', () => {
    expect(deriveOverallStatus(undefined, true, true)).toBe('loading')
  })
})
