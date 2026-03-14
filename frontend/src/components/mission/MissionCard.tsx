import { useNavigate } from 'react-router-dom'
import { Card, StatusBadge, BudgetBar } from '../scouter'
import type { Mission, ShoppingItem } from '../../types'

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
      className="card-enter"
      style={{ display: 'flex', flexDirection: 'column', gap: 12 }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 8 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span
            style={{
              fontSize: '1.8rem',
              lineHeight: 1,
              filter: 'drop-shadow(0 0 8px rgba(0,229,255,0.3))',
            }}
          >
            {mission.icon}
          </span>
          <div>
            <div
              style={{
                fontFamily: 'var(--font-display)',
                fontWeight: 700,
                fontSize: '1rem',
                color: 'var(--text)',
                lineHeight: 1.2,
                letterSpacing: '0.04em',
              }}
            >
              {mission.name}
            </div>
            <div
              style={{
                fontSize: '0.7rem',
                color: 'var(--text-dim)',
                marginTop: 3,
                fontFamily: 'var(--font-mono)',
                letterSpacing: '0.06em',
                textTransform: 'uppercase',
              }}
            >
              {mission.category}
            </div>
          </div>
        </div>
        <StatusBadge phase={mission.phase} />
      </div>

      <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />

      {mission.constraints.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {mission.constraints.slice(0, 3).map((c) => (
            <span
              key={c.key}
              style={{
                padding: '2px 8px',
                borderRadius: 4,
                fontSize: '0.7rem',
                fontFamily: 'var(--font-mono)',
                background: c.type === 'hard' ? 'var(--coral-dim)' : 'var(--raised)',
                color: c.type === 'hard' ? 'var(--coral)' : 'var(--text-mid)',
                border: `1px solid ${c.type === 'hard' ? 'rgba(255,61,113,0.2)' : 'var(--border)'}`,
              }}
            >
              {c.label}
            </span>
          ))}
          {mission.constraints.length > 3 && (
            <span style={{ fontSize: '0.7rem', color: 'var(--text-dim)', padding: '2px 4px', fontFamily: 'var(--font-mono)' }}>
              +{mission.constraints.length - 3}
            </span>
          )}
        </div>
      )}
    </Card>
  )
}
