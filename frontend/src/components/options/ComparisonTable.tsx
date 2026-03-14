import type { Option } from '../../types'
import styles from './ComparisonTable.module.css'

interface ComparisonTableProps {
  options: Option[]
  currency?: string
}

export function ComparisonTable({ options, currency = 'USD' }: ComparisonTableProps) {
  if (options.length === 0) return null

  // Collect all unique attribute keys
  const attrKeys = Array.from(
    new Set(options.flatMap((o) => o.attributes.map((a) => a.key))),
  )
  const attrLabels = Object.fromEntries(
    options
      .flatMap((o) => o.attributes)
      .map((a) => [a.key, a.label]),
  )

  return (
    <div className={styles.wrapper}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th className={`${styles.headerCell} ${styles.headerLeft}`}>ATTRIBUTE</th>
            {options.map((o) => (
              <th
                key={o.id}
                className={`${styles.headerCell} ${styles.headerCenter} ${o.badge === 'recommended' ? styles.headerRecommended : ''}`}
              >
                {o.name}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {/* Price row */}
          <tr>
            <td className={`${styles.cell} ${styles.cellLabel}`}>Best Price</td>
            {options.map((o) => (
              <td key={o.id} className={`${styles.cell} ${styles.cellPrice}`}>
                {o.priceRange
                  ? new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(
                      o.priceRange.best,
                    )
                  : '—'}
              </td>
            ))}
          </tr>
          {/* Attribute rows */}
          {attrKeys.map((key) => (
            <tr key={key}>
              <td className={`${styles.cell} ${styles.cellLabel}`}>{attrLabels[key] ?? key}</td>
              {options.map((o) => {
                const attr = o.attributes.find((a) => a.key === key)
                return (
                  <td key={o.id} className={`${styles.cell} ${styles.cellCenter}`}>
                    {attr ? (
                      <span
                        className={
                          attr.pass === true
                            ? styles.attrPass
                            : attr.pass === false
                            ? styles.attrFail
                            : styles.attrNeutral
                        }
                      >
                        {attr.type === 'boolean'
                          ? attr.value
                            ? '✓'
                            : '✗'
                          : String(attr.value)}
                      </span>
                    ) : (
                      <span className={styles.attrMissing}>—</span>
                    )}
                  </td>
                )
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
