import type { ReactNode } from 'react'

interface ScouterGridProps {
  children: ReactNode
  cols?: number
  gap?: number
}

export function ScouterGrid({ children, cols = 3, gap = 16 }: ScouterGridProps) {
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: `repeat(auto-fill, minmax(${Math.floor(320 / cols)}px, 1fr))`,
        gap,
      }}
    >
      {children}
    </div>
  )
}
