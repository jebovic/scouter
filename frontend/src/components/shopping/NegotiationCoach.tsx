import { useEffect, useRef, useState } from 'react'
import { useNegotiationAdvice } from '../../hooks/useNegotiationCoach'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { NegotiationTip } from '../../api/negotiation_coach'
import styles from './NegotiationCoach.module.css'

interface NegotiationCoachProps {
  itemId: string
}

const DIFFICULTY_LABEL: Record<string, string> = {
  easy: 'Facile',
  medium: 'Moyen',
  hard: 'Difficile',
}

const DIFFICULTY_CLASS: Record<string, string> = {
  easy: styles.badgeEasy,
  medium: styles.badgeMedium,
  hard: styles.badgeHard,
}

function TipRow({ tip }: { tip: NegotiationTip }) {
  const badgeClass = DIFFICULTY_CLASS[tip.difficulty] ?? ''
  return (
    <li className={styles.tipItem}>
      <div className={styles.tipHeader}>
        <span className={`${styles.badge} ${badgeClass}`}>
          {DIFFICULTY_LABEL[tip.difficulty] ?? tip.difficulty}
        </span>
        <span className={styles.tipTitle}>{tip.title}</span>
      </div>
      <p className={styles.tipDescription}>{tip.description}</p>
    </li>
  )
}

export function NegotiationCoach({ itemId }: NegotiationCoachProps) {
  const [enabled, setEnabled] = useState(false)
  const [open, setOpen] = useState(false)
  const [copied, setCopied] = useState(false)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const { fmt } = useFormatCurrency()

  const { data, isFetching } = useNegotiationAdvice(itemId, enabled)

  // Close panel on outside click.
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

  function handleCopy() {
    if (!data?.script) return
    navigator.clipboard.writeText(data.script).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <div ref={wrapperRef} className={styles.wrapper}>
      <button
        className={styles.triggerBtn}
        onClick={handleTrigger}
        aria-expanded={open}
        aria-label="Coach Negociation"
        title="Coach Negociation"
      >
        {isFetching ? <span className={styles.spinner} aria-hidden="true" /> : 'Negocier'}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Conseils de negociation">
          {isFetching && !data ? (
            <p className={styles.targetPrice}>Analyse en cours...</p>
          ) : data ? (
            <>
              <p className={styles.targetPrice}>
                Objectif: {fmt(data.targetPrice)}
              </p>

              <ul className={styles.tipsList}>
                {data.tips.map((tip, i) => (
                  <TipRow key={i} tip={tip} />
                ))}
              </ul>

              <div className={styles.scriptSection}>
                <p className={styles.scriptLabel}>Script pret a l&apos;emploi</p>
                <p className={styles.scriptText}>&ldquo;{data.script}&rdquo;</p>
                <button
                  className={styles.copyBtn}
                  onClick={handleCopy}
                  title="Copier le script"
                >
                  {copied ? 'Copie !' : 'Copier'}
                </button>
              </div>

              <button
                className={styles.closeBtn}
                onClick={() => setOpen(false)}
                aria-label="Fermer le coach negociation"
              >
                Fermer
              </button>
            </>
          ) : null}
        </div>
      )}
    </div>
  )
}
