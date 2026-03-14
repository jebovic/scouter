import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react'

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
      <div
        style={{
          position: 'fixed',
          bottom: '1.5rem',
          right: '1.5rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '0.5rem',
          zIndex: 9999,
          maxWidth: 360,
          pointerEvents: 'none',
        }}
      >
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
  return (
    <div
      style={{
        background: 'var(--surface)',
        border: `1px solid ${variantColor[t.variant]}`,
        borderRadius: 'var(--radius-sm)',
        padding: '0.75rem 1rem',
        fontFamily: 'var(--font-mono)',
        fontSize: '0.8rem',
        color: variantColor[t.variant],
        boxShadow: `0 0 12px ${variantColor[t.variant]}33`,
        animation: 'fade-in 0.2s ease both',
        pointerEvents: 'auto',
        display: 'flex',
        alignItems: 'flex-start',
        gap: '0.5rem',
      }}
    >
      <span style={{ flex: 1 }}>{t.message}</span>
      <button
        onClick={() => onDismiss(t.id)}
        style={{
          background: 'none',
          border: 'none',
          color: variantColor[t.variant],
          cursor: 'pointer',
          padding: 0,
          fontSize: '1rem',
          lineHeight: 1,
          opacity: 0.7,
          flexShrink: 0,
        }}
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
