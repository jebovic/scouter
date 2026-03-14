import type { Template } from '../../types'
import styles from './TemplateCard.module.css'

interface TemplateCardProps {
  template: Template
  onSelect: (t: Template) => void
}

export function TemplateCard({ template, onSelect }: TemplateCardProps) {
  return (
    <button className={styles.card} onClick={() => onSelect(template)}>
      <div className={styles.header}>
        <span className={styles.icon}>{template.icon}</span>
        <span className={styles.category}>{template.category}</span>
      </div>
      <h4 className={styles.name}>{template.name}</h4>
      <p className={styles.description}>{template.description}</p>
      {template.suggestedBudget && (
        <p className={styles.budget}>
          Budget: {template.suggestedBudget.min.toLocaleString()}–{template.suggestedBudget.max.toLocaleString()} USD
        </p>
      )}
    </button>
  )
}
