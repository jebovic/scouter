import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Topnav, LoadingPulse, ScouterGrid, UsageWidget } from '../components/scouter'
import { MissionCard, MissionForm } from '../components/mission'
import { useMissions, useCreateMission } from '../hooks'
import { useNavigate } from 'react-router-dom'
import type { MissionCreateRequest } from '../types'

export default function HQDashboard() {
  const { t } = useTranslation()
  const { missions, isLoading } = useMissions()
  const { createMission, isPending } = useCreateMission()
  const navigate = useNavigate()
  const [showForm, setShowForm] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)

  async function handleCreate(req: MissionCreateRequest) {
    setCreateError(null)
    try {
      const mission = await createMission(req)
      setShowForm(false)
      navigate(`/missions/${mission.slug}`)
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create mission')
    }
  }

  return (
    <div className="page grid-bg scanlines">
      <Topnav />
      <main style={{ flex: 1, padding: '2rem' }}>
        <div className="container">
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: '2rem',
            }}
          >
            <div>
              <h1
                style={{
                  fontFamily: 'var(--font-display)',
                  color: 'var(--cyan)',
                  letterSpacing: '0.1em',
                  fontSize: '1.8rem',
                }}
              >
                {t('nav.missionControl')}
              </h1>
              <p style={{ color: 'var(--text-dim)', fontSize: '0.85rem', marginTop: 4 }}>
                {missions.length} active mission{missions.length !== 1 ? 's' : ''}
              </p>
            </div>
            <button
              onClick={() => setShowForm(true)}
              style={{
                padding: '10px 20px',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--cyan)',
                background: 'var(--cyan-dim)',
                color: 'var(--cyan)',
                cursor: 'pointer',
                fontFamily: 'var(--font-display)',
                fontWeight: 700,
                fontSize: '0.875rem',
                letterSpacing: '0.06em',
                transition: 'all 0.15s',
              }}
            >
              + {t('mission.create')}
            </button>
          </div>

          {showForm && (
            <div
              style={{
                position: 'fixed',
                inset: 0,
                background: 'rgba(5,8,16,0.85)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 200,
                padding: '1rem',
              }}
              onClick={(e) => e.target === e.currentTarget && setShowForm(false)}
            >
              <MissionForm
                onSubmit={handleCreate}
                onCancel={() => setShowForm(false)}
                loading={isPending}
                error={createError ?? undefined}
              />
            </div>
          )}

          <div style={{ marginBottom: '1.5rem' }}>
            <UsageWidget />
          </div>

          {isLoading ? (
            <LoadingPulse label="Loading missions..." />
          ) : missions.length === 0 ? (
            <div
              style={{
                textAlign: 'center',
                padding: '5rem 2rem',
                color: 'var(--text-dim)',
                animation: 'fade-in 0.5s ease both',
              }}
            >
              <div
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.65rem',
                  color: 'var(--border-glow)',
                  letterSpacing: '0.15em',
                  marginBottom: '1.5rem',
                  lineHeight: 1.6,
                }}
              >
                {'[ NO ACTIVE MISSIONS ]'}
              </div>
              <p style={{ fontSize: '0.95rem', marginBottom: '0.5rem', color: 'var(--text-mid)', fontFamily: 'var(--font-display)', letterSpacing: '0.08em' }}>
                STANDBY
              </p>
              <p style={{ fontSize: '0.8rem', fontFamily: 'var(--font-mono)', color: 'var(--text-dim)' }}>
                Create your first mission to begin scouting
              </p>
            </div>
          ) : (
            <ScouterGrid>
              {missions.map((mission) => (
                <MissionCard key={mission.id} mission={mission} />
              ))}
            </ScouterGrid>
          )}
        </div>
      </main>
    </div>
  )
}
