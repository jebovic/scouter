import type { Option } from '../../types'

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

  const cellStyle: React.CSSProperties = {
    padding: '8px 12px',
    borderBottom: '1px solid var(--border)',
    fontSize: '0.8rem',
    verticalAlign: 'middle',
  }
  const headerStyle: React.CSSProperties = {
    ...cellStyle,
    color: 'var(--text-dim)',
    fontFamily: 'var(--font-mono)',
    fontSize: '0.7rem',
    letterSpacing: '0.05em',
    background: 'var(--raised)',
    whiteSpace: 'nowrap',
  }

  return (
    <div style={{ overflowX: 'auto' }}>
      <table
        style={{
          width: '100%',
          borderCollapse: 'collapse',
          background: 'var(--surface)',
          borderRadius: 'var(--radius)',
          overflow: 'hidden',
          border: '1px solid var(--border)',
        }}
      >
        <thead>
          <tr>
            <th style={{ ...headerStyle, textAlign: 'left' }}>ATTRIBUTE</th>
            {options.map((o) => (
              <th
                key={o.id}
                style={{
                  ...headerStyle,
                  textAlign: 'center',
                  color: o.badge === 'recommended' ? 'var(--cyan)' : 'var(--text-dim)',
                }}
              >
                {o.name}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {/* Price row */}
          <tr>
            <td style={{ ...cellStyle, color: 'var(--text-dim)' }}>Best Price</td>
            {options.map((o) => (
              <td key={o.id} style={{ ...cellStyle, textAlign: 'center', fontFamily: 'var(--font-mono)', color: 'var(--cyan)' }}>
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
              <td style={{ ...cellStyle, color: 'var(--text-dim)' }}>{attrLabels[key] ?? key}</td>
              {options.map((o) => {
                const attr = o.attributes.find((a) => a.key === key)
                return (
                  <td key={o.id} style={{ ...cellStyle, textAlign: 'center' }}>
                    {attr ? (
                      <span
                        style={{
                          color:
                            attr.pass === true
                              ? 'var(--green)'
                              : attr.pass === false
                              ? 'var(--coral)'
                              : 'var(--text-mid)',
                        }}
                      >
                        {attr.type === 'boolean'
                          ? attr.value
                            ? '✓'
                            : '✗'
                          : String(attr.value)}
                      </span>
                    ) : (
                      <span style={{ color: 'var(--text-dim)' }}>—</span>
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
