import styles from './SmartSortToggle.module.css'

interface SmartSortToggleProps {
  enabled: boolean
  onToggle: () => void
  isLoading?: boolean
}

export function SmartSortToggle({ enabled, onToggle, isLoading }: SmartSortToggleProps) {
  return (
    <button
      className={`${styles.toggle} ${enabled ? styles.active : ''}`}
      onClick={onToggle}
      disabled={isLoading}
      aria-pressed={enabled}
      title={enabled ? 'Désactiver le tri intelligent' : 'Activer le tri intelligent'}
    >
      {isLoading ? (
        <span className={styles.spinner} aria-hidden="true" />
      ) : (
        <span aria-hidden="true">{enabled ? '🧠' : '↕'}</span>
      )}
      <span>{enabled ? 'Tri intelligent' : 'Tri normal'}</span>
    </button>
  )
}
