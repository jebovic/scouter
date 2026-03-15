import { useShoppingPersona } from '../../hooks/usePersona'
import styles from './ShoppingPersonaCard.module.css'

const SCORE_KEYS = ['frugalité', 'recherche', 'budget', 'qualité', 'impulsivité'] as const

export function ShoppingPersonaCard() {
  const { shoppingPersona, isLoading } = useShoppingPersona()

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.skeletonIcon} />
        <div className={styles.skeletonTitle} />
        <div className={styles.skeletonLine} />
        <div className={styles.skeletonLine} style={{ width: '75%' } as React.CSSProperties} />
      </div>
    )
  }

  if (!shoppingPersona) return null

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.icon}>{shoppingPersona.icon}</span>
        <div className={styles.titleBlock}>
          <h3 className={styles.archetype}>{shoppingPersona.archetype}</h3>
          <p className={styles.description}>{shoppingPersona.description}</p>
        </div>
      </div>

      {/* Score profile — vertical bars */}
      <div className={styles.scoreSection}>
        {SCORE_KEYS.map((key) => {
          const value = shoppingPersona.scoreProfile[key] ?? 50
          return (
            <div key={key} className={styles.scoreRow}>
              <span className={styles.scoreLabel}>{key}</span>
              <div className={styles.scoreBarWrap}>
                <div
                  className={styles.scoreBar}
                  style={{ '--score-pct': `${value}%` } as React.CSSProperties}
                />
              </div>
              <span className={styles.scoreValue}>{value}</span>
            </div>
          )
        })}
      </div>

      <div className={styles.insightGrid}>
        {/* Strengths */}
        {shoppingPersona.strengths.length > 0 && (
          <div className={styles.insightBlock}>
            <div className={styles.insightTitle}>Points forts</div>
            <ul className={styles.insightList}>
              {shoppingPersona.strengths.map((s, i) => (
                <li key={i} className={styles.strengthItem}>
                  <span className={styles.strengthBullet}>+</span>
                  {s}
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Blindspots */}
        {shoppingPersona.blindspots.length > 0 && (
          <div className={styles.insightBlock}>
            <div className={styles.insightTitle}>Points d'attention</div>
            <ul className={styles.insightList}>
              {shoppingPersona.blindspots.map((b, i) => (
                <li key={i} className={styles.blindspotItem}>
                  <span className={styles.blindspotBullet}>!</span>
                  {b}
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {/* Tips */}
      {shoppingPersona.tips.length > 0 && (
        <div className={styles.tipsSection}>
          <div className={styles.insightTitle}>Conseils personnalisés</div>
          <ol className={styles.tipsList}>
            {shoppingPersona.tips.map((tip, i) => (
              <li key={i} className={styles.tipItem}>
                <span className={styles.tipNumber}>{i + 1}</span>
                {tip}
              </li>
            ))}
          </ol>
        </div>
      )}
    </div>
  )
}
