import { useState } from 'react'
import styles from './LessonsField.module.css'

interface LessonsFieldProps {
  missionSlug: string
  value?: string | null
  onSave: (lessons: string) => Promise<void>
}

export function LessonsField({ value, onSave }: LessonsFieldProps) {
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
      <h4 className={styles.title}>Lessons Learned</h4>
      {editing ? (
        <textarea
          className={styles.textarea}
          value={text}
          onChange={(e) => setText(e.target.value)}
          onBlur={handleBlur}
          autoFocus
          maxLength={2000}
          rows={4}
          placeholder="What did you learn from this mission? What would you do differently next time?"
        />
      ) : (
        <div
          className={styles.display}
          onClick={() => setEditing(true)}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => e.key === 'Enter' && setEditing(true)}
        >
          {text || <span className={styles.placeholder}>Click to add lessons learned...</span>}
        </div>
      )}
      {saving && <span className={styles.saving}>Saving...</span>}
    </div>
  )
}
