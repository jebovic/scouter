import styles from './OnboardingOverlay.module.css'

interface Step {
  icon: string
  title: string
  description: string
}

const STEPS: Step[] = [
  {
    icon: '🎯',
    title: 'Welcome to SCOUTER',
    description: 'SCOUTER helps you research, compare, and make smarter purchase decisions. Start by creating a mission for anything you want to buy.',
  },
  {
    icon: '🔍',
    title: 'Research & Compare',
    description: 'Run the Research Agent to discover options, then use Price Intel to track costs across merchants. Compare everything in one place.',
  },
  {
    icon: '✅',
    title: 'Track & Decide',
    description: 'Score your options, set price alerts, and let the Decision Engine recommend the best choice based on your priorities.',
  },
]

interface OnboardingOverlayProps {
  show: boolean
  step: number
  totalSteps: number
  onNext: () => void
  onPrev: () => void
  onDismiss: () => void
}

export function OnboardingOverlay({ show, step, totalSteps, onNext, onPrev, onDismiss }: OnboardingOverlayProps) {
  if (!show) return null

  const current = STEPS[step]
  const isLast = step === totalSteps - 1

  return (
    <div className={styles.backdrop}>
      <div
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
            Skip
          </button>
          <div className={styles.btnNav}>
            {step > 0 && (
              <button className={styles.btnBack} onClick={onPrev}>
                Back
              </button>
            )}
            {isLast ? (
              <button className={styles.btnDone} onClick={onDismiss}>
                Done
              </button>
            ) : (
              <button className={styles.btnNext} onClick={onNext}>
                Next
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
