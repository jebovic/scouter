import { useEffect, useId } from 'react'
import type { Template } from '../../types'
import styles from './TemplatePreview.module.css'

interface TemplatePreviewProps {
  template: Template
  onApply: (t: Template) => void
  onClose: () => void
}

export function TemplatePreview({ template, onApply, onClose }: TemplatePreviewProps) {
  const titleId = useId()

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [onClose])

  return (
    <div
      className={styles.overlay}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className={styles.panel}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
      >
        <div className={styles.header}>
          <span className={styles.icon}>{template.icon}</span>
          <div className={styles.titleGroup}>
            <h3 id={titleId} className={styles.name}>{template.name}</h3>
            <span className={styles.category}>{template.category}</span>
          </div>
        </div>

        <p className={styles.description}>{template.description}</p>

        {template.suggestedBudget && (
          <div className={styles.section}>
            <p className={styles.sectionLabel}>SUGGESTED BUDGET (USD)</p>
            <p className={styles.budget}>
              {template.suggestedBudget.min.toLocaleString()}–{template.suggestedBudget.max.toLocaleString()}
            </p>
          </div>
        )}

        <div className={styles.section}>
          <p className={styles.sectionLabel}>WEIGHT PROFILE</p>
          <div className={styles.weights}>
            <span>Price: {Math.round(template.weightProfile.price * 100)}%</span>
            <span>Quality: {Math.round(template.weightProfile.quality * 100)}%</span>
            <span>Feature: {Math.round(template.weightProfile.feature * 100)}%</span>
          </div>
        </div>

        {template.constraints.length > 0 && (
          <div className={styles.section}>
            <p className={styles.sectionLabel}>CONSTRAINTS</p>
            <ul className={styles.list}>
              {template.constraints.map((c) => {
                const badgeClass = c.type === 'hard' ? styles.hard : styles.soft
                return (
                  <li key={c.key} className={styles.constraint}>
                    <span className={`${styles.badge} ${badgeClass}`}>{c.type}</span>
                    {c.label}
                  </li>
                )
              })}
            </ul>
          </div>
        )}

        {template.costCategories.length > 0 && (
          <div className={styles.section}>
            <p className={styles.sectionLabel}>COST CATEGORIES</p>
            <div className={styles.tags}>
              {template.costCategories.map((cat) => (
                <span key={cat} className={styles.tag}>{cat}</span>
              ))}
            </div>
          </div>
        )}

        <div className={styles.actions}>
          {/* eslint-disable-next-line jsx-a11y/no-autofocus */}
          <button autoFocus className={styles.applyBtn} onClick={() => onApply(template)}>
            Apply Template
          </button>
          <button className={styles.cancelBtn} onClick={onClose}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  )
}
