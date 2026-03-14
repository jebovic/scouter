import type { Attribute } from '../../types'
import styles from './AttributeRenderer.module.css'

interface AttributeRendererProps {
  attribute: Attribute
  currency?: string
}

export function AttributeRenderer({ attribute, currency = 'USD' }: AttributeRendererProps) {
  function renderValue() {
    const { type, value, max, pass } = attribute

    switch (type) {
      case 'boolean':
        return (
          <span className={value ? styles.boolTrue : styles.boolFalse}>
            {value ? '✓' : '✗'}
          </span>
        )
      case 'price':
        return (
          <span className={styles.price}>
            {new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(
              Number(value),
            )}
          </span>
        )
      case 'score': {
        const scoreColor =
          pass === true ? 'var(--green)' : pass === false ? 'var(--coral)' : 'var(--cyan)'
        const scoreTextColor =
          pass === true ? 'var(--green)' : pass === false ? 'var(--coral)' : 'var(--text-mid)'
        const scoreWidth = `${Math.min((Number(value) / (max ?? 10)) * 100, 100)}%`
        return (
          <div className={styles.scoreWrapper}>
            <div className={styles.scoreTrack}>
              <div
                className={styles.scoreFill}
                style={
                  {
                    '--score-width': scoreWidth,
                    '--score-color': scoreColor,
                  } as React.CSSProperties
                }
              />
            </div>
            <span
              className={styles.scoreText}
              style={{ '--score-text-color': scoreTextColor } as React.CSSProperties}
            >
              {String(value)}{max ? `/${max}` : ''}
            </span>
          </div>
        )
      }
      default:
        return (
          <span className={styles.defaultValue}>
            {String(value)}
          </span>
        )
    }
  }

  return (
    <div className={styles.row}>
      <span className={styles.label}>
        {attribute.label}
      </span>
      {renderValue()}
    </div>
  )
}
