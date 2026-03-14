import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LLMStatus } from './LLMStatus'
import type { LLMOverallStatus } from '../../hooks/useLLMHealth'
import type { PoolHealth } from '../../api/health'

// Mock the hook at module level
vi.mock('../../hooks/useLLMHealth', () => ({
  useLLMHealth: vi.fn(),
}))

import { useLLMHealth } from '../../hooks/useLLMHealth'

const mockUseLLMHealth = vi.mocked(useLLMHealth)

function makePoolData(models: Array<{ name: string; healthy: boolean; circuit_state: 'closed' | 'open' | 'half_open' }>): PoolHealth {
  return {
    models: models.map(m => ({
      ...m,
      last_latency_ms: 150,
      capabilities: ['tool_use'],
    })),
  }
}

function setup(overallStatus: LLMOverallStatus, data?: PoolHealth) {
  mockUseLLMHealth.mockReturnValue({
    data,
    overallStatus,
    isLoading: overallStatus === 'loading',
    isError: overallStatus === 'error',
  } as ReturnType<typeof useLLMHealth>)
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('LLMStatus', () => {
  it('renders a status dot element', () => {
    setup('healthy', makePoolData([{ name: 'qwen3:14b', healthy: true, circuit_state: 'closed' }]))
    render(<LLMStatus />)
    expect(screen.getByTestId('llm-status-dot')).toBeInTheDocument()
  })

  it('has data-status="healthy" when all models are healthy', () => {
    setup('healthy', makePoolData([{ name: 'qwen3:14b', healthy: true, circuit_state: 'closed' }]))
    render(<LLMStatus />)
    expect(screen.getByTestId('llm-status-dot')).toHaveAttribute('data-status', 'healthy')
  })

  it('has data-status="degraded" when status is degraded', () => {
    setup('degraded', makePoolData([
      { name: 'qwen3:14b', healthy: true, circuit_state: 'closed' },
      { name: 'qwen3:4b', healthy: false, circuit_state: 'open' },
    ]))
    render(<LLMStatus />)
    expect(screen.getByTestId('llm-status-dot')).toHaveAttribute('data-status', 'degraded')
  })

  it('has data-status="down" when status is down', () => {
    setup('down', makePoolData([{ name: 'qwen3:4b', healthy: false, circuit_state: 'open' }]))
    render(<LLMStatus />)
    expect(screen.getByTestId('llm-status-dot')).toHaveAttribute('data-status', 'down')
  })

  it('has data-status="loading" when status is loading', () => {
    setup('loading', undefined)
    render(<LLMStatus />)
    expect(screen.getByTestId('llm-status-dot')).toHaveAttribute('data-status', 'loading')
  })

  it('has data-status="error" when status is error', () => {
    setup('error', undefined)
    render(<LLMStatus />)
    expect(screen.getByTestId('llm-status-dot')).toHaveAttribute('data-status', 'error')
  })

  it('has aria-label containing the overall status', () => {
    setup('healthy', makePoolData([{ name: 'qwen3:14b', healthy: true, circuit_state: 'closed' }]))
    render(<LLMStatus />)
    const wrapper = screen.getByRole('status')
    expect(wrapper).toHaveAttribute('aria-label', 'LLM status: healthy')
  })

  it('has role="status"', () => {
    setup('healthy', makePoolData([{ name: 'qwen3:14b', healthy: true, circuit_state: 'closed' }]))
    render(<LLMStatus />)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('renders tooltip title with model names and statuses when data is available', () => {
    const data = makePoolData([
      { name: 'qwen3:14b', healthy: true, circuit_state: 'closed' },
      { name: 'qwen3:4b', healthy: false, circuit_state: 'open' },
    ])
    setup('degraded', data)
    render(<LLMStatus />)
    const wrapper = screen.getByRole('status')
    const title = wrapper.getAttribute('title') ?? ''
    expect(title).toContain('qwen3:14b')
    expect(title).toContain('qwen3:4b')
    expect(title).toContain('closed')
    expect(title).toContain('open')
  })

  it('renders loading tooltip when data is undefined', () => {
    setup('loading', undefined)
    render(<LLMStatus />)
    const wrapper = screen.getByRole('status')
    expect(wrapper.getAttribute('title')).toContain('loading')
  })
})
