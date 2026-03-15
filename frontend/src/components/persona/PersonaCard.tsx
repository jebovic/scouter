import { useTranslation } from 'react-i18next'
import { usePersona, useRunPersona } from '../../hooks/usePersona'
import styles from './PersonaCard.module.css'

export function PersonaCard() {
  const { t } = useTranslation()
  const { persona, isLoading } = usePersona()
  const { runPersona, isPending } = useRunPersona()

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.header}>
          <h3 className={styles.sectionTitle}>{t('persona.title')}</h3>
        </div>
        <div className={styles.skeleton}>
          <div className={styles.skeletonArchetype} />
          <div className={styles.skeletonLine} />
          <div className={styles.skeletonLine} style={{ width: '80%' } as React.CSSProperties} />
          <div className={styles.skeletonBadgesRow}>
            <div className={styles.skeletonBadge} />
            <div className={styles.skeletonBadge} />
            <div className={styles.skeletonBadge} />
          </div>
        </div>
      </div>
    )
  }

  if (!persona) {
    return (
      <div className={styles.card}>
        <div className={styles.header}>
          <h3 className={styles.sectionTitle}>{t('persona.title')}</h3>
          <button
            className={styles.runBtn}
            onClick={() => runPersona()}
            disabled={isPending}
          >
            {isPending ? t('persona.analyzing') : t('persona.runAnalysis')}
          </button>
        </div>
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>◈</div>
          <div className={styles.emptyTitle}>{t('persona.noPersona')}</div>
          <div className={styles.emptyDesc}>{t('persona.noPersonaDesc')}</div>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <h3 className={styles.sectionTitle}>{t('persona.title')}</h3>
        <button
          className={styles.runBtn}
          onClick={() => runPersona()}
          disabled={isPending}
        >
          {isPending ? t('persona.analyzing') : t('persona.runAnalysis')}
        </button>
      </div>

      <div className={styles.body}>
        <div className={styles.archetype}>{persona.archetype}</div>

        <p className={styles.summary}>{persona.summary}</p>

        {persona.traits.length > 0 && (
          <div className={styles.section}>
            <div className={styles.sectionLabel}>{t('persona.traits').toUpperCase()}</div>
            <div className={styles.traitsRow}>
              {persona.traits.map((trait, i) => (
                <span key={i} className={styles.traitBadge}>{trait}</span>
              ))}
            </div>
          </div>
        )}

        {persona.tips.length > 0 && (
          <div className={styles.section}>
            <div className={styles.sectionLabel}>{t('persona.tips').toUpperCase()}</div>
            <ul className={styles.tipsList}>
              {persona.tips.map((tip, i) => (
                <li key={i} className={styles.tipItem}>
                  <span className={styles.tipBullet}>›</span>
                  <span>{tip}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        <div className={styles.meta}>
          <span className={styles.metaLabel}>{t('persona.lastAnalyzed')}</span>
          <span className={styles.metaValue}>
            {new Date(persona.createdAt).toLocaleDateString()}
          </span>
        </div>
      </div>
    </div>
  )
}
