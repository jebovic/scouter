import type { MissionCategory } from '../../types'

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
    <div
      style={{
        padding: '12px 16px',
        borderRadius: 'var(--radius-sm)',
        background: 'var(--raised)',
        border: '1px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
      }}
    >
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {tmpl.costCategories.map((cat) => (
          <span
            key={cat}
            style={{
              padding: '2px 8px',
              borderRadius: 4,
              fontSize: '0.75rem',
              background: 'var(--surface)',
              color: 'var(--text-mid)',
              border: '1px solid var(--border)',
            }}
          >
            {cat}
          </span>
        ))}
      </div>
      <p style={{ fontSize: '0.75rem', color: 'var(--text-dim)', margin: 0, fontStyle: 'italic' }}>
        💡 {tmpl.tip}
      </p>
    </div>
  )
}
