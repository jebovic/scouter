import { describe, it, expect } from 'vitest'
import { ModelStatusSchema, PoolHealthSchema } from './health'

describe('ModelStatusSchema', () => {
  it('validates a correct model status', () => {
    const valid = {
      name: 'qwen3:14b',
      healthy: true,
      circuit_state: 'closed',
      last_latency_ms: 342,
      capabilities: ['tool_use', 'streaming'],
    }
    expect(() => ModelStatusSchema.parse(valid)).not.toThrow()
    const result = ModelStatusSchema.parse(valid)
    expect(result.name).toBe('qwen3:14b')
    expect(result.healthy).toBe(true)
    expect(result.circuit_state).toBe('closed')
    expect(result.last_latency_ms).toBe(342)
    expect(result.capabilities).toEqual(['tool_use', 'streaming'])
  })

  it('validates all circuit_state values', () => {
    const base = {
      name: 'model',
      healthy: false,
      last_latency_ms: 0,
      capabilities: [],
    }
    expect(() => ModelStatusSchema.parse({ ...base, circuit_state: 'closed' })).not.toThrow()
    expect(() => ModelStatusSchema.parse({ ...base, circuit_state: 'open' })).not.toThrow()
    expect(() => ModelStatusSchema.parse({ ...base, circuit_state: 'half_open' })).not.toThrow()
  })

  it('rejects invalid circuit_state', () => {
    const invalid = {
      name: 'model',
      healthy: true,
      circuit_state: 'unknown',
      last_latency_ms: 100,
      capabilities: [],
    }
    expect(() => ModelStatusSchema.parse(invalid)).toThrow()
  })

  it('rejects missing name', () => {
    const invalid = {
      healthy: true,
      circuit_state: 'closed',
      last_latency_ms: 100,
      capabilities: [],
    }
    expect(() => ModelStatusSchema.parse(invalid)).toThrow()
  })

  it('rejects non-boolean healthy', () => {
    const invalid = {
      name: 'model',
      healthy: 'yes',
      circuit_state: 'closed',
      last_latency_ms: 100,
      capabilities: [],
    }
    expect(() => ModelStatusSchema.parse(invalid)).toThrow()
  })

  it('rejects non-numeric last_latency_ms', () => {
    const invalid = {
      name: 'model',
      healthy: true,
      circuit_state: 'closed',
      last_latency_ms: 'fast',
      capabilities: [],
    }
    expect(() => ModelStatusSchema.parse(invalid)).toThrow()
  })
})

describe('PoolHealthSchema', () => {
  it('validates a correct payload with multiple models', () => {
    const valid = {
      models: [
        {
          name: 'qwen3:14b',
          healthy: true,
          circuit_state: 'closed',
          last_latency_ms: 200,
          capabilities: ['tool_use'],
        },
        {
          name: 'qwen3:4b',
          healthy: false,
          circuit_state: 'open',
          last_latency_ms: 0,
          capabilities: [],
        },
      ],
    }
    expect(() => PoolHealthSchema.parse(valid)).not.toThrow()
    const result = PoolHealthSchema.parse(valid)
    expect(result.models).toHaveLength(2)
    expect(result.models[0].name).toBe('qwen3:14b')
  })

  it('validates an empty models array', () => {
    const valid = { models: [] }
    expect(() => PoolHealthSchema.parse(valid)).not.toThrow()
    expect(PoolHealthSchema.parse(valid).models).toHaveLength(0)
  })

  it('rejects payload missing models field', () => {
    const invalid = {}
    expect(() => PoolHealthSchema.parse(invalid)).toThrow()
  })

  it('rejects payload where models is not an array', () => {
    const invalid = { models: 'all-good' }
    expect(() => PoolHealthSchema.parse(invalid)).toThrow()
  })

  it('rejects payload with invalid model inside models array', () => {
    const invalid = {
      models: [{ name: 'model', healthy: true, circuit_state: 'bad', last_latency_ms: 0, capabilities: [] }],
    }
    expect(() => PoolHealthSchema.parse(invalid)).toThrow()
  })
})
