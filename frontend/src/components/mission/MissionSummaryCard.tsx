import { useState } from 'react'
import { useMissionSummary } from '../../hooks/useSummary'
import type { MissionSummaryDTO } from '../../api/summary'
import styles from './MissionSummaryCard.module.css'

interface Props {
  missionSlug: string
}

// ── Verdict helpers ───────────────────────────────────────────────────────────

type Verdict = MissionSummaryDTO['verdict']

function verdictLabel(v: Verdict): string {
  switch (v) {
    case 'go':           return '🟢 ACHETEZ'
    case 'wait':         return '🟡 ATTENDEZ'
    case 'research_more': return '🔵 À EXPLORER'
  }
}

function verdictMod(v: Verdict): string {
  switch (v) {
    case 'go':           return styles.verdictGo
    case 'wait':         return styles.verdictWait
    case 'research_more': return styles.verdictResearch
  }
}

function relativeTime(unixSec: number): string {
  const diffSec = Math.max(0, Math.floor(Date.now() / 1000) - unixSec)
  if (diffSec < 60) return 'il y a quelques secondes'
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `il y a ${diffMin} minute${diffMin > 1 ? 's' : ''}`
  const diffH = Math.floor(diffMin / 60)
  return `il y a ${diffH} heure${diffH > 1 ? 's' : ''}`
}

// ── Component ─────────────────────────────────────────────────────────────────

export function MissionSummaryCard({ missionSlug }: Props) {
  const [enabled, setEnabled] = useState(false)
  const { brief, isLoading, isError, refetch, invalidate } = useMissionSummary(missionSlug, enabled)

  function handleGenerate() {
    if (!enabled) {
      setEnabled(true)
    } else {
      // Force re-fetch by invalidating cache then re-fetching
      invalidate()
      // Small timeout to let invalidate settle before refetch
      setTimeout(() => refetch(), 50)
    }
  }

  // ── Idle state (button not yet clicked) ──────────────────────────────────
  if (!enabled) {
    return (
      <section className={styles.card} aria-label="Brief IA de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Brief exécutif IA</span>
          <button
            className={styles.briefBtn}
            onClick={handleGenerate}
            aria-label="Générer le brief IA"
          >
            ⚡ Brief IA
          </button>
        </div>
        <p className={styles.hint}>
          Obtenez un résumé IA en un clic : meilleure option, budget, risques et prochaine action.
        </p>
      </section>
    )
  }

  // ── Loading state ─────────────────────────────────────────────────────────
  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Brief IA de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Brief exécutif IA</span>
        </div>
        <div className={styles.loadingState} data-testid="summary-loading">
          <span className={styles.spinner} aria-hidden="true" />
          <span className={styles.loadingText}>Génération du brief...</span>
        </div>
      </section>
    )
  }

  // ── Error state ───────────────────────────────────────────────────────────
  if (isError || !brief) {
    return (
      <section className={styles.card} aria-label="Brief IA de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Brief exécutif IA</span>
          <button
            className={styles.briefBtn}
            onClick={handleGenerate}
            aria-label="Réessayer"
          >
            ⚡ Réessayer
          </button>
        </div>
        <p className={styles.errorState}>Impossible de générer le brief. Réessayez dans un moment.</p>
      </section>
    )
  }

  // ── Result state ──────────────────────────────────────────────────────────
  return (
    <section className={styles.card} aria-label="Brief IA de la mission">
      <div className={styles.header}>
        <span className={styles.title}>Brief exécutif IA</span>
        <div className={styles.headerActions}>
          <span className={`${styles.verdictBadge} ${verdictMod(brief.verdict)}`}>
            {verdictLabel(brief.verdict)}
          </span>
          <button
            className={styles.refreshBtn}
            onClick={handleGenerate}
            aria-label="Rafraîchir le brief IA"
            title="Rafraîchir le brief"
          >
            ↺
          </button>
        </div>
      </div>

      <ul className={styles.bulletList} aria-label="Points clés">
        {brief.bullets.map((bullet, i) => (
          <li key={i} className={styles.bullet}>
            {bullet}
          </li>
        ))}
      </ul>

      <p className={styles.timestamp} aria-label="Heure de génération">
        {relativeTime(brief.cachedAt)}
      </p>
    </section>
  )
}
