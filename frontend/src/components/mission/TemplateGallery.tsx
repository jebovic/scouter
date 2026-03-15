import { useTranslation } from 'react-i18next'
import type { Template } from '../../types'
import { Skeleton } from '../scouter/Skeleton'
import { TemplateCard } from './TemplateCard'
import styles from './TemplateGallery.module.css'

interface TemplateGalleryProps {
  templates: Template[]
  onSelect: (t: Template) => void
  isLoading?: boolean
}

export function TemplateGallery({ templates, onSelect, isLoading }: TemplateGalleryProps) {
  const { t } = useTranslation()
  return (
    <section className={styles.section}>
      <h3 className={styles.heading}>{t('templates.galleryHeading', { defaultValue: 'Quick-Start Templates' })}</h3>
      {isLoading ? (
        <div className={styles.grid}>
          {[1, 2, 3, 4, 5, 6].map((n) => (
            <Skeleton key={n} variant="card" data-testid="skeleton" />
          ))}
        </div>
      ) : templates.length === 0 ? (
        <p className={styles.empty}>{t('templates.noTemplates', { defaultValue: 'No templates available' })}</p>
      ) : (
        <div className={styles.grid}>
          {templates.map((t) => (
            <TemplateCard key={t.slug} template={t} onSelect={onSelect} />
          ))}
        </div>
      )}
    </section>
  )
}
