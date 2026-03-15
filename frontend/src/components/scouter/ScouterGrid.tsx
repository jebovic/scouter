import type { ReactNode } from 'react'
import styles from './ScouterGrid.module.css'

interface ScouterGridProps {
  children: ReactNode
  cols?: number
  gap?: number
}

export function ScouterGrid({ children, cols = 3, gap = 16 }: ScouterGridProps) {
  return (
    <div
      className={styles.grid}
      style={{ '--grid-gap': `${gap}px`, '--grid-cols': cols } as React.CSSProperties}
    >
      {children}
    </div>
  )
}
