import { getOptionsExportURL } from '../../api/options'
import styles from './ExportButton.module.css'

interface ExportButtonProps {
  slug: string
}

export function ExportButton({ slug }: ExportButtonProps) {
  const downloadUrl = getOptionsExportURL(slug)

  return (
    <a
      href={downloadUrl}
      download={`options-${slug}.csv`}
      className={styles.button}
      title="Export options as CSV"
    >
      <span aria-hidden="true">📥</span> Exporter CSV
    </a>
  )
}
