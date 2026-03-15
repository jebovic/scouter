import styles from './MissionCSVExportButton.module.css'

interface MissionCSVExportButtonProps {
  missionId: string
}

export function MissionCSVExportButton({ missionId }: MissionCSVExportButtonProps) {
  const exportURL = `/api/missions/${missionId}/export/csv`

  return (
    <a
      href={exportURL}
      download
      className={styles.button}
      title="Exporter les articles en CSV"
      aria-label="Exporter les articles de la mission en CSV"
    >
      📥 Export CSV
    </a>
  )
}
