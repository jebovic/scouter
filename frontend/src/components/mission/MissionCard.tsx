import { useNavigate } from 'react-router-dom'
import { Card, StatusBadge, BudgetBar } from '../scouter'
import type { Mission, ShoppingItem } from '../../types'
import styles from './MissionCard.module.css'

interface MissionCardProps {
  mission: Mission
  items?: ShoppingItem[]
}

export function MissionCard({ mission, items = [] }: MissionCardProps) {
  const navigate = useNavigate()
  const spent = items.reduce((sum, item) => sum + item.price, 0)

  return (
    <Card
      onClick={() => navigate(`/missions/${mission.slug}`)}
      className={`card-enter ${styles.card}`}
    >
      <div className={styles.header}>
        <div className={styles.identity}>
          <span className={styles.icon}>{mission.icon}</span>
          <div>
            <div className={styles.name}>{mission.name}</div>
            <div className={styles.category}>{mission.category}</div>
          </div>
        </div>
        <StatusBadge phase={mission.phase} />
      </div>

      <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />

      {mission.constraints.length > 0 && (
        <div className={styles.constraintList}>
          {mission.constraints.slice(0, 3).map((c) => (
            <span
              key={c.key}
              className={`${styles.constraintTag} ${c.type === 'hard' ? styles.constraintHard : styles.constraintSoft}`}
            >
              {c.label}
            </span>
          ))}
          {mission.constraints.length > 3 && (
            <span className={styles.constraintOverflow}>
              +{mission.constraints.length - 3}
            </span>
          )}
        </div>
      )}
    </Card>
  )
}
