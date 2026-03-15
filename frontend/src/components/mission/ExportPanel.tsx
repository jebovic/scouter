import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { downloadMissionExport } from '../../api/export'
import styles from './ExportPanel.module.css'

interface Props {
  missionId: string
}

type Format = 'json' | 'markdown' | 'pdf'

const FORMATS: { id: Format; label: string; icon: string }[] = [
  { id: 'pdf', label: 'PDF', icon: '📄' },
  { id: 'markdown', label: 'Markdown', icon: '📝' },
  { id: 'json', label: 'JSON', icon: '{}' },
]

export function ExportPanel({ missionId }: Props) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState<Format | null>(null)
  const [error, setError] = useState<string | null>(null)

  async function handleExport(format: Format) {
    setLoading(format)
    setError(null)
    try {
      await downloadMissionExport(missionId, format)
    } catch {
      setError(t('export.error'))
    } finally {
      setLoading(null)
    }
  }

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <span className={styles.title}>{t('export.title')}</span>
      </div>
      <div className={styles.buttons}>
        {FORMATS.map((f) => (
          <button
            key={f.id}
            className={styles.exportBtn}
            disabled={loading !== null}
            onClick={() => handleExport(f.id)}
          >
            <span className={styles.icon}>{f.icon}</span>
            {loading === f.id ? t('export.downloading') : f.label}
          </button>
        ))}
      </div>
      {error && <div className={styles.error}>{error}</div>}
    </div>
  )
}
