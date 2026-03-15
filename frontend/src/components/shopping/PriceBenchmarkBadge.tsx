import { useEffect, useRef, useState } from 'react'
import { usePriceBenchmark } from '../../hooks/useBenchmark'
import styles from './PriceBenchmarkBadge.module.css'

interface PriceBenchmarkBadgeProps {
  itemId: string
  itemName: string
}

const VERDICT_ICON: Record<string, string> = {
  good_deal: '🟢',
  average: '🟡',
  overpriced: '🔴',
}

const VERDICT_LABEL: Record<string, string> = {
  good_deal: 'Bon prix',
  average: 'Moyen',
  overpriced: 'Surévalué',
}

const VERDICT_CHIP_CLASS: Record<string, string> = {
  good_deal: styles.chipGoodDeal,
  average: styles.chipAverage,
  overpriced: styles.chipOverpriced,
}

function formatDiff(diffPercent: number): string {
  const abs = Math.abs(diffPercent).toLocaleString('fr-FR', { maximumFractionDigits: 1 })
  return diffPercent < 0 ? `-${abs}%` : `+${abs}%`
}

function formatPrice(price: number): string {
  return price.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })
}

export function PriceBenchmarkBadge({ itemId, itemName: _itemName }: PriceBenchmarkBadgeProps) {
  const [enabled, setEnabled] = useState(false)
  const [open, setOpen] = useState(false)
  const wrapperRef = useRef<HTMLDivElement>(null)

  const { data, isFetching } = usePriceBenchmark(itemId, enabled)

  // Close popover on outside click.
  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  function handleTrigger() {
    setEnabled(true)
    setOpen((v) => !v)
  }

  const diffClass = data
    ? data.diffPercent < 0
      ? styles.diffNeg
      : styles.diffPos
    : ''

  const chipClass = data ? VERDICT_CHIP_CLASS[data.verdict] ?? '' : ''

  return (
    <div ref={wrapperRef} className={styles.wrapper}>
      {/* If data is loaded, show the inline verdict chip; otherwise show the trigger icon */}
      {data ? (
        <button
          className={`${styles.chip} ${chipClass}`}
          onClick={handleTrigger}
          aria-expanded={open}
          aria-label={`Benchmark marché: ${VERDICT_LABEL[data.verdict]}`}
          title="Comparer au marché français"
        >
          <span aria-hidden="true">{VERDICT_ICON[data.verdict]}</span>
          {VERDICT_LABEL[data.verdict]}
          <span className={`${styles.diff} ${diffClass}`}>
            {formatDiff(data.diffPercent)}
          </span>
        </button>
      ) : (
        <button
          className={styles.triggerBtn}
          onClick={handleTrigger}
          aria-expanded={open}
          aria-label="Comparer au marché français"
          title="Comparer au marché français"
        >
          {isFetching ? <span className={styles.loadingDot} /> : '📊'}
        </button>
      )}

      {open && (
        <div className={styles.popover} role="tooltip">
          {isFetching && !data ? (
            <p className={styles.explanation}>Analyse en cours...</p>
          ) : data ? (
            <>
              <p className={styles.popoverTitle}>Marché français estimé</p>
              <div className={styles.rangeRow}>
                <div className={styles.rangeItem}>
                  <span className={styles.rangeLabel}>Bas</span>
                  <span className={styles.rangeValue}>{formatPrice(data.marketRange.low)}</span>
                </div>
                <div className={styles.rangeItem}>
                  <span className={styles.rangeLabel}>Moy</span>
                  <span className={styles.rangeValue}>{formatPrice(data.marketRange.average)}</span>
                </div>
                <div className={styles.rangeItem}>
                  <span className={styles.rangeLabel}>Haut</span>
                  <span className={styles.rangeValue}>{formatPrice(data.marketRange.high)}</span>
                </div>
              </div>
              <p className={styles.explanation}>{data.explanation}</p>
            </>
          ) : null}
        </div>
      )}
    </div>
  )
}
