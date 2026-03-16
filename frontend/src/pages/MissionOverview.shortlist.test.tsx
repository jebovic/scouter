import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'

const mockUsePurchaseRecord = vi.fn(() => ({ data: null, isLoading: false }))

vi.mock('../hooks', () => ({
  useMission: () => ({
    mission: {
      id: 'mission-1',
      slug: 'test',
      name: 'Test',
      phase: 'buying',
      budget: 2500,
      currency: 'EUR',
      constraints: [],
      category: 'computing',
    },
    isLoading: false,
    error: null,
  }),
  useUpdateMission: () => ({ updateMission: vi.fn(), isPending: false }),
  useDeleteMission: () => ({ deleteMission: vi.fn(), isPending: false }),
  useArchiveMission: () => ({ archiveMission: vi.fn(), isPending: false }),
  useUnarchiveMission: () => ({ unarchiveMission: vi.fn(), isPending: false }),
  useOptions: () => ({
    options: [{ id: 'opt-1', name: 'MacBook Pro', badge: 'recommended', pinned: true, priceRange: { min: 2000, max: 2500, best: 2200 }, attributes: [], warnings: [], rejected: false }],
    isLoading: false,
    error: null,
  }),
  useShopping: () => ({ shoppingItems: [], isLoading: false, items: [] }),
  usePurchaseRecord: (...args: unknown[]) => mockUsePurchaseRecord(...args),
  useResearch: () => ({ researchJob: null, isLoading: false }),
  usePriceIntel: () => ({ priceIntel: null, triggerPricing: vi.fn(), isPending: false }),
  useKeyboardShortcuts: () => undefined,
  useSuggestCategory: () => ({ suggestedCategory: null, mutate: vi.fn(), isPending: false, data: null }),
  useUnpinAllOptions: () => ({ unpinAllOptions: vi.fn(), isPending: false }),
}))

vi.mock('../hooks/useBudgetAlerts', () => ({
  useBudgetAlerts: () => ({ alerts: [] }),
}))

vi.mock('../hooks/useScorecard', () => ({
  useScorecard: () => ({ data: null, isLoading: false }),
}))

vi.mock('../hooks/useResearch', () => ({
  useTriggerResearch: () => ({ mutateAsync: vi.fn(), isPending: false }),
}))

vi.mock('../components/mission', () => ({
  CategoryTemplate: () => null,
  DecisionPanel: () => null,
  MissionTimeline: () => null,
  PurchaseForm: () => <div data-testid="purchase-form" />,
  LessonsField: () => null,
  CollaboratorsPanel: () => null,
  TravelSearchWidget: () => null,
  TimingAdvisorCard: () => null,
  ExportPanel: () => null,
  ReceiptScanner: () => null,
  SummaryReport: () => null,
  CoachPanel: () => null,
  HealthScoreCard: () => null,
  MissionSummaryCard: () => null,
  CommentThread: () => null,
  CategoryBadge: () => null,
  MissionGoalTracker: () => null,
  BudgetRecommendations: () => null,
  SalesCalendar: () => null,
  EcoScorePanel: () => null,
  MissionProgressWidget: () => null,
  GiftFinderWidget: () => null,
  LoyaltySummaryPanel: () => null,
  MissionROICard: () => null,
  InflationTrackerPanel: () => null,
  DecisionMatrixTable: () => null,
  SmartAlertsPanel: () => null,
  VoteSummaryPanel: () => null,
  MissionReportButton: () => null,
  ReorderSuggestionsPanel: () => null,
  NegotiationOutcomePanel: () => null,
  BundleDealsPanel: () => null,
  BurnRateCard: () => null,
  RegretAnalyzerCard: () => null,
  ListOptimizerPanel: () => null,
  CashbackSummaryPanel: () => null,
  PriceDropWatchlist: () => null,
  SeasonalCalendarPanel: () => null,
  BudgetAdvisorPanel: () => null,
  ComparisonScorePanel: () => null,
  PriceAlertDigestPanel: () => null,
  SpendingVelocityCard: () => null,
  ExpenseCategoryPanel: () => null,
  MissionActionBar: () => null,
  MissionEditModal: () => null,
  ShortlistPanel: ({ options, onSelect }: { options: { id: string; name: string; pinned: boolean }[]; onSelect: (o: unknown) => void }) => (
    <div>
      <span>Your Shortlist</span>
      {options.filter(o => o.pinned).map(o => (
        <div key={o.id}>
          <span>{o.name}</span>
          <button onClick={() => onSelect(o)}>Select</button>
        </div>
      ))}
    </div>
  ),
}))

vi.mock('../components/forecast', () => ({
  ForecastPanel: () => null,
}))

vi.mock('../components/scouter', () => ({
  LoadingPulse: () => null,
  BudgetBar: () => null,
  StatusBadge: () => null,
  ToastContainer: () => null,
  useToasts: () => ({ toasts: [], toast: vi.fn(), addToast: vi.fn(), removeToast: vi.fn() }),
  NextActionNudge: () => null,
  useToast: () => ({ toast: vi.fn() }),
}))

vi.mock('../components/scouter/LastVisitCard', () => ({
  setLastVisitedMission: vi.fn(),
}))

vi.mock('./mission-overview/PhaseSection', () => ({
  PhaseSection: () => null,
}))

vi.mock('./mission-overview/QuickNavSection', () => ({
  QuickNavSection: () => null,
}))

vi.mock('../components/mission/MissionActionBar', () => ({
  MissionActionBar: () => null,
}))

vi.mock('../components/mission/MissionEditModal', () => ({
  MissionEditModal: () => null,
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { language: 'en', changeLanguage: vi.fn() },
  }),
  Trans: ({ children }: { children: React.ReactNode }) => children,
  initReactI18next: { type: '3rdParty', init: vi.fn() },
}))

import MissionOverview from './MissionOverview'

function renderOverview() {
  const queryClient = new QueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/missions/test']}>
        <Routes>
          <Route path="/missions/:slug" element={<MissionOverview />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('MissionOverview shortlist (buying phase)', () => {
  beforeEach(() => {
    mockUsePurchaseRecord.mockReturnValue({ data: null, isLoading: false })
  })

  it('renders ShortlistPanel when phase is buying', () => {
    renderOverview()
    expect(screen.getByText(/your shortlist/i)).toBeInTheDocument()
  })

  it('shows pinned option in shortlist', () => {
    renderOverview()
    expect(screen.getByText('MacBook Pro')).toBeInTheDocument()
  })

  it('shows replace-confirm dialog when a purchase record exists and Select is clicked', async () => {
    mockUsePurchaseRecord.mockReturnValue({ data: { id: 'p1' }, isLoading: false })
    const user = userEvent.setup()
    renderOverview()

    const selectBtn = screen.getByRole('button', { name: /select/i })
    await user.click(selectBtn)

    expect(screen.getByText('shortlist.replaceConfirm')).toBeInTheDocument()
  })
})
