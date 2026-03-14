import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import HQDashboard from './pages/HQDashboard'
import { ErrorBoundary, LoadingPulse } from './components/scouter'
import { Layout } from './layouts/Layout'
import { MissionLayout } from './layouts/MissionLayout'

const MissionOverview = lazy(() => import('./pages/MissionOverview'))
const OptionsExplorer = lazy(() => import('./pages/OptionsExplorer'))
const ShoppingTracker = lazy(() => import('./pages/ShoppingTracker'))

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<LoadingPulse label="Loading..." />}>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<HQDashboard />} />
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
