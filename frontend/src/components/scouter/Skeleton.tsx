import styles from './Skeleton.module.css'

type SkeletonVariant = 'card' | 'row' | 'chart' | 'text' | 'circle'

interface SkeletonProps {
  variant: SkeletonVariant
  width?: string | number
  height?: string | number
  className?: string
  'data-testid'?: string
}

export function Skeleton({ variant, width, height, className, 'data-testid': testId }: SkeletonProps) {
  const style: React.CSSProperties = {}
  if (width !== undefined) style.width = typeof width === 'number' ? `${width}px` : width
  if (height !== undefined) style.height = typeof height === 'number' ? `${height}px` : height

  return (
    <div
      className={[styles.shimmer, styles[variant], className].filter(Boolean).join(' ')}
      style={style}
      aria-hidden="true"
      data-testid={testId}
    />
  )
}

interface SkeletonGridProps {
  count?: number
}

export function SkeletonGrid({ count = 6 }: SkeletonGridProps) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <Skeleton key={i} variant="card" />
      ))}
    </>
  )
}
