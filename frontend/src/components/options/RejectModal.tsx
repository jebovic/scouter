import { useState } from 'react'
import styles from './RejectModal.module.css'

interface RejectModalProps {
  optionName: string
  onConfirm: (reason: string) => void
  onCancel: () => void
}

export function RejectModal({ optionName, onConfirm, onCancel }: RejectModalProps) {
  const [reason, setReason] = useState('')

  return (
    <div
      className={styles.overlay}
      onClick={(e) => e.target === e.currentTarget && onCancel()}
    >
      <div className={styles.modal}>
        <h3 className={styles.title}>REJECT OPTION</h3>
        <p className={styles.optionName}>{optionName}</p>
        <label className={styles.label}>Reason (optional)</label>
        <textarea
          className={styles.textarea}
          placeholder="Why are you rejecting this option?"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          rows={3}
          autoFocus
        />
        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onCancel}>
            Cancel
          </button>
          <button className={styles.confirmBtn} onClick={() => onConfirm(reason)}>
            Reject
          </button>
        </div>
      </div>
    </div>
  )
}
