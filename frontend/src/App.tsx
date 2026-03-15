import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import HQDashboard from './pages/HQDashboard'
import { ErrorBoundary, LoadingPulse } from './components/scouter'
import { Layout } from './layouts/Layout'
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
const NotificationsPage = lazy(() => import('./pages/NotificationsPage'))
const EnvelopesPage = lazy(() => import('./pages/EnvelopesPage'))
const DealCalendarPage = lazy(() => import('./pages/DealCalendarPage'))
const PerformancePage = lazy(() => import('./pages/PerformancePage'))

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<LoadingPulse label="Loading..." />}>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<HQDashboard />} />
            <Route path="/search" element={<SearchPage />} />
            <Route path="/invites/:token" element={<JoinPage />} />
            <Route path="/history" element={<HistoryPage />} />
            <Route path="/stats" element={<StatsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/wishlist" element={<WishListPage />} />
            <Route path="/notifications" element={<NotificationsPage />} />
            <Route path="/envelopes" element={<EnvelopesPage />} />
            <Route path="/deal-calendar" element={<DealCalendarPage />} />
            <Route path="/performance" element={<PerformancePage />} />
            <Route path="/missions/:slug" element={<MissionLayout />}>
              <Route index element={<MissionOverview />} />
              <Route path="options" element={<OptionsExplorer />} />
              <Route path="shopping" element={<ShoppingTracker />} />
            </Route>
          </Route>
        </Routes>
      </Suspense>
    </ErrorBoundary>
  )
}

export default App
