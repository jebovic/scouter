import { getMissionReportUrl } from '../../api/missionreport'
import styles from './MissionReportButton.module.css'

interface MissionReportButtonProps {
  missionId: string
}

export function MissionReportButton({ missionId }: MissionReportButtonProps) {
  function handleDownload() {
    window.open(getMissionReportUrl(missionId), '_blank')
  }

  return (
    <button onClick={handleDownload} className={styles.btn} title="Télécharger le rapport de mission">
      📄 Rapport
    </button>
  )
}
