import { useTranslation } from 'react-i18next'
import type { Template } from '../../types'
import styles from './TemplateCard.module.css'

interface TemplateCardProps {
  template: Template
  onSelect: (t: Template) => void
}

export function TemplateCard({ template, onSelect }: TemplateCardProps) {
  const { t } = useTranslation()
  return (
    <button className={styles.card} onClick={() => onSelect(template)}>
      <div className={styles.header}>
        <span className={styles.icon}>{template.icon}</span>
        <span className={styles.category}>{template.category}</span>
      </div>
      <h4 className={styles.name}>{t('templates.' + template.slug + '.name', { defaultValue: template.name })}</h4>
      <p className={styles.description}>{t('templates.' + template.slug + '.description', { defaultValue: template.description })}</p>
      {template.suggestedBudget && (
        <p className={styles.budget}>
          Budget: {template.suggestedBudget.min.toLocaleString()}–{template.suggestedBudget.max.toLocaleString()} USD
        </p>
      )}
    </button>
  )
}
