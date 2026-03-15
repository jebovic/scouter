import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { HealthScoreCard } from './HealthScoreCard'

vi.mock('../../hooks/useMissionHealth', () => ({
  useMissionHealth: vi.fn(),
}))

import { useMissionHealth } from '../../hooks/useMissionHealth'
const mockUseHealth = vi.mocked(useMissionHealth)

const mockReport = {
  missionId: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  overallScore: 72,
  grade: 'C' as const,
  signals: [
    { name: 'Research Depth', score: 60, weight: 0.25, description: 'Options researched', icon: '🔬' },
    { name: 'Option Coverage', score: 80, weight: 0.25, description: 'Options evaluated', icon: '📊' },
    { name: 'Budget Clarity', score: 100, weight: 0.20, description: 'Budget configured', icon: '💰' },
    { name: 'Decision Progress', score: 50, weight: 0.30, description: 'Comparing options', icon: '🎯' },
  ],
  summary: 'Mission en bonne progression.',
  advice: 'Continuez la recherche.',
  generatedAt: '2026-03-15T10:00:00Z',
}

function makeHealthMock(overrides?: Partial<ReturnType<typeof useMissionHealth>>) {
  return {
    data: undefined,
    isLoading: false,
    isError: false,
    isSuccess: false,
    error: null,
    ...overrides,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any
}

beforeEach(() => {
  mockUseHealth.mockReturnValue(makeHealthMock())
})

describe('HealthScoreCard', () => {
  it('renders the section title', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ data: mockReport, isSuccess: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText(/santé de la mission/i)).toBeTruthy()
  })

  it('shows the overall score when data is available', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ data: mockReport, isSuccess: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText('72')).toBeTruthy()
  })

  it('shows the grade when data is available', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ data: mockReport, isSuccess: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText('C')).toBeTruthy()
  })

  it('renders all four signal names', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ data: mockReport, isSuccess: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText('Research Depth')).toBeTruthy()
    expect(screen.getByText('Option Coverage')).toBeTruthy()
    expect(screen.getByText('Budget Clarity')).toBeTruthy()
    expect(screen.getByText('Decision Progress')).toBeTruthy()
  })

  it('shows the AI summary text', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ data: mockReport, isSuccess: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText('Mission en bonne progression.')).toBeTruthy()
  })

  it('shows the AI advice text', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ data: mockReport, isSuccess: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText('Continuez la recherche.')).toBeTruthy()
  })

  it('shows a loading skeleton when isLoading is true', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ isLoading: true }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    // Score should not appear; loading indicator should be present
    expect(screen.queryByText('72')).toBeNull()
    expect(document.querySelector('[data-testid="health-loading"]')).toBeTruthy()
  })

  it('shows an error message when fetch fails', () => {
    mockUseHealth.mockReturnValue(makeHealthMock({ isError: true, error: new Error('Network error') }))
    render(<HealthScoreCard missionSlug="mon-laptop" />)
    expect(screen.getByText(/impossible de charger/i)).toBeTruthy()
  })
})
