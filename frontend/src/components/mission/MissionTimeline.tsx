import { useState } from 'react'
import { useTimeline } from '../../hooks/useTimeline'
import { LoadingPulse } from '../scouter'
import styles from './MissionTimeline.module.css'

interface MissionTimelineProps {
  missionId: string
  missionSlug: string
}

const EVENTS_PER_PAGE = 5

export function MissionTimeline({ missionId, missionSlug }: MissionTimelineProps) {
  const { events, isLoading } = useTimeline(missionId, missionSlug)
  const [expanded, setExpanded] = useState(false)

  if (isLoading) {
    return <LoadingPulse label="Chargement de l'historique..." />
  }

  if (events.length === 0) {
    return (
      <div className={styles.emptyState} role="status">
        Aucune activité enregistrée
      </div>
    )
  }

  const displayedEvents = expanded ? events : events.slice(0, EVENTS_PER_PAGE)
  const hasMore = events.length > EVENTS_PER_PAGE

  const dateFormatter = new Intl.DateTimeFormat('fr-FR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div className={styles.timeline} role="region" aria-label="Mission activity timeline">
      <div className={styles.events}>
        {displayedEvents.map((event, index) => (
          <div
            key={event.id}
            className={styles.event}
            data-testid={`timeline-event-${event.id}`}
          >
            <div className={styles.eventDot} style={{ '--event-color': getEventColor(event.type) } as React.CSSProperties}>
              {event.icon}
            </div>
            <div className={styles.eventContent}>
              <div className={styles.eventTitle}>{event.title}</div>
              <div className={styles.eventDate}>{dateFormatter.format(event.date)}</div>
              {event.description && (
                <div className={styles.eventDescription}>{event.description}</div>
              )}
            </div>
            {index < displayedEvents.length - 1 && <div className={styles.eventLine} />}
          </div>
        ))}
      </div>

      {hasMore && (
        <button
          className={styles.expandButton}
          onClick={() => setExpanded(!expanded)}
          aria-expanded={expanded}
        >
          {expanded ? `Voir moins` : `Voir tout (${events.length})`}
        </button>
      )}
    </div>
  )
}

function getEventColor(type: string): string {
  const colors: Record<string, string> = {
    mission_created: 'var(--accent, #00e5ff)',
    option_added: 'var(--gold, #ffd93d)',
    research_run: 'var(--purple, #a78bfa)',
    item_added: 'var(--cyan, #67e8f9)',
    item_purchased: 'var(--green, #00d68f)',
    phase_changed: 'var(--text-dim, #4a5570)',
  }
  return colors[type] || 'var(--text-mid)'
}
