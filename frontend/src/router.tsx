import { lazy, Suspense } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import HQDashboard from './pages/HQDashboard'
import NotFoundPage from './pages/NotFoundPage'
import { ErrorBoundary, LoadingPulse } from './components/scouter'
import { Layout } from './layouts/Layout'
import { GlobalLayout } from './layouts/GlobalLayout'
import { MissionLayout } from './layouts/MissionLayout'

const MissionOverview = lazy(() => import('./pages/MissionOverview'))
const OptionsExplorer = lazy(() => import('./pages/OptionsExplorer'))
const ShoppingTracker = lazy(() => import('./pages/ShoppingTracker'))
const SearchPage = lazy(() => import('./pages/SearchPage').then((m) => ({ default: m.SearchPage })))
const HistoryPage = lazy(() => import('./pages/HistoryPage'))
const StatsPage = lazy(() => import('./pages/StatsPage'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))
const JoinPage = lazy(() => import('./pages/JoinPage'))
const WishListPage = lazy(() => import('./pages/WishListPage'))
const SharedWishlistPage = lazy(() => import('./pages/SharedWishlistPage'))
const NotificationsPage = lazy(() => import('./pages/NotificationsPage'))
const EnvelopesPage = lazy(() => import('./pages/EnvelopesPage'))
const DealCalendarPage = lazy(() => import('./pages/DealCalendarPage'))
const DigestPage = lazy(() => import('./pages/DigestPage'))
const LoyaltyPage = lazy(() => import('./pages/LoyaltyPage'))
const KanbanPage = lazy(() => import('./pages/KanbanPage'))
const CashbackPage = lazy(() => import('./pages/CashbackPage'))
const InsightsPage = lazy(() => import('./pages/InsightsPage'))

function SuspenseWrapper({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary>
      <Suspense fallback={<LoadingPulse label="Loading..." />}>{children}</Suspense>
    </ErrorBoundary>
  )
}

export const router = createBrowserRouter([
  {
    path: '/wishlist/shared',
    element: (
      <SuspenseWrapper>
        <SharedWishlistPage />
      </SuspenseWrapper>
    ),
  },
  {
    element: <Layout />,
    children: [
      // Global pages — wrapped with GlobalLayout (Topnav + page container)
      {
        element: <GlobalLayout />,
        children: [
          {
            path: '/',
            element: (
              <SuspenseWrapper>
                <HQDashboard />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/search',
            element: (
              <SuspenseWrapper>
                <SearchPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/invites/:token',
            element: (
              <SuspenseWrapper>
                <JoinPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/history',
            element: (
              <SuspenseWrapper>
                <HistoryPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/stats',
            element: (
              <SuspenseWrapper>
                <StatsPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/analytics',
            element: <Navigate to="/stats" replace />,
          },
          {
            path: '/settings',
            element: (
              <SuspenseWrapper>
                <SettingsPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/wishlist',
            element: (
              <SuspenseWrapper>
                <WishListPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/notifications',
            element: (
              <SuspenseWrapper>
                <NotificationsPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/envelopes',
            element: (
              <SuspenseWrapper>
                <EnvelopesPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/deal-calendar',
            element: (
              <SuspenseWrapper>
                <DealCalendarPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/performance',
            element: <Navigate to="/settings" replace />,
          },
          {
            path: '/digest',
            element: (
              <SuspenseWrapper>
                <DigestPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/loyalty',
            element: (
              <SuspenseWrapper>
                <LoyaltyPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/kanban',
            element: (
              <SuspenseWrapper>
                <KanbanPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/cashback',
            element: (
              <SuspenseWrapper>
                <CashbackPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/insights',
            element: (
              <SuspenseWrapper>
                <InsightsPage />
              </SuspenseWrapper>
            ),
          },
        ],
      },
      // Mission pages — MissionLayout has its own Topnav + BottomNav
      {
        path: '/missions/:slug',
        element: (
          <SuspenseWrapper>
            <MissionLayout />
          </SuspenseWrapper>
        ),
        children: [
          {
            index: true,
            element: (
              <SuspenseWrapper>
                <MissionOverview />
              </SuspenseWrapper>
            ),
          },
          {
            path: 'options',
            element: (
              <SuspenseWrapper>
                <OptionsExplorer />
              </SuspenseWrapper>
            ),
          },
          {
            path: 'shopping',
            element: (
              <SuspenseWrapper>
                <ShoppingTracker />
              </SuspenseWrapper>
            ),
          },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <NotFoundPage />,
  },
])
