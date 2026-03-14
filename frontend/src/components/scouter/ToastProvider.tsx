import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import styles from './ToastProvider.module.css'

export type ToastVariant = 'success' | 'error' | 'info'

export interface Toast {
  id: string
  message: string
  variant: ToastVariant
}

interface ToastContextValue {
  toast: (message: string, variant?: ToastVariant) => void
}

const TOAST_DURATION_MS = 4000

const ToastContext = createContext<ToastContextValue | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const nextId = useRef(0)
  const timers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())

  useEffect(() => {
    const t = timers.current
    return () => { t.forEach(clearTimeout) }
  }, [])

  const dismiss = useCallback((id: string) => {
    clearTimeout(timers.current.get(id))
    timers.current.delete(id)
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const toast = useCallback((message: string, variant: ToastVariant = 'info') => {
    const id = String(++nextId.current)
    setToasts((prev) => [...prev, { id, message, variant }])
    const handle = setTimeout(() => {
      timers.current.delete(id)
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, TOAST_DURATION_MS)
    timers.current.set(id, handle)
  }, [])

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className={styles.container}>
        {toasts.map((t) => (
          <ToastItem key={t.id} toast={t} onDismiss={dismiss} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

const variantColor: Record<ToastVariant, string> = {
  success: 'var(--cyan)',
  error: 'var(--coral)',
  info: 'var(--text-mid)',
}

function ToastItem({ toast: t, onDismiss }: { toast: Toast; onDismiss: (id: string) => void }) {
  const color = variantColor[t.variant]
  return (
    <div
      className={styles.toast}
      style={{
        '--toast-color': color,
        '--toast-glow': `${color}33`,
      } as React.CSSProperties}
    >
      <span className={styles.toastMessage}>{t.message}</span>
      <button
        onClick={() => onDismiss(t.id)}
        className={styles.dismissBtn}
        aria-label="Dismiss"
      >
        ×
      </button>
    </div>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used inside ToastProvider')
  return ctx
}
