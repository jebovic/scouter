import { useItemTags } from '../../hooks/useItemTags'
import styles from './ItemTagsDisplay.module.css'

interface ItemTagsDisplayProps {
  missionId: string
  itemId: string
}

export function ItemTagsDisplay({ missionId, itemId }: ItemTagsDisplayProps) {
  const { data, isLoading, isError } = useItemTags(missionId, itemId)

  if (isLoading || isError || !data) return null

  // Only show brand tag if it's not already the first tag
  const extraTags = data.brand
    ? data.tags.filter((t) => t.toLowerCase() !== data.brand.toLowerCase())
    : data.tags

  return (
    <div className={styles.wrap}>
      {/* Suggested category */}
      {data.suggestedCategory && data.suggestedCategory !== 'Autre' && (
        <span className={styles.categoryPill}>
          🏷 {data.suggestedCategory}
        </span>
      )}

      {/* Brand badge */}
      {data.brand && (
        <span className={styles.brandBadge}>
          {data.brand}
        </span>
      )}

      {/* Special badges */}
      {data.isLuxury && (
        <span className={styles.luxuryBadge}>⭐ Luxe</span>
      )}
      {data.isTech && !data.isLuxury && (
        <span className={styles.techBadge}>💻 Tech</span>
      )}
      {data.refurbishedAvailable && (
        <span className={styles.refurbishedBadge}>🔧 Reconditionné dispo</span>
      )}

      {/* Tags (up to 4, excluding brand) */}
      {extraTags.slice(0, 4).map((tag) => (
        <span key={tag} className={styles.tag}>{tag}</span>
      ))}

      {/* Warranty hint */}
      {data.warrantyHint && (
        <span className={styles.warrantyText}>{data.warrantyHint}</span>
      )}
    </div>
  )
}
