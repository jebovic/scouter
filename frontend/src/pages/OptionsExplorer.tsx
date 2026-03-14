import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topnav, LoadingPulse } from '../components/scouter'
import { OptionCard } from '../components/options'
import { ComparisonTable } from '../components/options'
import { RadarChart } from '../components/options'
import { ConstraintChecker } from '../components/options'
import { useMission, useOptions, useResearch, useDecision } from '../hooks'
import type { OptionBadge } from '../types'

type ViewMode = 'grid' | 'compare'

const BADGE_FILTERS: { label: string; value: OptionBadge | 'all' }[] = [
  { label: 'All', value: 'all' },
  { label: 'Recommended', value: 'recommended' },
  { label: 'Alternative', value: 'alternative' },
  { label: 'Watch', value: 'watch' },
  { label: 'Rejected', value: 'rejected' },
]

export default function OptionsExplorer() {
  const { slug } = useParams<{ slug: string }>()
  const { t } = useTranslation()
  const { mission, isLoading: missionLoading } = useMission(slug!)
  const { options, isLoading: optionsLoading } = useOptions(mission?.id ?? '')
  const { triggerResearch, isPending: researchPending } = useResearch(mission?.id ?? '')
  const { decision } = useDecision(mission?.id ?? '')
  const [viewMode, setViewMode] = useState<ViewMode>('grid')
  const [badgeFilter, setBadgeFilter] = useState<OptionBadge | 'all'>('all')

  const isLoading = missionLoading || optionsLoading

  const filtered = badgeFilter === 'all' ? options : options.filter((o) => o.badge === badgeFilter)

  if (isLoading) {
    return (
      <div className="page grid-bg scanlines">
        <Topnav missionSlug={slug} />
        <main style={{ flex: 1, padding: '2rem', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <LoadingPulse label="Loading options..." />
        </main>
      </div>
    )
  }

  return (
    <div className="page grid-bg scanlines">
      <Topnav missionSlug={slug} missionName={mission?.name} />
      <main style={{ flex: 1, padding: '2rem' }}>
        <div className="container">
          {/* Header */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: '2rem',
              flexWrap: 'wrap',
              gap: '1rem',
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
                {t('nav.options')}
              </h1>
              <p style={{ color: 'var(--text-dim)', fontSize: '0.85rem', marginTop: 4 }}>
                {filtered.length} option{filtered.length !== 1 ? 's' : ''}
                {badgeFilter !== 'all' ? ` (${badgeFilter})` : ''}
              </p>
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center' }}>
              {/* View mode toggle */}
              <div
                style={{
                  display: 'flex',
                  border: '1px solid var(--border)',
                  borderRadius: 'var(--radius-sm)',
                  overflow: 'hidden',
                }}
              >
                {(['grid', 'compare'] as ViewMode[]).map((mode) => (
                  <button
                    key={mode}
                    onClick={() => setViewMode(mode)}
                    style={{
                      padding: '6px 14px',
                      border: 'none',
                      background: viewMode === mode ? 'var(--cyan-dim)' : 'transparent',
                      color: viewMode === mode ? 'var(--cyan)' : 'var(--text-dim)',
                      cursor: 'pointer',
                      fontFamily: 'var(--font-display)',
                      fontSize: '0.75rem',
                      letterSpacing: '0.08em',
                      textTransform: 'uppercase',
                    }}
                  >
                    {mode === 'grid' ? '⊞ Grid' : '⊟ Compare'}
                  </button>
                ))}
              </div>

              {/* Research button */}
              <button
                onClick={() => triggerResearch()}
                disabled={researchPending}
                style={{
                  padding: '8px 18px',
                  borderRadius: 'var(--radius-sm)',
                  border: '1px solid var(--gold)',
                  background: 'rgba(255,198,0,0.08)',
                  color: 'var(--gold)',
                  cursor: researchPending ? 'not-allowed' : 'pointer',
                  fontFamily: 'var(--font-display)',
                  fontSize: '0.8rem',
                  letterSpacing: '0.06em',
                  opacity: researchPending ? 0.6 : 1,
                }}
              >
                {researchPending ? '⚡ Running...' : '⚡ Re-run Research'}
              </button>
            </div>
          </div>

          {/* Badge filters */}
          <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
            {BADGE_FILTERS.map((f) => (
              <button
                key={f.value}
                onClick={() => setBadgeFilter(f.value)}
                style={{
                  padding: '4px 12px',
                  borderRadius: 'var(--radius-sm)',
                  border: `1px solid ${badgeFilter === f.value ? 'var(--cyan)' : 'var(--border)'}`,
                  background: badgeFilter === f.value ? 'var(--cyan-dim)' : 'transparent',
                  color: badgeFilter === f.value ? 'var(--cyan)' : 'var(--text-dim)',
                  cursor: 'pointer',
                  fontFamily: 'var(--font-display)',
                  fontSize: '0.75rem',
                  letterSpacing: '0.06em',
                }}
              >
                {f.label}
              </button>
            ))}
          </div>

          {/* Empty state */}
          {options.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '5rem 2rem', color: 'var(--text-dim)', animation: 'fade-in 0.5s ease both' }}>
              <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.65rem', color: 'var(--border-glow)', letterSpacing: '0.15em', marginBottom: '1.5rem' }}>
                {'[ AWAITING INTELLIGENCE ]'}
              </div>
              <p style={{ fontSize: '0.95rem', marginBottom: '0.5rem', color: 'var(--text-mid)', fontFamily: 'var(--font-display)', letterSpacing: '0.08em' }}>
                NO CONTACTS
              </p>
              <p style={{ fontSize: '0.8rem', fontFamily: 'var(--font-mono)', color: 'var(--text-dim)' }}>
                Run the Research Agent to discover options
              </p>
            </div>
          ) : viewMode === 'compare' ? (
            <div>
              <ComparisonTable options={filtered} />
              {mission && (
                <div style={{ marginTop: '2rem' }}>
                  {filtered.map((o) => (
                <div key={o.id} style={{ marginBottom: '0.5rem' }}>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)', marginBottom: 4, fontFamily: 'var(--font-mono)' }}>{o.name}</div>
                  <ConstraintChecker option={o} constraints={mission.constraints} />
                </div>
              ))}
                </div>
              )}
              <div style={{ marginTop: '2rem' }}>
                <RadarChart options={filtered} />
              </div>
            </div>
          ) : (
            <div>
              {/* Radar chart above grid when ≥3 options */}
              {filtered.length >= 3 && (
                <div style={{ marginBottom: '2rem' }}>
                  <RadarChart options={filtered} />
                </div>
              )}
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
                  gap: '1.25rem',
                }}
              >
                {filtered.map((option) => (
                  <OptionCard
                    key={option.id}
                    option={option}
                    score={decision?.scores.find((s) => s.optionId === option.id)?.score}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
