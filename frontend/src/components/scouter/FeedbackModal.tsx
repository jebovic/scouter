import { useState, useEffect } from 'react'
import styles from './FeedbackModal.module.css'

interface FeedbackModalProps {
  title: string
  placeholder?: string
  onConfirm: (feedback: string) => void
  onClose: () => void
  isPending?: boolean
}

export function FeedbackModal({ title, placeholder, onConfirm, onClose, isPending }: FeedbackModalProps) {
  const [feedback, setFeedback] = useState('')
  const [submitted, setSubmitted] = useState(false)

  useEffect(() => {
    if (!isPending) setSubmitted(false)
  }, [isPending])

  return (
    <div
      className={styles.overlay}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className={styles.modal}
        role="dialog"
        aria-modal="true"
        aria-labelledby="feedback-modal-title"
      >
        <h3 id="feedback-modal-title" className={styles.title}>{title}</h3>
        <label className={styles.label}>Feedback for the agent (optional)</label>
        <textarea
          className={styles.textarea}
          placeholder={placeholder ?? 'Add optional guidance for the agent...'}
          value={feedback}
          onChange={(e) => setFeedback(e.target.value)}
          onKeyDown={(e) => e.key === 'Escape' && onClose()}
          rows={4}
          autoFocus
        />
        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onClose} disabled={isPending}>
            Cancel
          </button>
          <button
            className={styles.confirmBtn}
            onClick={() => { if (!submitted && !isPending) { setSubmitted(true); onConfirm(feedback) } }}
            disabled={isPending || submitted}
          >
            {isPending ? 'Running...' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  )
}
