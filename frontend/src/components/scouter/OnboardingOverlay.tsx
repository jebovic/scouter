import { useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import styles from './OnboardingOverlay.module.css'

interface OnboardingOverlayProps {
  show: boolean
  step: number
  totalSteps: number
  onNext: () => void
  onPrev: () => void
  onDismiss: () => void
}

export function OnboardingOverlay({ show, step, totalSteps, onNext, onPrev, onDismiss }: OnboardingOverlayProps) {
  const { t } = useTranslation()
  const dialogRef = useRef<HTMLDivElement>(null)

  // Focus trap: keep focus inside dialog when visible
  useEffect(() => {
    if (!show) return
    const dialog = dialogRef.current
    if (!dialog) return
    const focusable = dialog.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    first?.focus()
    function onKeyDown(e: KeyboardEvent) {
      if (e.key !== 'Tab') return
      if (e.shiftKey) {
        if (document.activeElement === first) { e.preventDefault(); last?.focus() }
      } else {
        if (document.activeElement === last) { e.preventDefault(); first?.focus() }
      }
    }
    dialog.addEventListener('keydown', onKeyDown)
    return () => dialog.removeEventListener('keydown', onKeyDown)
  }, [show, step])

  const STEPS = [
    {
      icon: '🎯',
      title: t('onboarding.step0Title'),
      description: t('onboarding.step0Desc'),
    },
    {
      icon: '🔍',
      title: t('onboarding.step1Title'),
      description: t('onboarding.step1Desc'),
    },
    {
      icon: '✅',
      title: t('onboarding.step2Title'),
      description: t('onboarding.step2Desc'),
    },
  ]

  if (!show) return null

  const current = STEPS[step]
  const isLast = step === totalSteps - 1

  return (
    <div className={styles.backdrop}>
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-label="Onboarding"
        className={styles.dialog}
      >
        <div className={styles.iconArea}>{current.icon}</div>
        <p className={styles.title}>{current.title}</p>
        <p className={styles.description}>{current.description}</p>

        <div className={styles.dots}>
          {Array.from({ length: totalSteps }).map((_, i) => (
            <button
              key={i}
              aria-label={`Step ${i + 1}`}
              className={[styles.dot, i === step ? styles.dotActive : ''].filter(Boolean).join(' ')}
              onClick={() => {/* dots are visual only */}}
            />
          ))}
        </div>

        <div className={styles.actions}>
          <button className={styles.btnSkip} onClick={onDismiss}>
            {t('common.skip')}
          </button>
          <div className={styles.btnNav}>
            {step > 0 && (
              <button className={styles.btnBack} onClick={onPrev}>
                {t('common.back')}
              </button>
            )}
            {isLast ? (
              <button className={styles.btnDone} onClick={onDismiss}>
                {t('common.done')}
              </button>
            ) : (
              <button className={styles.btnNext} onClick={onNext}>
                {t('common.next')}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
