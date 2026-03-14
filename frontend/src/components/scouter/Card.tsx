import type { CSSProperties, ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
  style?: CSSProperties
  glow?: boolean
  onClick?: () => void
}

export function Card({ children, className = '', style, glow, onClick }: CardProps) {
  const interactiveClass = onClick ? 'card-interactive' : ''
  const glowClass = glow ? 'card-glow' : ''

  return (
    <div
      onClick={onClick}
      style={{
        background: 'var(--surface)',
        border: `1px solid ${glow ? 'var(--border-glow)' : 'var(--border)'}`,
        borderRadius: 'var(--radius)',
        padding: 'var(--gap-lg)',
        boxShadow: glow ? '0 0 20px var(--cyan-dim)' : undefined,
        ...style,
      }}
      className={`${interactiveClass} ${glowClass} ${className}`.trim()}
    >
      {children}
    </div>
  )
}
