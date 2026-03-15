import styles from './StarRating.module.css'

interface StarRatingProps {
  value: number | null | undefined
  onChange?: (v: number) => void
  readOnly?: boolean
  size?: number
}

export function StarRating({ value, onChange, readOnly = false, size = 20 }: StarRatingProps) {
  return (
    <div className={styles.stars} role={readOnly ? undefined : 'group'} aria-label="Rating">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          className={`${styles.star} ${(value ?? 0) >= star ? styles.filled : ''}`}
          style={{ '--size': `${size}px` } as React.CSSProperties}
          onClick={readOnly ? undefined : () => onChange?.(star)}
          disabled={readOnly}
          aria-label={`${star} star${star !== 1 ? 's' : ''}`}
          tabIndex={readOnly ? -1 : 0}
        >
          ★
        </button>
      ))}
    </div>
  )
}
