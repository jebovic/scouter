import { useState, useEffect } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { Sidebar, OnboardingOverlay, NavRail, GlobalBottomNav } from '../components/scouter'
import { SidebarContext } from '../contexts/sidebar'
import { useOnboarding, useMissions } from '../hooks'
import { useSettings } from '../hooks/useSettings'
import { syncLanguageFromSettings } from '../i18n'
import styles from './Layout.module.css'

export function Layout() {
  const { show, step, totalSteps, nextStep, prevStep, dismiss } = useOnboarding()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const { missions } = useMissions()
  const { data: settings } = useSettings()
  const navigate = useNavigate()

  useEffect(() => {
    if (settings?.locale) {
      syncLanguageFromSettings(settings.locale)
    }
  }, [settings?.locale])

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
      <NavRail missions={missions} onNewMission={() => navigate('/')} />
      <GlobalBottomNav />
      <Sidebar
        open={sidebarOpen}
        missions={missions}
        onClose={() => setSidebarOpen(false)}
      />
      <div className={styles.content}><Outlet /></div>
    </SidebarContext.Provider>
  )
}
