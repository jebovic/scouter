import type { ReactNode } from 'react'
import styles from './ScouterGrid.module.css'

interface ScouterGridProps {
  children: ReactNode
  cols?: number
  gap?: number
}

export function ScouterGrid({ children, cols: _cols = 3, gap = 16 }: ScouterGridProps) {
  return (
    <div
      className={styles.grid}
      style={{ '--grid-gap': `${gap}px` } as React.CSSProperties}
    >
      {children}
    </div>
  )
}
