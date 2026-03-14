import { useState, useEffect, useRef } from 'react'
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
  const modalRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!isPending) setSubmitted(false)
  }, [isPending])

  // Focus trap: cycle Tab/Shift+Tab among focusable elements
  useEffect(() => {
    const el = modalRef.current
    if (!el) return
    const focusable = el.querySelectorAll<HTMLElement>(
      'button:not(:disabled), textarea, input, [tabindex]:not([tabindex="-1"])'
    )
    if (focusable.length === 0) return
    const first = focusable[0]
    const last = focusable[focusable.length - 1]

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') { onClose(); return }
      if (e.key !== 'Tab') return
      if (e.shiftKey) {
        if (document.activeElement === first) { e.preventDefault(); last.focus() }
      } else {
        if (document.activeElement === last) { e.preventDefault(); first.focus() }
      }
    }

    el.addEventListener('keydown', handleKeyDown)
    return () => el.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  return (
    <div
      className={styles.overlay}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        ref={modalRef}
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
