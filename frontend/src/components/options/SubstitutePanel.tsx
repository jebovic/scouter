import { useState } from 'react'
import { useSubstitutes, useFetchSubstitutes } from '../../hooks/useSubstitutes'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { Substitute } from '../../api/substitutes'
import styles from './SubstitutePanel.module.css'

interface Props {
  optionId: string
}

type Advantage = Substitute['advantage']

const ADVANTAGE_LABELS: Record<Advantage, string> = {
  cheaper: '🏷️ Prix bas',
  better_rated: '⭐ Mieux noté',
  eco_friendly: '🌱 Éco',
  local: '🇫🇷 Français',
}

const ADVANTAGE_CLASS: Record<Advantage, string> = {
  cheaper: styles.advantageCheaper,
  better_rated: styles.advantageBetterRated,
  eco_friendly: styles.advantageEcoFriendly,
  local: styles.advantageLocal,
}

function SubstituteItem({ sub }: { sub: Substitute }) {
  const { fmt } = useFormatCurrency()
  return (
    <div className={styles.item}>
      <div className={styles.itemTop}>
        <div className={styles.nameGroup}>
          <span className={styles.name}>{sub.name}</span>
          <span className={styles.brand}>{sub.brand}</span>
        </div>
        <div className={styles.priceGroup}>
          <span className={styles.price}>{fmt(sub.price)}</span>
          <span className={`${styles.advantageBadge} ${ADVANTAGE_CLASS[sub.advantage]}`}>
            {ADVANTAGE_LABELS[sub.advantage]}
          </span>
        </div>
      </div>

      <span className={styles.retailerBadge}>{sub.retailer}</span>
      <p className={styles.reason}>{sub.reason}</p>

      {sub.url && (
        <a
          href={sub.url}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.buyLink}
        >
          Voir chez {sub.retailer}
        </a>
      )}
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className={styles.skeleton}>
      {[1, 2].map((i) => (
        <div key={i} className={styles.skeletonLine} style={{ width: i === 1 ? '80%' : '60%' }} />
      ))}
    </div>
  )
}

export function SubstitutePanel({ optionId }: Props) {
  const [open, setOpen] = useState(false)

  const { data, isLoading } = useSubstitutes(open ? optionId : undefined)
  const { fetch, isPending } = useFetchSubstitutes(optionId)

  const handleOpen = () => {
    setOpen(true)
    fetch()
  }

  if (!open) {
    return (
      <div className={styles.panel}>
        <button className={styles.triggerBtn} onClick={handleOpen}>
          Trouver des alternatives
        </button>
      </div>
    )
  }

  const loading = isLoading || isPending

  return (
    <div className={styles.panel}>
      <div className={styles.content}>
        <div className={styles.header}>Alternatives disponibles en France</div>

        {loading && <LoadingSkeleton />}

        {!loading && data && data.substitutes.length === 0 && (
          <p className={styles.empty}>Aucune alternative trouvée</p>
        )}

        {!loading && data && data.substitutes.length > 0 && (
          <div className={styles.list}>
            {data.substitutes.map((sub, idx) => (
              <SubstituteItem key={`${sub.name}-${idx}`} sub={sub} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
