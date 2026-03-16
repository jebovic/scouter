import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { SummaryReport } from './SummaryReport'

// Mock both hooks so no real network calls are made.
vi.mock('../../hooks/useSummary', () => ({
  useSummary: vi.fn(),
  useGenerateSummary: vi.fn(),
}))

import { useSummary, useGenerateSummary } from '../../hooks/useSummary'
const mockUseSummary = vi.mocked(useSummary)
const mockUseGenerateSummary = vi.mocked(useGenerateSummary)

const MISSION_ID = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'

const MISSION_SLUG = 'test-mission'

const sampleSummary = {
  id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  missionId: MISSION_ID,
  content: '# Rapport\n\nContenu du rapport en markdown.',
  generatedAt: '2026-03-15T10:00:00Z',
}

function makeGenerateMock(overrides?: Partial<ReturnType<typeof useGenerateSummary>>) {
  return {
    generate: vi.fn(),
    isPending: false,
    error: null,
    ...overrides,
  }
}

function makeSummaryMock(overrides?: Partial<ReturnType<typeof useSummary>>) {
  return {
    summary: null,
    isLoading: false,
    ...overrides,
  }
}

beforeEach(() => {
  mockUseSummary.mockReturnValue(makeSummaryMock())
  mockUseGenerateSummary.mockReturnValue(makeGenerateMock())
})

describe('SummaryReport', () => {
  // ── Initial empty state ────────────────────────────────────────────────────

  it('shows a generate button when no summary exists', () => {
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByRole('button', { name: /generer/i })).toBeTruthy()
  })

  it('shows a prompt text when no summary exists', () => {
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByText(/generer.*rapport/i)).toBeTruthy()
  })

  it('calls generate when the generate button is clicked', () => {
    const generate = vi.fn()
    mockUseGenerateSummary.mockReturnValue(makeGenerateMock({ generate }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    fireEvent.click(screen.getByRole('button', { name: /generer/i }))
    expect(generate).toHaveBeenCalledOnce()
  })

  // ── Loading / generating state ─────────────────────────────────────────────

  it('shows loading message while generating', () => {
    mockUseGenerateSummary.mockReturnValue(makeGenerateMock({ isPending: true }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByText(/analyse votre mission/i)).toBeTruthy()
  })

  it('disables generate button while generating', () => {
    mockUseGenerateSummary.mockReturnValue(makeGenerateMock({ isPending: true }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    const btn = screen.getByRole('button', { name: /analyse/i })
    expect((btn as HTMLButtonElement).disabled).toBe(true)
  })

  // ── Content displayed when summary available ───────────────────────────────

  it('renders markdown content when summary is available', () => {
    mockUseSummary.mockReturnValue(makeSummaryMock({ summary: sampleSummary }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByText(/Contenu du rapport en markdown/)).toBeTruthy()
  })

  it('shows copy button when summary is available', () => {
    mockUseSummary.mockReturnValue(makeSummaryMock({ summary: sampleSummary }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByRole('button', { name: /copier/i })).toBeTruthy()
  })

  it('shows download button when summary is available', () => {
    mockUseSummary.mockReturnValue(makeSummaryMock({ summary: sampleSummary }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByRole('button', { name: /telecharger/i })).toBeTruthy()
  })

  it('shows regenerate button when summary is available', () => {
    mockUseSummary.mockReturnValue(makeSummaryMock({ summary: sampleSummary }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByRole('button', { name: /regenerer/i })).toBeTruthy()
  })

  it('does not show initial generate button when summary exists', () => {
    mockUseSummary.mockReturnValue(makeSummaryMock({ summary: sampleSummary }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    // The header generate button should not be present; only regenerate
    const buttons = screen.getAllByRole('button')
    const buttonLabels = buttons.map((b) => b.textContent ?? '')
    expect(buttonLabels.some((l) => /generer le rapport/i.test(l))).toBe(false)
  })

  it('calls generate when regenerate button is clicked', () => {
    const generate = vi.fn()
    mockUseSummary.mockReturnValue(makeSummaryMock({ summary: sampleSummary }))
    mockUseGenerateSummary.mockReturnValue(makeGenerateMock({ generate }))
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    fireEvent.click(screen.getByRole('button', { name: /regenerer/i }))
    expect(generate).toHaveBeenCalledOnce()
  })

  // ── Header title ───────────────────────────────────────────────────────────

  it('renders the "Rapport IA" title', () => {
    render(<SummaryReport missionSlug={MISSION_SLUG} />)
    expect(screen.getByText(/rapport ia/i)).toBeTruthy()
  })
})
