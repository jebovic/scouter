import type { MissionCategory } from '../../types'
import styles from './CategoryTemplate.module.css'

interface CategoryTemplateProps {
  category: MissionCategory
}

const TEMPLATES: Record<MissionCategory, { icon: string; costCategories: string[]; tip: string }> = {
  travel: {
    icon: '✈️',
    costCategories: ['Flights', 'Accommodation', 'Transport', 'Food', 'Activities', 'Insurance'],
    tip: 'Book flights 6-8 weeks ahead for best prices.',
  },
  electronics: {
    icon: '📱',
    costCategories: ['Device', 'Accessories', 'Case & Protection', 'Warranty', 'Cables'],
    tip: 'Check for student/employee discounts and refurb options.',
  },
  computing: {
    icon: '💻',
    costCategories: ['Hardware', 'Peripherals', 'Software', 'Storage', 'Display', 'Warranty'],
    tip: 'Spec for 3–4 years ahead; avoid bottlenecking on RAM.',
  },
  renovation: {
    icon: '🏠',
    costCategories: ['Materials', 'Labour', 'Permits', 'Tools', 'Contingency'],
    tip: 'Add 20% contingency — surprises are guaranteed.',
  },
  custom: {
    icon: '🎯',
    costCategories: ['Primary', 'Secondary', 'Taxes & Fees', 'Misc'],
    tip: 'Define your own cost categories.',
  },
}

export function CategoryTemplate({ category }: CategoryTemplateProps) {
  const tmpl = TEMPLATES[category]

  return (
    <div className={styles.root}>
      <div className={styles.tagList}>
        {tmpl.costCategories.map((cat) => (
          <span key={cat} className={styles.tag}>
            {cat}
          </span>
        ))}
      </div>
      <p className={styles.tip}>💡 {tmpl.tip}</p>
    </div>
  )
}
