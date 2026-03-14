import styles from './EmptyState.module.css'

interface EmptyStateProps {
  icon: string
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
}

export function EmptyState({ icon, title, description, actionLabel, onAction }: EmptyStateProps) {
  return (
    <div className={styles.container}>
      <div className={styles.icon}>{icon}</div>
      <p className={styles.title}>{title}</p>
      <p className={styles.description}>{description}</p>
      {actionLabel && onAction && (
        <button className={styles.action} onClick={onAction}>
          {actionLabel}
        </button>
      )}
    </div>
  )
}
