import { useState } from 'react'
import { useNegotiationScriptContextual } from '../../hooks/useNegotiationScriptContextual'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { ScriptSections } from '../../api/negotiationscript'
import styles from './NegotiationScriptCard.module.css'

interface NegotiationScriptCardProps {
  missionId: string
  itemId: string
}

const SECTION_KEYS: Array<{ key: keyof ScriptSections; label: string; icon: string }> = [
  { key: 'opening', label: 'Ouverture', icon: '👋' },
  { key: 'argument', label: 'Argument', icon: '📣' },
  { key: 'counterOffer', label: 'Contre-offre', icon: '💰' },
  { key: 'closing', label: 'Conclusion', icon: '🤝' },
]

interface AccordionSectionProps {
  icon: string
  label: string
  text: string
}

function AccordionSection({ icon, label, text }: AccordionSectionProps) {
  const [open, setOpen] = useState(false)
  const [copied, setCopied] = useState(false)

  function handleCopy() {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }

  return (
    <div className={styles.section}>
      <button
        className={styles.sectionHeader}
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label={`${open ? 'Fermer' : 'Ouvrir'} la section ${label}`}
      >
        <span className={styles.sectionLabel}>
          <span aria-hidden="true">{icon}</span>
          {label}
        </span>
        <span className={`${styles.chevron} ${open ? styles.chevronOpen : ''}`} aria-hidden="true">
          ▼
        </span>
      </button>
      {open && (
        <div className={styles.sectionBody}>
          <p className={styles.sectionText}>&ldquo;{text}&rdquo;</p>
          <button
            className={styles.copyBtn}
            onClick={handleCopy}
            aria-label={`Copier la section ${label}`}
          >
            {copied ? '✓ Copié' : '📋 Copier'}
          </button>
        </div>
      )}
    </div>
  )
}

export function NegotiationScriptCard({ missionId, itemId }: NegotiationScriptCardProps) {
  const { data, isLoading } = useNegotiationScriptContextual(missionId, itemId)
  const { fmt } = useFormatCurrency()
  const [allCopied, setAllCopied] = useState(false)

  if (isLoading) return null
  if (!data) return null

  function handleCopyAll() {
    if (!data) return
    const full = SECTION_KEYS
      .map(({ label, key }) => `[${label}]\n${data.script[key]}`)
      .join('\n\n')
    navigator.clipboard.writeText(full).then(() => {
      setAllCopied(true)
      setTimeout(() => setAllCopied(false), 1800)
    })
  }

  const savingsFormatted =
    data.estimatedSavings > 0
      ? fmt(data.estimatedSavings)
      : null

  return (
    <div className={styles.card} role="region" aria-label="Script de négociation contextuel">
      <div className={styles.header}>
        <span className={styles.title}>Script de négociation — {data.name}</span>
        {savingsFormatted && (
          <span className={styles.savings} aria-label={`Économie potentielle : ${savingsFormatted}`}>
            Économie potentielle&nbsp;: {savingsFormatted}
          </span>
        )}
      </div>

      {SECTION_KEYS.map(({ key, label, icon }) => (
        <AccordionSection
          key={key}
          icon={icon}
          label={label}
          text={data.script[key]}
        />
      ))}

      {data.tips.length > 0 && (
        <section className={styles.tipsSection} aria-label="Conseils de négociation">
          <h4 className={styles.tipsTitle}>Conseils</h4>
          <ul className={styles.tipsList}>
            {data.tips.map((tip, i) => (
              <li key={i} className={styles.tipItem}>
                <span className={styles.tipIcon} aria-hidden="true">💡</span>
                {tip}
              </li>
            ))}
          </ul>
        </section>
      )}

      <button
        className={styles.copyAllBtn}
        onClick={handleCopyAll}
        aria-label="Copier le script complet"
      >
        {allCopied ? '✓ Script copié !' : '📋 Copier le script complet'}
      </button>
    </div>
  )
}
