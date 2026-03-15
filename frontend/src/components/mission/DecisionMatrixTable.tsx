import type { RankedItem, ItemScores } from '../../api/decisionmatrix'
import { useDecisionMatrix } from '../../hooks/useDecisionMatrix'
import styles from './DecisionMatrixTable.module.css'

// ── Score colour helpers ────────────────────────────────────────────────────

/** Returns a CSS background and text colour based on a 0–20 criterion score. */
function scoreBg(score: number): React.CSSProperties {
  // Intensity: 0 = dim red, 10 = neutral, 20 = vivid green
  const ratio = score / 20 // 0..1
  if (ratio >= 0.75) {
    return {
      '--score-bg': 'color-mix(in srgb, var(--green) 20%, transparent)',
      '--score-color': 'var(--green)',
    } as React.CSSProperties
  }
  if (ratio >= 0.5) {
    return {
      '--score-bg': 'color-mix(in srgb, var(--cyan) 15%, transparent)',
      '--score-color': 'var(--cyan)',
    } as React.CSSProperties
  }
  if (ratio >= 0.25) {
    return {
      '--score-bg': 'color-mix(in srgb, var(--gold) 15%, transparent)',
      '--score-color': 'var(--gold)',
    } as React.CSSProperties
  }
  return {
    '--score-bg': 'color-mix(in srgb, var(--coral) 15%, transparent)',
    '--score-color': 'var(--coral)',
  } as React.CSSProperties
}

/** Returns badge CSS variables based on the recommendation text. */
function badgeStyle(recommendation: string): React.CSSProperties {
  if (recommendation === 'Achetez maintenant') {
    return {
      '--badge-bg': 'color-mix(in srgb, var(--green) 20%, transparent)',
      '--badge-color': 'var(--green)',
    } as React.CSSProperties
  }
  if (recommendation === 'Très bon moment pour acheter') {
    return {
      '--badge-bg': 'color-mix(in srgb, var(--cyan) 20%, transparent)',
      '--badge-color': 'var(--cyan)',
    } as React.CSSProperties
  }
  if (recommendation === 'À surveiller de près') {
    return {
      '--badge-bg': 'color-mix(in srgb, var(--gold) 20%, transparent)',
      '--badge-color': 'var(--gold)',
    } as React.CSSProperties
  }
  if (recommendation === 'Attendez') {
    return {
      '--badge-bg': 'color-mix(in srgb, var(--orange) 20%, transparent)',
      '--badge-color': 'var(--orange)',
    } as React.CSSProperties
  }
  // Différez l'achat
  return {
    '--badge-bg': 'color-mix(in srgb, var(--coral) 20%, transparent)',
    '--badge-color': 'var(--coral)',
  } as React.CSSProperties
}

/** Returns a CSS colour for the total score column. */
function totalColor(score: number): React.CSSProperties {
  if (score >= 80) return { '--total-color': 'var(--green)' } as React.CSSProperties
  if (score >= 60) return { '--total-color': 'var(--cyan)' } as React.CSSProperties
  if (score >= 40) return { '--total-color': 'var(--gold)' } as React.CSSProperties
  if (score >= 20) return { '--total-color': 'var(--orange)' } as React.CSSProperties
  return { '--total-color': 'var(--coral)' } as React.CSSProperties
}

// ── Skeleton ────────────────────────────────────────────────────────────────

function SkeletonRows() {
  return (
    <>
      {[1, 2, 3].map((i) => (
        <div key={i} className={styles.skeletonRow}>
          {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((j) => (
            <div key={j} className={`${styles.skeleton} ${styles.skeletonCell}`} />
          ))}
        </div>
      ))}
    </>
  )
}

// ── Score cell ───────────────────────────────────────────────────────────────

function ScoreCell({ value }: { value: number }) {
  return (
    <td className={styles.scoreCell} style={scoreBg(value)}>
      {value}
    </td>
  )
}

// ── Row ──────────────────────────────────────────────────────────────────────

function MatrixRow({ item, isTop }: { item: RankedItem; isTop: boolean }) {
  const scores: ItemScores = item.scores
  return (
    <tr className={isTop ? styles.topRow : undefined}>
      <td>
        <span className={`${styles.rank} ${isTop ? styles.rankTop : ''}`}>
          {isTop ? '★' : item.rank}
        </span>
      </td>
      <td>
        <span className={styles.itemName} title={item.name}>
          {item.name}
        </span>
      </td>
      <ScoreCell value={scores.prix} />
      <ScoreCell value={scores.urgence} />
      <ScoreCell value={scores.valeur} />
      <ScoreCell value={scores.disponibilite} />
      <ScoreCell value={scores.confiance} />
      <td className={styles.totalCell} style={totalColor(item.totalScore)}>
        {item.totalScore}
      </td>
      <td>
        <span className={styles.badge} style={badgeStyle(item.recommendation)}>
          {item.recommendation}
        </span>
      </td>
    </tr>
  )
}

// ── Main component ───────────────────────────────────────────────────────────

interface Props {
  missionId: string | undefined
}

export function DecisionMatrixTable({ missionId }: Props) {
  const { data, isLoading, isError } = useDecisionMatrix(missionId)

  return (
    <section className={styles.card} aria-label="Matrice de décision d'achat">
      <div className={styles.header}>
        <span className={styles.title}>Matrice de décision</span>
      </div>

      {isLoading && (
        <div aria-busy="true" aria-label="Chargement de la matrice">
          <SkeletonRows />
        </div>
      )}

      {isError && (
        <p className={styles.errorState}>Impossible de charger la matrice de décision.</p>
      )}

      {!isLoading && !isError && data && data.items.length === 0 && (
        <p className={styles.empty}>Aucun article dans cette mission.</p>
      )}

      {!isLoading && !isError && data && data.items.length > 0 && (
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>#</th>
                <th>Article</th>
                <th title="Score prix vs objectif">Prix</th>
                <th title="Score urgence selon statut">Urgence</th>
                <th title="Score valeur relative">Valeur</th>
                <th title="Score disponibilité">Dispo</th>
                <th title="Score confiance données">Confiance</th>
                <th>Total</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <MatrixRow
                  key={item.itemId}
                  item={item}
                  isTop={data.topRecommendation?.itemId === item.itemId}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
