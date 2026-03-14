import { useState } from 'react'
import { useUsageSummary } from '../../hooks/useUsage'
import type { UsagePeriod } from '../../api/usage'

const PERIODS: { value: UsagePeriod; label: string }[] = [
  { value: '24h', label: '24H' },
  { value: '7d', label: '7D' },
  { value: '30d', label: '30D' },
]

// Anthropic rough pricing: $3 / 1M input, $15 / 1M output (claude-sonnet-4-6)
function estimateCost(inputTokens: number, outputTokens: number): string {
  const cost = (inputTokens / 1_000_000) * 3 + (outputTokens / 1_000_000) * 15
  return cost < 0.01 ? '<$0.01' : `$${cost.toFixed(2)}`
}

function fmt(n: number): string {
  return n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n)
}

export function UsageWidget() {
  const [period, setPeriod] = useState<UsagePeriod>('7d')
  const { summary, isLoading, error } = useUsageSummary(period)

  const anthropicStats = summary?.by_provider.find((p) => p.provider === 'anthropic')
  const ollamaStats = summary?.by_provider.find((p) => p.provider === 'ollama')

  return (
    <div
      style={{
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--radius)',
        padding: '1.25rem 1.5rem',
        minWidth: 260,
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem' }}>
        <span
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '0.65rem',
            letterSpacing: '0.15em',
            color: 'var(--border-glow)',
          }}
        >
          LLM USAGE
        </span>
        <div style={{ display: 'flex', gap: 4 }}>
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              style={{
                padding: '2px 8px',
                fontSize: '0.65rem',
                fontFamily: 'var(--font-mono)',
                letterSpacing: '0.08em',
                border: `1px solid ${period === p.value ? 'var(--cyan)' : 'var(--border)'}`,
                background: period === p.value ? 'var(--cyan-dim)' : 'transparent',
                color: period === p.value ? 'var(--cyan)' : 'var(--text-dim)',
                borderRadius: 4,
                cursor: 'pointer',
              }}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div style={{ color: 'var(--text-dim)', fontSize: '0.8rem', fontFamily: 'var(--font-mono)' }}>
          loading...
        </div>
      ) : error ? (
        <div style={{ color: 'var(--coral)', fontSize: '0.8rem', fontFamily: 'var(--font-mono)' }}>
          usage unavailable
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem' }}>
          {/* Ollama row */}
          <StatRow
            label="Ollama"
            color="var(--purple)"
            calls={ollamaStats?.calls ?? 0}
            inputTokens={ollamaStats?.input_tokens ?? 0}
            outputTokens={ollamaStats?.output_tokens ?? 0}
            costLabel="local"
          />
          {/* Anthropic fallback row */}
          <StatRow
            label="Claude (fallback)"
            color="var(--coral)"
            calls={anthropicStats?.calls ?? 0}
            inputTokens={anthropicStats?.input_tokens ?? 0}
            outputTokens={anthropicStats?.output_tokens ?? 0}
            costLabel={
              anthropicStats
                ? estimateCost(anthropicStats.input_tokens, anthropicStats.output_tokens)
                : '$0.00'
            }
          />
          {/* Divider */}
          <div style={{ borderTop: '1px solid var(--border)', paddingTop: '0.5rem', marginTop: '0.1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', fontFamily: 'var(--font-mono)', color: 'var(--text-dim)' }}>
              <span>fallback calls: <span style={{ color: (summary?.fallback_calls ?? 0) > 0 ? 'var(--coral)' : 'var(--text-dim)' }}>{summary?.fallback_calls ?? 0}</span></span>
              <span>total tokens: {fmt((summary?.total_input_tokens ?? 0) + (summary?.total_output_tokens ?? 0))}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StatRow({
  label,
  color,
  calls,
  inputTokens,
  outputTokens,
  costLabel,
}: {
  label: string
  color: string
  calls: number
  inputTokens: number
  outputTokens: number
  costLabel: string
}) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        fontSize: '0.75rem',
        fontFamily: 'var(--font-mono)',
      }}
    >
      <span style={{ color, fontWeight: 600, minWidth: 110 }}>{label}</span>
      <span style={{ color: 'var(--text-mid)' }}>{calls} calls</span>
      <span style={{ color: 'var(--text-dim)' }}>{fmt(inputTokens + outputTokens)} tok</span>
      <span style={{ color: calls > 0 ? 'var(--text-mid)' : 'var(--text-dim)' }}>{costLabel}</span>
    </div>
  )
}
