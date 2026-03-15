import { useState } from 'react'
import { useNegotiationScript } from '../../hooks/useNegotiationScript'
import type { ScriptStep } from '../../api/negotiationsim'
import styles from './NegotiationSimulator.module.css'

interface NegotiationSimulatorProps {
  missionId: string
  itemId: string
}

const PHASE_LABEL: Record<string, string> = {
  opening: 'Ouverture',
  negotiation: 'Négociation',
  closing: 'Conclusion',
}

function SavingsBadge({ savings }: { savings: number }) {
  if (savings <= 0) return null
  const formatted = savings.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })
  return (
    <span className={styles.savingsBadge} aria-label={`Économie potentielle : ${formatted}`}>
      -{formatted}
    </span>
  )
}

function StepBubble({ step }: { step: ScriptStep }) {
  const isYou = step.speaker === 'vous'
  return (
    <div className={`${styles.bubble} ${isYou ? styles.bubbleYou : styles.bubbleVendeur} ${styles[`phase_${step.phase}`]}`}>
      <span className={styles.bubbleSpeaker}>{isYou ? 'Vous' : 'Vendeur'}</span>
      <p className={styles.bubbleText}>&ldquo;{step.text}&rdquo;</p>
      {step.tip && (
        <p className={styles.bubbleTip}>
          <span aria-hidden="true">💡</span> {step.tip}
        </p>
      )}
    </div>
  )
}

export function NegotiationSimulator({ missionId, itemId }: NegotiationSimulatorProps) {
  const [enabled, setEnabled] = useState(false)
  const [open, setOpen] = useState(false)

  const { data, isFetching } = useNegotiationScript(missionId, itemId, enabled)

  function handleToggle() {
    setEnabled(true)
    setOpen((v) => !v)
  }

  // Group steps by phase for section headers.
  const phases: Array<{ phase: string; steps: ScriptStep[] }> = []
  if (data) {
    let currentPhase = ''
    for (const step of data.scripts) {
      if (step.phase !== currentPhase) {
        currentPhase = step.phase
        phases.push({ phase: step.phase, steps: [] })
      }
      phases[phases.length - 1].steps.push(step)
    }
  }

  return (
    <div className={styles.wrapper}>
      <button
        className={styles.triggerBtn}
        onClick={handleToggle}
        aria-expanded={open}
        aria-label="Simulateur de négociation"
        title="Simulateur de négociation"
      >
        {isFetching ? <span className={styles.spinner} aria-hidden="true" /> : '🎯 Négocier'}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Script de négociation">
          {isFetching && !data ? (
            <p className={styles.loading}>Génération du script...</p>
          ) : data ? (
            <>
              <div className={styles.header}>
                <div className={styles.headerInfo}>
                  <span className={styles.itemName}>{data.itemName}</span>
                  <span className={styles.priceTarget}>
                    Objectif&nbsp;:&nbsp;
                    {data.targetPrice.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })}
                  </span>
                </div>
                <SavingsBadge savings={data.potentialSavings} />
              </div>

              <div className={styles.conversation}>
                {phases.map(({ phase, steps }) => (
                  <section key={phase} className={styles.phaseSection}>
                    <h4 className={`${styles.phaseLabel} ${styles[`phaseLabel_${phase}`]}`}>
                      {PHASE_LABEL[phase] ?? phase}
                    </h4>
                    {steps.map((step, i) => (
                      <StepBubble key={i} step={step} />
                    ))}
                  </section>
                ))}
              </div>

              {data.tips.length > 0 && (
                <section className={styles.tipsSection} aria-label="Conseils de négociation">
                  <h4 className={styles.tipsTitle}>Conseils</h4>
                  <ul className={styles.tipsList}>
                    {data.tips.map((tip, i) => (
                      <li key={i} className={styles.tipItem}>{tip}</li>
                    ))}
                  </ul>
                </section>
              )}

              <button
                className={styles.closeBtn}
                onClick={() => setOpen(false)}
                aria-label="Fermer le simulateur"
              >
                Fermer
              </button>
            </>
          ) : null}
        </div>
      )}
    </div>
  )
}
