import { useState } from 'react'
import { LOYALTY_PROGRAMS, getNextMilestone, getEquivalentValue } from '../../utils/loyaltyPrograms'
import { useLoyaltyPoints } from '../../hooks/useLoyaltyPoints'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './LoyaltyTracker.module.css'

export function LoyaltyTracker() {
  const { balances, setPoints, clearPoints } = useLoyaltyPoints()
  const { fmt, locale } = useFormatCurrency()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')

  function handleStartEdit(programId: string) {
    setEditingId(programId)
    setEditValue(String(balances[programId] ?? 0))
  }

  function handleSaveEdit(programId: string) {
    const newPoints = Math.max(0, parseInt(editValue, 10) || 0)
    setPoints(programId, newPoints)
    setEditingId(null)
  }

  function handleKeyDown(e: React.KeyboardEvent, programId: string) {
    if (e.key === 'Enter') {
      handleSaveEdit(programId)
    } else if (e.key === 'Escape') {
      setEditingId(null)
    }
  }

  const totalValue = Object.entries(balances).reduce((sum, [programId, points]) => {
    const program = LOYALTY_PROGRAMS.find((p) => p.id === programId)
    if (!program) return sum
    return sum + getEquivalentValue(program, points)
  }, 0)

  return (
    <div>
      {Object.keys(balances).length > 0 && (
        <div className={styles.summaryBar}>
          <strong>Valeur totale : </strong>
          <span className={styles.summaryValue}>
            {fmt(totalValue)}
          </span>
        </div>
      )}

      <div className={styles.grid}>
        {LOYALTY_PROGRAMS.map((program) => {
          const currentPoints = balances[program.id] ?? 0
          const nextMilestone = getNextMilestone(program, currentPoints)
          const equivalentValue = getEquivalentValue(program, currentPoints)
          const progressPercent = nextMilestone
            ? Math.min(100, (currentPoints / nextMilestone.points) * 100)
            : 100

          return (
            <div
              key={program.id}
              className={styles.card}
              style={{ '--program-color': program.color } as React.CSSProperties}
            >
              <div className={styles.header}>
                <span className={styles.icon}>{program.icon}</span>
                <h3 className={styles.name}>{program.name}</h3>
              </div>

              <div className={styles.points}>
                {editingId === program.id ? (
                  <input
                    autoFocus
                    type="number"
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    onBlur={() => handleSaveEdit(program.id)}
                    onKeyDown={(e) => handleKeyDown(e, program.id)}
                    className={styles.pointsInput}
                    min="0"
                  />
                ) : (
                  <span
                    className={styles.pointsValue}
                    onClick={() => handleStartEdit(program.id)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        handleStartEdit(program.id)
                      }
                    }}
                  >
                    {currentPoints.toLocaleString(locale)}
                  </span>
                )}
              </div>

              <div className={styles.value}>
                <span>Valeur :</span>
                <span className={styles.valueAmount}>
                  {fmt(equivalentValue)}
                </span>
              </div>

              <div className={styles.milestoneSection}>
                {nextMilestone ? (
                  <>
                    <div className={styles.milestoneLabel}>Prochain palier</div>
                    <div className={styles.progressContainer}>
                      <div className={styles.progressBar}>
                        <div
                          className={styles.progressFill}
                          style={{ width: `${progressPercent}%` }}
                        />
                      </div>
                      <span className={styles.progressText}>
                        {nextMilestone.points - currentPoints}
                      </span>
                    </div>
                    <div className={styles.milestonePill}>{nextMilestone.reward}</div>
                  </>
                ) : (
                  <div className={styles.maxedMessage}>
                    🎉 Niveau max atteint !
                  </div>
                )}
              </div>

              {currentPoints > 0 && (
                <button
                  className={styles.clearBtn}
                  onClick={() => clearPoints(program.id)}
                  aria-label={`Effacer les points de ${program.name}`}
                >
                  Effacer
                </button>
              )}
            </div>
          )
        })}
      </div>

      {Object.keys(balances).length === 0 && (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>💎</div>
          <p>Cliquez sur les valeurs pour ajouter vos points</p>
        </div>
      )}
    </div>
  )
}
