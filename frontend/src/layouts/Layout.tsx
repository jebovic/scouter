import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar, OnboardingOverlay } from '../components/scouter'
import { SidebarContext } from '../contexts/sidebar'
import { useOnboarding, useMissions } from '../hooks'

export function Layout() {
  const { show, step, totalSteps, nextStep, prevStep, dismiss } = useOnboarding()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const { missions } = useMissions()

  return (
    <SidebarContext.Provider value={{ openSidebar: () => setSidebarOpen(true) }}>
      <a href="#main-content" className="skip-to-content">Skip to main content</a>
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
      <Outlet />
    </SidebarContext.Provider>
  )
}
