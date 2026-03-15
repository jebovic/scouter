import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CoachPanel } from './CoachPanel'
import * as coachHooks from '../../hooks/useCoach'

vi.mock('../../hooks/useCoach')
vi.mock('../scouter', () => ({
  Skeleton: ({ variant }: { variant: string }) => (
    <div data-testid="skeleton-{variant}" />
  ),
  SkeletonGrid: ({ count }: { count: number }) => (
    <div data-testid="skeleton-grid">{count} skeletons</div>
  ),
}))

describe('CoachPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockCoachTips = vi.mocked(coachHooks.useCoachTips)
  const mockRefreshCoach = vi.mocked(coachHooks.useRefreshCoach)

  it('renders nothing when missionSlug is undefined', () => {
    mockCoachTips.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      isSuccess: false,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    const { container } = render(<CoachPanel missionSlug={undefined} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders loading state with skeletons', () => {
    mockCoachTips.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
      isSuccess: false,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByTestId('skeleton-grid')).toBeInTheDocument()
    expect(screen.getByText('2 skeletons')).toBeInTheDocument()
  })

  it('renders error state', () => {
    mockCoachTips.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
      isSuccess: false,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByText(/Failed to load coaching tips/)).toBeInTheDocument()
  })

  it('renders empty state when tips array is empty', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByText(/No tips available at the moment/)).toBeInTheDocument()
  })

  it('renders tips with all fields', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Expand your options',
            body: 'Consider more brands for comparison',
            action: 'Launch research',
          },
          {
            id: 'tip2',
            category: 'budget',
            priority: 'medium',
            title: 'Set budget alerts',
            body: 'Monitor spending to stay on track',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByText('Expand your options')).toBeInTheDocument()
    expect(screen.getByText('Consider more brands for comparison')).toBeInTheDocument()
    expect(screen.getByText('Set budget alerts')).toBeInTheDocument()
    expect(screen.getByText('Monitor spending to stay on track')).toBeInTheDocument()
  })

  it('renders action button when tip has action field', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Expand your options',
            body: 'Consider more brands',
            action: 'Launch research',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByText('Launch research')).toBeInTheDocument()
  })

  it('does not render action button when action field is missing', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'budget',
            priority: 'low',
            title: 'Monitor spending',
            body: 'Keep track of expenses',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.queryByRole('button', { name: /Monitor/ })).toBeNull()
  })

  it('expands and collapses panel on header click', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Test tip',
            body: 'Test body',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    // Initially expanded, so tip should be visible
    expect(screen.getByText('Test tip')).toBeInTheDocument()

    // Click expand button to collapse
    const expandBtn = screen.getByRole('button', { name: /Collapse/ })
    fireEvent.click(expandBtn)

    // After collapse, tip should not be visible
    expect(screen.queryByText('Test tip')).not.toBeInTheDocument()

    // Click expand button again to expand
    const expandBtnAgain = screen.getByRole('button', { name: /Expand/ })
    fireEvent.click(expandBtnAgain)

    // Tip should be visible again
    expect(screen.getByText('Test tip')).toBeInTheDocument()
  })

  it('calls refresh function when refresh button is clicked', () => {
    const mockRefresh = vi.fn()
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Test tip',
            body: 'Test body',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: mockRefresh,
      isPending: false,
    } as any)

    const { container } = render(<CoachPanel missionSlug="test-mission" />)

    const refreshBtn = container.querySelector('button[title="Refresh tips"]') as HTMLButtonElement
    fireEvent.click(refreshBtn)

    expect(mockRefresh).toHaveBeenCalledTimes(1)
  })

  it('disables refresh button while refreshing', () => {
    const mockRefresh = vi.fn()
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Test tip',
            body: 'Test body',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: mockRefresh,
      isPending: true,
    } as any)

    const { container } = render(<CoachPanel missionSlug="test-mission" />)

    const refreshBtn = container.querySelector('button[title="Refreshing..."]') as HTMLButtonElement
    expect(refreshBtn).toBeDisabled()
  })

  it('does not show refresh button during loading', () => {
    mockCoachTips.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
      isSuccess: false,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.queryByRole('button', { name: /Refresh/ })).not.toBeInTheDocument()
  })

  it('renders generated time in footer', () => {
    const generatedAt = new Date('2026-03-15T10:30:00Z').toISOString()
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Test tip',
            body: 'Test body',
          },
        ],
        generatedAt,
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByText(/Generated/)).toBeInTheDocument()
  })

  it('stops propagation when refresh button is clicked', () => {
    const mockRefresh = vi.fn()
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Test tip',
            body: 'Test body',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: mockRefresh,
      isPending: false,
    } as any)

    const { container } = render(<CoachPanel missionSlug="test-mission" />)

    const refreshBtn = container.querySelector('button[title="Refresh tips"]') as HTMLButtonElement
    fireEvent.click(refreshBtn)

    // Verify refresh was called
    expect(mockRefresh).toHaveBeenCalledTimes(1)
  })

  it('renders panel title with icon', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    render(<CoachPanel missionSlug="test-mission" />)

    expect(screen.getByText(/AI Coach/)).toBeInTheDocument()
  })

  it('renders category icons for each tip', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'Research tip',
            body: 'Test body',
          },
          {
            id: 'tip2',
            category: 'budget',
            priority: 'medium',
            title: 'Budget tip',
            body: 'Test body',
          },
          {
            id: 'tip3',
            category: 'timing',
            priority: 'low',
            title: 'Timing tip',
            body: 'Test body',
          },
          {
            id: 'tip4',
            category: 'decision',
            priority: 'high',
            title: 'Decision tip',
            body: 'Test body',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    const { container } = render(<CoachPanel missionSlug="test-mission" />)

    const categoryIcons = container.querySelectorAll('[class*="categoryIcon"]')
    expect(categoryIcons.length).toBe(4)
    expect(container.innerHTML).toContain('🔬') // research
    expect(container.innerHTML).toContain('💰') // budget
    expect(container.innerHTML).toContain('⏰') // timing
    expect(container.innerHTML).toContain('⚖️') // decision
  })

  it('renders priority badges for each tip', () => {
    mockCoachTips.mockReturnValue({
      data: {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research',
            priority: 'high',
            title: 'High priority',
            body: 'Test body',
          },
          {
            id: 'tip2',
            category: 'budget',
            priority: 'medium',
            title: 'Medium priority',
            body: 'Test body',
          },
          {
            id: 'tip3',
            category: 'timing',
            priority: 'low',
            title: 'Low priority',
            body: 'Test body',
          },
        ],
        generatedAt: new Date().toISOString(),
      },
      isLoading: false,
      error: null,
      isSuccess: true,
    } as any)
    mockRefreshCoach.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)

    const { container } = render(<CoachPanel missionSlug="test-mission" />)

    expect(container.innerHTML).toContain('🔴') // high
    expect(container.innerHTML).toContain('🟡') // medium
    expect(container.innerHTML).toContain('🟢') // low
  })
})
