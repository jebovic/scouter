import { useState, useCallback } from 'react'
import { Toast } from './Toast'
import styles from './ToastContainer.module.css'
import type { BudgetAlert } from '../../utils/budgetAlerts'

export function useToasts() {
  const [toasts, setToasts] = useState<BudgetAlert[]>([])

  const addToast = useCallback((alert: BudgetAlert) => {
    setToasts((prev) => [...prev, alert])
  }, [])

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  return { toasts, addToast, removeToast }
}

interface ToastContainerProps {
  toasts: BudgetAlert[]
  onRemove: (id: string) => void
}

export function ToastContainer({ toasts, onRemove }: ToastContainerProps) {
  if (toasts.length === 0) return null

  return (
    <div className={styles.container}>
      {toasts.map((t) => (
        <Toast key={t.id} alert={t} onDismiss={() => onRemove(t.id)} />
      ))}
    </div>
  )
}
