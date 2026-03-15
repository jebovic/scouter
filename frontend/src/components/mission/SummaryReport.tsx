import { useSummary, useGenerateSummary } from '../../hooks/useSummary'
import styles from './SummaryReport.module.css'

interface Props {
  missionId: string
}

export function SummaryReport({ missionId }: Props) {
  const { summary, isLoading: loadingCached } = useSummary(missionId)
  const { generate, isPending: generating } = useGenerateSummary(missionId)

  const isPending = loadingCached || generating

  function handleCopy() {
    if (!summary) return
    navigator.clipboard.writeText(summary.content).catch(() => {
      // clipboard API may be unavailable in non-secure contexts — silently ignore
    })
  }

  function handleDownload() {
    if (!summary) return
    const blob = new Blob([summary.content], { type: 'text/markdown;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `rapport-achat-${summary.missionId}.md`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>Rapport IA</span>
        {!summary && (
          <button
            className={styles.generateBtn}
            disabled={isPending}
            onClick={() => generate()}
          >
            {generating ? "L'IA analyse..." : 'Generer le rapport'}
          </button>
        )}
      </div>

      {!summary && !isPending && (
        <div className={styles.prompt}>
          Generez un rapport de synthese IA pour cette mission.
        </div>
      )}

      {isPending && !summary && (
        <div className={styles.loading}>
          {"L'IA analyse votre mission..."}
        </div>
      )}

      {summary && (
        <>
          <div className={styles.content}>{summary.content}</div>
          <div className={styles.actions}>
            <button className={styles.iconBtn} onClick={handleCopy} title="Copier le rapport">
              Copier
            </button>
            <button className={styles.iconBtn} onClick={handleDownload} title="Telecharger en markdown">
              Telecharger .md
            </button>
            <button
              className={`${styles.iconBtn} ${styles.regenerateBtn}`}
              disabled={generating}
              onClick={() => generate()}
              title="Regenerer le rapport"
            >
              {generating ? 'Generation...' : 'Regenerer'}
            </button>
            <span className={styles.meta}>
              {new Date(summary.generatedAt).toLocaleString('fr-FR')}
            </span>
          </div>
        </>
      )}
    </div>
  )
}
