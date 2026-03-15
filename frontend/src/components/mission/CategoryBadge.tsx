import styles from './CategoryBadge.module.css'

interface CategoryBadgeProps {
  category?: string
  size?: 'sm' | 'md'
}

const CATEGORY_EMOJI: Record<string, string> = {
  'Électronique': '💡',
  'Informatique': '💻',
  'Audiovisuel': '📺',
  'Alimentation': '🍕',
  'Cuisine': '🍳',
  'Boissons': '🥤',
  'Voyage': '✈️',
  'Transport': '🚆',
  'Hébergement': '🏨',
  'Mode': '👗',
  'Vêtements': '👕',
  'Chaussures': '👟',
  'Maison': '🏠',
  'Mobilier': '🛋️',
  'Décoration': '🖼️',
  'Sport': '⚽',
  'Loisirs': '🎯',
  'Jeux': '🎮',
  'Santé': '💊',
  'Beauté': '💄',
  'Bien-être': '🧘',
  'Auto': '🚗',
  'Moto': '🏍️',
  'Vélo': '🚲',
  'Livres': '📚',
  'Culture': '🎭',
  'Musique': '🎵',
  'Services': '🔧',
  'Abonnements': '📋',
  'Autre': '🏷️',
}

export function CategoryBadge({ category, size = 'md' }: CategoryBadgeProps) {
  if (!category) return null

  const emoji = CATEGORY_EMOJI[category] ?? '🏷️'

  return (
    <span className={`${styles.badge} ${size === 'sm' ? styles.sm : styles.md}`}>
      <span className={styles.emoji} aria-hidden="true">{emoji}</span>
      <span className={styles.label}>{category}</span>
    </span>
  )
}
