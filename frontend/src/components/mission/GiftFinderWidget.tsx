import { useState } from 'react'
import { useGiftSuggestions } from '../../hooks/useGiftSuggestions'
import type { GiftSuggestion, GiftSuggestionsParams } from '../../api/giftfinder'
import styles from './GiftFinderWidget.module.css'

const OCCASIONS = [
  'Anniversaire',
  'Noël',
  'Saint-Valentin',
  'Fête des Mères',
  'Fête des Pères',
]

const RECIPIENTS = ['Lui', 'Elle', 'Enfant', 'Couple']

const CATEGORY_ICON: Record<string, string> = {
  tech: '💻',
  mode: '👗',
  beauté: '💄',
  maison: '🏠',
  expérience: '🎭',
  gastronomie: '🍷',
  'jeux & jouets': '🎮',
  sport: '🏃',
}

function GiftCard({ gift }: { gift: GiftSuggestion }) {
  const icon = CATEGORY_ICON[gift.category] ?? '🎁'
  return (
    <article className={styles.giftCard}>
      <div className={styles.giftCardHeader}>
        <span className={styles.categoryIcon}>{icon}</span>
        <div className={styles.giftCardMeta}>
          <span className={styles.categoryLabel}>{gift.category}</span>
          {gift.score > 0 && (
            <span className={styles.scoreBadge} title="Pertinence pour cette mission">
              ★ {gift.score}
            </span>
          )}
        </div>
      </div>
      <h4 className={styles.giftName}>{gift.name}</h4>
      <p className={styles.giftDescription}>{gift.description}</p>
      <div className={styles.giftFooter}>
        <span className={styles.priceRange}>{gift.priceRange}</span>
        <div className={styles.platforms}>
          {gift.platforms.slice(0, 2).map((p) => (
            <span key={p} className={styles.platform}>
              {p}
            </span>
          ))}
        </div>
      </div>
    </article>
  )
}

interface GiftFinderWidgetProps {
  missionId: string
  budget: number
}

export function GiftFinderWidget({ missionId, budget }: GiftFinderWidgetProps) {
  const [occasion, setOccasion] = useState<string>('')
  const [recipient, setRecipient] = useState<string>('')
  const [maxPrice, setMaxPrice] = useState<number>(budget)

  const params: GiftSuggestionsParams = {
    occasion: occasion || undefined,
    recipient: recipient || undefined,
    maxPrice: maxPrice > 0 ? maxPrice : undefined,
  }

  const { data, isLoading, isError } = useGiftSuggestions(missionId, params)

  function toggleOccasion(o: string) {
    setOccasion((prev) => (prev === o ? '' : o))
  }

  function toggleRecipient(r: string) {
    setRecipient((prev) => (prev === r ? '' : r))
  }

  return (
    <section className={styles.widget}>
      <div className={styles.filters}>
        <div className={styles.filterGroup}>
          <span className={styles.filterLabel}>Occasion</span>
          <div className={styles.pills}>
            {OCCASIONS.map((o) => (
              <button
                key={o}
                className={`${styles.pill} ${occasion === o ? styles.pillActive : ''}`}
                onClick={() => toggleOccasion(o)}
                type="button"
              >
                {o}
              </button>
            ))}
          </div>
        </div>

        <div className={styles.filterGroup}>
          <span className={styles.filterLabel}>Destinataire</span>
          <div className={styles.pills}>
            {RECIPIENTS.map((r) => (
              <button
                key={r}
                className={`${styles.pill} ${recipient === r ? styles.pillActive : ''}`}
                onClick={() => toggleRecipient(r)}
                type="button"
              >
                {r}
              </button>
            ))}
          </div>
        </div>

        <div className={styles.filterGroup}>
          <label className={styles.filterLabel} htmlFor="gift-max-price">
            Budget max : <strong>{maxPrice.toFixed(0)} €</strong>
          </label>
          <input
            id="gift-max-price"
            type="range"
            min={10}
            max={Math.max(budget, 1500)}
            step={10}
            value={maxPrice}
            onChange={(e) => setMaxPrice(Number(e.target.value))}
            className={styles.priceSlider}
          />
        </div>
      </div>

      {isLoading && (
        <p className={styles.status}>Recherche des meilleures idées cadeaux…</p>
      )}
      {isError && (
        <p className={styles.statusError}>
          Impossible de charger les suggestions. Réessayez plus tard.
        </p>
      )}

      {data && data.suggestions.length === 0 && (
        <p className={styles.status}>
          Aucune idée cadeau ne correspond à ces critères. Essayez d'ajuster les filtres.
        </p>
      )}

      {data && data.suggestions.length > 0 && (
        <>
          <p className={styles.resultCount}>
            {data.totalFound} idée{data.totalFound > 1 ? 's' : ''} trouvée
            {data.totalFound > 1 ? 's' : ''}
          </p>
          <div className={styles.grid}>
            {data.suggestions.map((gift) => (
              <GiftCard key={gift.id} gift={gift} />
            ))}
          </div>
        </>
      )}
    </section>
  )
}
