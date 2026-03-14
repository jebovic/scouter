import styles from './LoadingPulse.module.css'

interface LoadingPulseProps {
  label?: string
  color?: string
}

export function LoadingPulse({ label = 'Loading...', color = 'var(--cyan)' }: LoadingPulseProps) {
  return (
    <div className={styles.wrapper}>
      {/* Radar scan bar */}
      <div className={styles.track}>
        <div
          className={styles.scanner}
          style={{ '--scan-color': color } as React.CSSProperties}
        />
      </div>
      <span className={styles.label}>
        {label}
      </span>
    </div>
  )
}
