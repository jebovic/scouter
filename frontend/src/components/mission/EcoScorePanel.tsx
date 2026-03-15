import { useEcoScore } from '../../hooks/useEcoScore'
import type { ItemEcoScore } from '../../api/ecoscore'
import styles from './EcoScorePanel.module.css'

interface Props {
  missionId: string
}

function ItemRow({ item }: { item: ItemEcoScore }) {
  return (
    <div className={styles.itemRow}>
      <span className={styles.itemName} title={item.itemName}>{item.itemName}</span>
      <div className={styles.itemRight}>
        <span className={styles.itemCO2}>{item.co2kg.toFixed(1)} kg CO₂</span>
        <span className={styles.itemGrade} data-grade={item.grade}>{item.grade}</span>
      </div>
    </div>
  )
}

export function EcoScorePanel({ missionId }: Props) {
  const { data, isLoading, isError } = useEcoScore(missionId)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Score écologique">
        <div className={styles.header}>
          <span className={styles.title}>Score écologique</span>
        </div>
        <div className={styles.gradeRow}>
          <div className={`${styles.skeleton} ${styles.skeletonCircle}`} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '50%' }} />
            <div className={`${styles.skeleton} ${styles.skeletonLine}`} style={{ width: '35%' }} />
          </div>
        </div>
        {[1, 2, 3].map((i) => (
          <div key={i} className={`${styles.skeleton} ${styles.skeletonLine}`} />
        ))}
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Score écologique">
        <div className={styles.header}>
          <span className={styles.title}>Score écologique</span>
        </div>
        <p className={styles.errorState}>Impossible de calculer le score écologique.</p>
      </section>
    )
  }

  if (data.itemScores.length === 0) {
    return (
      <section className={styles.card} aria-label="Score écologique">
        <div className={styles.header}>
          <span className={styles.title}>Score écologique</span>
        </div>
        <p className={styles.emptyState}>Ajoutez des articles pour voir votre empreinte carbone.</p>
      </section>
    )
  }

  // Top 3 highest-CO2 items
  const topItems = [...data.itemScores]
    .sort((a, b) => b.co2kg - a.co2kg)
    .slice(0, 3)

  const tipsToShow = data.ecoTips.slice(0, 2)

  return (
    <section className={styles.card} aria-label="Score écologique">
      <div className={styles.header}>
        <span className={styles.title}>Score écologique</span>
      </div>

      {/* Grade badge + CO2 total */}
      <div className={styles.gradeRow}>
        <div className={styles.gradeBadge} data-grade={data.ecoGrade} aria-label={`Note ${data.ecoGrade}`}>
          {data.ecoGrade}
        </div>
        <div className={styles.gradeMeta}>
          <span className={styles.co2Value}>{data.totalCO2kg.toFixed(1)} kg</span>
          <span className={styles.co2Label}>CO₂ estimé</span>
          <div className={styles.treesRow}>
            <span className={styles.treesCount}>🌳 {data.compensationTreesNeeded} arbre{data.compensationTreesNeeded !== 1 ? 's' : ''}</span>
            {' '}pour compenser
          </div>
        </div>
      </div>

      {/* Top 3 CO2 items */}
      {topItems.length > 0 && (
        <div className={styles.itemList}>
          {topItems.map((item) => (
            <ItemRow key={item.itemId} item={item} />
          ))}
        </div>
      )}

      {/* Eco tips */}
      {tipsToShow.length > 0 && (
        <div className={styles.tipsSection}>
          <span className={styles.tipsTitle}>Conseils éco</span>
          {tipsToShow.map((tip, i) => (
            <p key={i} className={styles.tipItem}>{tip}</p>
          ))}
        </div>
      )}
    </section>
  )
}
