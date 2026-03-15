import { Link } from 'react-router-dom'
import type { Mission } from '../../types'
import styles from './KanbanBoard.module.css'

interface KanbanBoardProps {
  missions: Mission[]
}

type MissionStatus = 'active' | 'pending' | 'done' | 'archived'

function getMissionStatus(mission: Mission): MissionStatus {
  if (mission.archivedAt) {
    return 'archived'
  }

  const now = Date.now()
  const deadline = (mission.timeline ?? []).find((e) => e.label.toLowerCase().includes('deadline'))
  const hasDeadline = deadline && new Date(deadline.date).getTime() > now

  // Heuristic: mission is done if phase is 'done'
  if (mission.phase === 'done') {
    return 'done'
  }

  // No deadline = pending
  if (!hasDeadline) {
    return 'pending'
  }

  // Otherwise active
  return 'active'
}

const COLUMN_CONFIG: Record<MissionStatus, { label: string; icon: string; color: string }> = {
  active: {
    label: 'Actif',
    icon: '⚡',
    color: 'var(--cyan)',
  },
  pending: {
    label: 'En attente',
    icon: '⏸',
    color: 'var(--gold)',
  },
  done: {
    label: 'Terminé',
    icon: '✓',
    color: 'var(--green)',
  },
  archived: {
    label: 'Archivé',
    icon: '📦',
    color: 'var(--text-dim)',
  },
}

function KanbanColumn({
  status,
  missions: columnMissions,
}: {
  status: MissionStatus
  missions: Mission[]
}) {
  const config = COLUMN_CONFIG[status]

  return (
    <div className={styles.column}>
      <div className={styles.columnHeader} style={{ '--accent-color': config.color } as React.CSSProperties}>
        <div className={styles.columnTitle}>
          <span className={styles.columnIcon}>{config.icon}</span>
          <span className={styles.columnLabel}>{config.label}</span>
          <span className={styles.columnCount}>{columnMissions.length}</span>
        </div>
      </div>

      <div className={styles.columnBody}>
        {columnMissions.length === 0 ? (
          <div className={styles.emptyColumn}>
            <div className={styles.emptyIcon}>○</div>
            <div className={styles.emptyText}>Aucune mission</div>
          </div>
        ) : (
          <div className={styles.cardList}>
            {columnMissions.map((mission) => (
              <Link
                key={mission.id}
                to={`/missions/${mission.slug}`}
                className={styles.missionCard}
              >
                <div className={styles.cardHeader}>
                  <span className={styles.cardIcon}>{mission.icon}</span>
                  <span className={styles.cardName}>{mission.name}</span>
                </div>

                <div className={styles.cardMeta}>
                  <span className={styles.budget}>
                    {mission.budget} {mission.currency}
                  </span>
                </div>

                {(mission.timeline ?? []).length > 0 && (
                  <div className={styles.deadline}>
                    📅 {new Date(mission.timeline[0].date).toLocaleDateString('fr-FR', {
                      day: 'numeric',
                      month: 'short',
                    })}
                  </div>
                )}

                <div className={styles.constraintCount}>
                  {(mission.constraints ?? []).length} contrainte{(mission.constraints ?? []).length !== 1 ? 's' : ''}
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export function KanbanBoard({ missions }: KanbanBoardProps) {
  const grouped = {
    active: missions.filter((m) => getMissionStatus(m) === 'active'),
    pending: missions.filter((m) => getMissionStatus(m) === 'pending'),
    done: missions.filter((m) => getMissionStatus(m) === 'done'),
    archived: missions.filter((m) => getMissionStatus(m) === 'archived'),
  }

  return (
    <div className={styles.board}>
      {(['active', 'pending', 'done', 'archived'] as const).map((status) => (
        <KanbanColumn key={status} status={status} missions={grouped[status]} />
      ))}
    </div>
  )
}
