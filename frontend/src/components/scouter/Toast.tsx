import { useEffect, useState } from 'react'
import styles from './Toast.module.css'
import type { BudgetAlert } from '../../utils/budgetAlerts'

interface ToastProps {
  alert: BudgetAlert
  onDismiss: () => void
}

export function Toast({ alert, onDismiss }: ToastProps) {
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    const t = setTimeout(() => {
      setVisible(false)
      setTimeout(onDismiss, 300)
    }, 5000)
    return () => clearTimeout(t)
  }, [onDismiss])

  return (
    <div className={`${styles.toast} ${styles[alert.severity]} ${!visible ? styles.exit : ''}`}>
      <div className={styles.title}>{alert.title}</div>
      <div className={styles.message}>{alert.message}</div>
      <button className={styles.close} onClick={() => { setVisible(false); setTimeout(onDismiss, 300) }}>×</button>
    </div>
  )
}
