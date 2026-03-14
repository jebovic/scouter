import { useLLMHealth } from '../../hooks/useLLMHealth'
import styles from './LLMStatus.module.css'

export function LLMStatus() {
  const { data, overallStatus } = useLLMHealth()

  const tooltipLines = data?.models.map(m =>
    `${m.name}: ${m.healthy ? '✓' : '✗'} (${m.circuit_state})`
  ) ?? ['LLM status loading…']

  return (
    <div
      className={styles.wrapper}
      role="status"
      aria-label={`LLM status: ${overallStatus}`}
      title={tooltipLines.join('\n')}
    >
      <span
        className={`${styles.dot} ${styles[overallStatus]}`}
        data-testid="llm-status-dot"
        data-status={overallStatus}
      />
    </div>
  )
}
