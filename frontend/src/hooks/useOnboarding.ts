import { useState } from 'react'

const STORAGE_KEY = 'scouter_onboarding_dismissed'
const TOTAL_STEPS = 3

function isOnboardingDismissed(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

export interface OnboardingState {
  show: boolean
  step: number
  totalSteps: number
  nextStep: () => void
  prevStep: () => void
  dismiss: () => void
}

export function useOnboarding(): OnboardingState {
  const [show, setShow] = useState(() => !isOnboardingDismissed())
  const [step, setStep] = useState(0)

  function nextStep() {
    setStep(s => Math.min(s + 1, TOTAL_STEPS - 1))
  }

  function prevStep() {
    setStep(s => Math.max(s - 1, 0))
  }

  function dismiss() {
    try {
      localStorage.setItem(STORAGE_KEY, 'true')
    } catch {
      // ignore private browsing
    }
    setShow(false)
  }

  return { show, step, totalSteps: TOTAL_STEPS, nextStep, prevStep, dismiss }
}
