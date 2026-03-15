import styles from './CompareBar.module.css'

interface CompareBarProps {
  isCompareMode: boolean
  selectedCount: number
  onCompare: () => void
  onClear: () => void
}

export function CompareBar({ isCompareMode, selectedCount, onCompare, onClear }: CompareBarProps) {
  if (!isCompareMode || selectedCount === 0) return null

  const label =
    selectedCount === 1
      ? '1 option sélectionnée'
      : `${selectedCount} options sélectionnées`

  return (
    <div className={styles.bar} role="region" aria-label="Compare bar">
      <span className={styles.info}>
        <span className={styles.count}>{selectedCount}</span>
        {' '}
        {selectedCount === 1 ? 'option sélectionnée' : 'options sélectionnées'}
      </span>
      <div className={styles.actions}>
        <button
          className={styles.compareBtn}
          onClick={onCompare}
          disabled={selectedCount < 2}
          aria-label="Comparer"
          aria-describedby={selectedCount < 2 ? 'compare-hint' : undefined}
        >
          Comparer
        </button>
        {selectedCount < 2 && (
          <span id="compare-hint" className={styles.info} style={{ fontSize: '0.75rem' }}>
            (min. 2)
          </span>
        )}
        <button className={styles.clearBtn} onClick={onClear} aria-label="Effacer">
          Effacer
        </button>
      </div>
      {/* visually hidden accessible label for screen readers */}
      <span className="sr-only">{label}</span>
    </div>
  )
}
