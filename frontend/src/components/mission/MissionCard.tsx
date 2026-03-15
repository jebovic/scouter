import { useNavigate } from 'react-router-dom'
import { Card, StatusBadge, BudgetBar } from '../scouter'
import { CategoryBadge } from './CategoryBadge'
import { useSwipeGesture } from '../../hooks/useSwipeGesture'
import { useDuplicateMission } from '../../hooks/useMission'
import type { Mission, ShoppingItem } from '../../types'
import styles from './MissionCard.module.css'

interface MissionCardProps {
  mission: Mission
  items?: ShoppingItem[]
  onArchive?: (missionId: string) => void
}

export function MissionCard({ mission, items = [], onArchive }: MissionCardProps) {
  const navigate = useNavigate()
  const { duplicateMission, isPending: isDuplicating } = useDuplicateMission()
  const spent = items.reduce((sum, item) => sum + item.price, 0)

  const { swipeX, handlers } = useSwipeGesture({
    threshold: 80,
    onSwipeLeft: () => onArchive?.(mission.id),
    onSwipeRight: () => navigate(`/missions/${mission.slug}?research=1`),
  })

  const isSwipingLeft = swipeX < -20
  const isSwipingRight = swipeX > 20

  return (
    <div
      className={styles.swipeWrapper}
      {...handlers}
    >
      {/* Archive action hint (revealed on swipe-left) */}
      <div
        className={`${styles.swipeAction} ${styles.swipeActionLeft} ${isSwipingLeft ? styles.swipeActionVisible : ''}`}
        aria-hidden="true"
      >
        <span className={styles.swipeActionIcon}>📦</span>
        <span className={styles.swipeActionLabel}>Archive</span>
      </div>

      {/* Research action hint (revealed on swipe-right) */}
      <div
        className={`${styles.swipeAction} ${styles.swipeActionRight} ${isSwipingRight ? styles.swipeActionVisible : ''}`}
        aria-hidden="true"
      >
        <span className={styles.swipeActionIcon}>⚡</span>
        <span className={styles.swipeActionLabel}>Research</span>
      </div>

      <div
        className={styles.swipeCard}
        style={{ '--swipe-x': `${swipeX}px` } as React.CSSProperties}
      >
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
                {mission.autotagCategory && (
                  <div className={styles.autotagBadge}>
                    <CategoryBadge category={mission.autotagCategory} size="sm" />
                  </div>
                )}
              </div>
            </div>
            <div className={styles.headerActions}>
              <StatusBadge phase={mission.phase} />
              <button
                className={styles.duplicateBtn}
                title="Duplicate mission"
                aria-label="Duplicate mission"
                disabled={isDuplicating}
                onClick={async (e) => {
                  e.stopPropagation()
                  const copy = await duplicateMission(mission.slug)
                  navigate(`/missions/${copy.slug}`)
                }}
              >
                {isDuplicating ? '…' : '📋'}
              </button>
            </div>
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
      </div>
    </div>
  )
}
