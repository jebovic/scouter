import type { Attribute } from '../../types'

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
          <span style={{ color: value ? 'var(--green)' : 'var(--coral)', fontSize: '0.85rem' }}>
            {value ? '✓' : '✗'}
          </span>
        )
      case 'price':
        return (
          <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--cyan)', fontSize: '0.85rem' }}>
            {new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(
              Number(value),
            )}
          </span>
        )
      case 'score':
        return (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <div
              style={{
                height: 4,
                width: 60,
                background: 'var(--raised)',
                borderRadius: 2,
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  height: '100%',
                  width: `${Math.min((Number(value) / (max ?? 10)) * 100, 100)}%`,
                  background:
                    pass === true
                      ? 'var(--green)'
                      : pass === false
                      ? 'var(--coral)'
                      : 'var(--cyan)',
                  borderRadius: 2,
                }}
              />
            </div>
            <span
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '0.75rem',
                color: pass === true ? 'var(--green)' : pass === false ? 'var(--coral)' : 'var(--text-mid)',
              }}
            >
              {String(value)}{max ? `/${max}` : ''}
            </span>
          </div>
        )
      default:
        return (
          <span style={{ fontSize: '0.8rem', color: 'var(--text-mid)' }}>
            {String(value)}
          </span>
        )
    }
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
      <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', flexShrink: 0 }}>
        {attribute.label}
      </span>
      {renderValue()}
    </div>
  )
}
