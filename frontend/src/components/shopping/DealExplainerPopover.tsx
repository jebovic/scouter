import { useEffect, useRef, useState } from 'react'
import { useDealExplain } from '../../hooks/useDealExplain'
import styles from './DealExplainerPopover.module.css'

interface DealExplainerPopoverProps {
  itemId: string
  dealScore: number
  trend?: string
}

export function DealExplainerPopover({ itemId, dealScore, trend }: DealExplainerPopoverProps) {
  const [open, setOpen] = useState(false)
  const wrapperRef = useRef<HTMLDivElement>(null)

  // Only fetch when the user opens the popover for the first time.
  const { data, isFetching } = useDealExplain(
    open ? itemId : null,
    dealScore,
    trend,
  )

  // Close on outside click.
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

  return (
    <div ref={wrapperRef} className={styles.wrapper}>
      <button
        className={styles.triggerBtn}
        aria-expanded={open}
        aria-label="Pourquoi c'est une bonne affaire ?"
        title="Pourquoi c'est une bonne affaire ?"
        onClick={() => setOpen((v) => !v)}
      >
        ?
      </button>

      {open && (
        <div className={styles.popover} role="tooltip">
          {isFetching && !data ? (
            <div className={styles.loading}>
              <span className={styles.dot} />
              Analyse en cours...
            </div>
          ) : data ? (
            <>
              <p className={styles.explanation}>{data.explanation}</p>
              {data.tags.length > 0 && (
                <div className={styles.tags}>
                  {data.tags.map((tag) => (
                    <span key={tag} className={styles.tag}>
                      {tag}
                    </span>
                  ))}
                </div>
              )}
            </>
          ) : null}
        </div>
      )}
    </div>
  )
}
