import { useState } from 'react'
import { useUsageSummary } from '../../hooks/useUsage'
import type { UsagePeriod } from '../../api/usage'
import styles from './UsageWidget.module.css'

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

  const fallbackCalls = summary?.fallback_calls ?? 0

  return (
    <div className={styles.widget}>
      {/* Header */}
      <div className={styles.header}>
        <span className={styles.title}>LLM USAGE</span>
        <div className={styles.periodToggle}>
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              className={period === p.value ? `${styles.periodBtn} ${styles.periodBtnActive}` : styles.periodBtn}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div className={styles.loading}>loading...</div>
      ) : error ? (
        <div className={styles.errorMsg}>usage unavailable</div>
      ) : (
        <div className={styles.statsList}>
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
          <div className={styles.divider}>
            <div className={styles.totals}>
              <span>
                fallback calls:{' '}
                <span
                  className={styles.fallbackCount}
                  style={{
                    '--fallback-color': fallbackCalls > 0 ? 'var(--coral)' : 'var(--text-dim)',
                  } as React.CSSProperties}
                >
                  {fallbackCalls}
                </span>
              </span>
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
    <div className={styles.statRow}>
      <span
        className={styles.statLabel}
        style={{ '--stat-color': color } as React.CSSProperties}
      >
        {label}
      </span>
      <span className={styles.statCalls}>{calls} calls</span>
      <span className={styles.statTokens}>{fmt(inputTokens + outputTokens)} tok</span>
      <span
        className={styles.statCost}
        style={{
          '--stat-cost-color': calls > 0 ? 'var(--text-mid)' : 'var(--text-dim)',
        } as React.CSSProperties}
      >
        {costLabel}
      </span>
    </div>
  )
}
