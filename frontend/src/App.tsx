import { lazy, Suspense, useState } from 'react'
import { Routes, Route } from 'react-router-dom'
import HQDashboard from './pages/HQDashboard'
import MissionOverview from './pages/MissionOverview'
import { ErrorBoundary, LoadingPulse, OnboardingOverlay, Sidebar } from './components/scouter'
import { useOnboarding, useMissions } from './hooks'
import { SidebarContext } from './contexts/sidebar'

const OptionsExplorer = lazy(() => import('./pages/OptionsExplorer'))
const ShoppingTracker = lazy(() => import('./pages/ShoppingTracker'))

function AppInner() {
  const { show, step, totalSteps, nextStep, prevStep, dismiss } = useOnboarding()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const { missions } = useMissions()

  return (
    <SidebarContext.Provider value={{ openSidebar: () => setSidebarOpen(true) }}>
      <OnboardingOverlay
        show={show}
        step={step}
        totalSteps={totalSteps}
        onNext={nextStep}
        onPrev={prevStep}
        onDismiss={dismiss}
      />
      <Sidebar
        open={sidebarOpen}
        missions={missions}
        onClose={() => setSidebarOpen(false)}
      />
      <Suspense fallback={<LoadingPulse label="Loading..." />}>
        <Routes>
          <Route path="/" element={<HQDashboard />} />
          <Route path="/missions/:slug" element={<MissionOverview />} />
          <Route path="/missions/:slug/options" element={<OptionsExplorer />} />
          <Route path="/missions/:slug/shopping" element={<ShoppingTracker />} />
        </Routes>
      </Suspense>
    </SidebarContext.Provider>
  )
}

function App() {
  return (
    <ErrorBoundary>
      <AppInner />
    </ErrorBoundary>
  )
}

export default App
