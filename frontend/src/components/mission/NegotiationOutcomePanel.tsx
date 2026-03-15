import { useNegotiationOutcomes } from '../../hooks/useNegotiationOutcomes'
import type { NegotiationOutcome } from '../../api/negotiationoutcome'
import styles from './NegotiationOutcomePanel.module.css'

interface Props {
  missionId: string
}

function verdictClass(verdict: NegotiationOutcome['verdict']): string {
  switch (verdict) {
    case 'success': return styles.verdictSuccess
    case 'partial': return styles.verdictPartial
    case 'failed': return styles.verdictFailed
    default: return styles.verdictPending
  }
}

function verdictLabel(verdict: NegotiationOutcome['verdict']): string {
  switch (verdict) {
    case 'success': return '✅ Succès'
    case 'partial': return '⚡ Partiel'
    case 'failed': return '❌ Échoué'
    default: return '⏳ En cours'
  }
}

function fmt(value: number): string {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(value)
}

export function NegotiationOutcomePanel({ missionId }: Props) {
  const { summary, isLoading } = useNegotiationOutcomes(missionId)

  if (isLoading) return null

  const hasRealOutcomes = summary && summary.outcomes.some(o => o.verdict !== 'pending')

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <h3>🤝 Historique de Négociation</h3>
        {summary && summary.totalSaved > 0 && (
          <span className={styles.totalSaved}>Économisé {fmt(summary.totalSaved)}</span>
        )}
      </div>

      {!summary || !hasRealOutcomes ? (
        <p className={styles.empty}>Aucune négociation enregistrée pour cette mission</p>
      ) : (
        <>
          <div className={styles.summary}>
            <div className={styles.summaryItem}>
              <span className={styles.summaryLabel}>Total économisé</span>
              <span className={styles.summaryValue}>{fmt(summary.totalSaved)}</span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.summaryLabel}>Taux de succès moy.</span>
              <span className={styles.summaryValue}>{Math.round(summary.avgSuccessRate * 100)}%</span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.summaryLabel}>Meilleur marchand</span>
              <span className={styles.summaryValue}>{summary.bestMerchant || '—'}</span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.summaryLabel}>Tentatives</span>
              <span className={styles.summaryValue}>{summary.totalAttempts}</span>
            </div>
          </div>

          <div className={styles.tableWrapper}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Article</th>
                  <th>Marchand</th>
                  <th>Prix demandé</th>
                  <th>Prix négocié</th>
                  <th>Économie</th>
                  <th>Statut</th>
                  <th>Conseil</th>
                </tr>
              </thead>
              <tbody>
                {summary.outcomes.map((o) => (
                  <tr key={o.itemId}>
                    <td>{o.itemName}</td>
                    <td>{o.merchant}</td>
                    <td>{fmt(o.askingPrice)}</td>
                    <td>{o.negotiatedPrice != null ? fmt(o.negotiatedPrice) : '—'}</td>
                    <td>{o.saved > 0 ? fmt(o.saved) : '—'}</td>
                    <td>
                      <span className={verdictClass(o.verdict)}>{verdictLabel(o.verdict)}</span>
                    </td>
                    <td className={styles.tip}>{o.tipForMerchant}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
