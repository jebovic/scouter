import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import HQDashboard from './pages/HQDashboard'
import MissionOverview from './pages/MissionOverview'
import { ErrorBoundary, LoadingPulse } from './components/scouter'

const OptionsExplorer = lazy(() => import('./pages/OptionsExplorer'))
const ShoppingTracker = lazy(() => import('./pages/ShoppingTracker'))

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<LoadingPulse label="Loading..." />}>
        <Routes>
          <Route path="/" element={<HQDashboard />} />
          <Route path="/missions/:slug" element={<MissionOverview />} />
          <Route path="/missions/:slug/options" element={<OptionsExplorer />} />
          <Route path="/missions/:slug/shopping" element={<ShoppingTracker />} />
        </Routes>
      </Suspense>
    </ErrorBoundary>
  )
}

export default App
