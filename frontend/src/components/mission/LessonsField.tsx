import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import styles from './LessonsField.module.css'

interface LessonsFieldProps {
  missionSlug: string
  value?: string | null
  onSave: (lessons: string) => Promise<void>
}

export function LessonsField({ value, onSave }: LessonsFieldProps) {
  const { t } = useTranslation()
  const [editing, setEditing] = useState(false)
  const [text, setText] = useState(value ?? '')
  const [saving, setSaving] = useState(false)

  async function handleBlur() {
    if (text === (value ?? '')) {
      setEditing(false)
      return
    }
    setSaving(true)
    try {
      await onSave(text)
    } finally {
      setSaving(false)
      setEditing(false)
    }
  }

  return (
    <div className={styles.wrap}>
      <h4 className={styles.title}>{t('lessons.title')}</h4>
      {editing ? (
        <textarea
          className={styles.textarea}
          value={text}
          onChange={(e) => setText(e.target.value)}
          onBlur={handleBlur}
          autoFocus
          maxLength={2000}
          rows={4}
          placeholder={t('lessons.placeholder')}
        />
      ) : (
        <div
          className={styles.display}
          onClick={() => setEditing(true)}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => e.key === 'Enter' && setEditing(true)}
        >
          {text || <span className={styles.placeholder}>{t('lessons.addPlaceholder')}</span>}
        </div>
      )}
      {saving && <span className={styles.saving}>{t('lessons.saving')}</span>}
    </div>
  )
}
