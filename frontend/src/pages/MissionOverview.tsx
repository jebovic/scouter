import { useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LoadingPulse, BudgetBar, StatusBadge } from '../components/scouter'
import { CategoryTemplate, DecisionPanel, MissionTimeline, PurchaseForm, LessonsField } from '../components/mission'
import { useMission, useShopping, useResearch, usePriceIntel, useUpdateMission, useKeyboardShortcuts, usePurchaseRecord, useDecision } from '../hooks'
import type { MissionPhase } from '../types'

const PHASES: MissionPhase[] = ['researching', 'comparing', 'buying', 'done']

export default function MissionOverview() {
  const { slug } = useParams<{ slug: string }>()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { mission, isLoading } = useMission(slug!)
  const { items } = useShopping(mission?.id ?? '')
  const { triggerResearch, isPending: researchPending } = useResearch(mission?.id ?? '')
  const { triggerPricing, isPending: pricingPending } = usePriceIntel(mission?.id ?? '')
  const { updateMission, isPending: updatePending } = useUpdateMission(slug!)
  const { data: purchaseRecord } = usePurchaseRecord(mission?.id)
  const { decision } = useDecision(mission?.id ?? '')

  const spent = items.reduce((sum, i) => sum + i.price, 0)

  async function handlePhase(phase: MissionPhase) {
    if (!mission || updatePending) return
    await updateMission({ phase })
  }

  async function handleResearch() {
    await triggerResearch(undefined)
    navigate(`/missions/${slug}/options`)
  }

  async function handlePricing() {
    await triggerPricing(undefined)
    navigate(`/missions/${slug}/shopping`)
  }

  const shortcuts = useMemo(() => ({
    r: () => { if (!researchPending) handleResearch() },
    p: () => { if (!pricingPending) handlePricing() },
  }), [researchPending, pricingPending]) // eslint-disable-line react-hooks/exhaustive-deps
  useKeyboardShortcuts(shortcuts)

  if (isLoading) {
    return (
      <main style={{ flex: 1, padding: '2rem', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        <LoadingPulse label="Loading mission..." />
      </main>
    )
  }

  if (!mission) {
    return (
      <main style={{ flex: 1, padding: '2rem' }}>
        <p style={{ color: 'var(--coral)', fontFamily: 'var(--font-display)' }}>Mission not found.</p>
      </main>
    )
  }

  return (
    <main style={{ flex: 1, padding: '2rem' }}>
        <div className="container">
          {/* Header */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '2rem' }}>
            <span style={{ fontSize: '2.5rem' }}>{mission.icon}</span>
            <div>
              <h1
                style={{
                  fontFamily: 'var(--font-display)',
                  color: 'var(--cyan)',
                  letterSpacing: '0.1em',
                  fontSize: '1.8rem',
                  marginBottom: 4,
                }}
              >
                {mission.name}
              </h1>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <StatusBadge phase={mission.phase} />
                <span style={{ color: 'var(--text-dim)', fontSize: '0.8rem', fontFamily: 'var(--font-mono)' }}>
                  {mission.category}
                </span>
              </div>
            </div>
          </div>

          {/* Main grid */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(340px, 1fr))', gap: '1.5rem', marginBottom: '1.5rem' }}>
            {/* Budget card */}
            <div
              style={{
                background: 'var(--surface)',
                borderRadius: 'var(--radius)',
                border: '1px solid var(--border)',
                padding: '1.5rem',
              }}
            >
              <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--text-mid)', fontSize: '0.8rem', letterSpacing: '0.1em', marginBottom: '1rem' }}>
                {t('mission.budget')}
              </h3>
              <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />
              <div style={{ marginTop: '1rem', display: 'flex', gap: '1rem', fontSize: '0.8rem', color: 'var(--text-dim)' }}>
                <span>
                  {new Intl.NumberFormat(mission.locale, { style: 'currency', currency: mission.currency }).format(spent)} spent
                </span>
                <span>
                  {new Intl.NumberFormat(mission.locale, { style: 'currency', currency: mission.currency }).format(Math.max(0, mission.budget - spent))} remaining
                </span>
              </div>
            </div>

            {/* Phase switcher */}
            <div
              style={{
                background: 'var(--surface)',
                borderRadius: 'var(--radius)',
                border: '1px solid var(--border)',
                padding: '1.5rem',
              }}
            >
              <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--text-mid)', fontSize: '0.8rem', letterSpacing: '0.1em', marginBottom: '1rem' }}>
                PHASE
              </h3>
              <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                {PHASES.map((p) => (
                  <button
                    key={p}
                    onClick={() => handlePhase(p)}
                    disabled={updatePending}
                    style={{
                      padding: '6px 14px',
                      borderRadius: 'var(--radius-sm)',
                      border: `1px solid ${mission.phase === p ? 'var(--cyan)' : 'var(--border)'}`,
                      background: mission.phase === p ? 'var(--cyan-dim)' : 'transparent',
                      color: mission.phase === p ? 'var(--cyan)' : 'var(--text-dim)',
                      cursor: updatePending ? 'not-allowed' : 'pointer',
                      fontFamily: 'var(--font-display)',
                      fontSize: '0.75rem',
                      letterSpacing: '0.08em',
                      textTransform: 'uppercase',
                      transition: 'all 0.15s',
                    }}
                  >
                    {p}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Agent action buttons */}
          <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
            <button
              onClick={handleResearch}
              disabled={researchPending}
              style={{
                padding: '12px 24px',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--gold)',
                background: researchPending ? 'var(--surface)' : 'rgba(255,198,0,0.08)',
                color: 'var(--gold)',
                cursor: researchPending ? 'not-allowed' : 'pointer',
                fontFamily: 'var(--font-display)',
                fontWeight: 700,
                fontSize: '0.875rem',
                letterSpacing: '0.06em',
                transition: 'all 0.15s',
                opacity: researchPending ? 0.6 : 1,
              }}
            >
              {researchPending ? '⚡ Researching...' : '⚡ Run Research Agent'}
            </button>
            <button
              onClick={handlePricing}
              disabled={pricingPending}
              style={{
                padding: '12px 24px',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--purple)',
                background: pricingPending ? 'var(--surface)' : 'rgba(139,92,246,0.08)',
                color: 'var(--purple)',
                cursor: pricingPending ? 'not-allowed' : 'pointer',
                fontFamily: 'var(--font-display)',
                fontWeight: 700,
                fontSize: '0.875rem',
                letterSpacing: '0.06em',
                transition: 'all 0.15s',
                opacity: pricingPending ? 0.6 : 1,
              }}
            >
              {pricingPending ? '💰 Scouting prices...' : '💰 Run Price Intel'}
            </button>
          </div>

          {/* Category template */}
          <div style={{ marginBottom: '1.5rem' }}>
            <CategoryTemplate category={mission.category} />
          </div>

          {/* Constraints */}
          {mission.constraints.length > 0 && (
            <div
              style={{
                background: 'var(--surface)',
                borderRadius: 'var(--radius)',
                border: '1px solid var(--border)',
                padding: '1.5rem',
                marginBottom: '1.5rem',
              }}
            >
              <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--text-mid)', fontSize: '0.8rem', letterSpacing: '0.1em', marginBottom: '1rem' }}>
                CONSTRAINTS
              </h3>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                {mission.constraints.map((c) => (
                  <div
                    key={c.key}
                    style={{
                      padding: '4px 12px',
                      borderRadius: 'var(--radius-sm)',
                      background: c.type === 'hard' ? 'rgba(255,90,90,0.1)' : 'var(--surface-raised)',
                      border: `1px solid ${c.type === 'hard' ? 'var(--coral)' : 'var(--border)'}`,
                      fontSize: '0.8rem',
                      color: c.type === 'hard' ? 'var(--coral)' : 'var(--text-mid)',
                      fontFamily: 'var(--font-mono)',
                    }}
                  >
                    {c.label}: {String(c.value)}
                    <span style={{ marginLeft: 6, opacity: 0.5, fontSize: '0.7rem' }}>{c.type}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Decision Engine */}
          <div style={{ marginBottom: '1.5rem' }}>
            <DecisionPanel mission={mission} />
          </div>

          {/* Mission Timeline */}
          <div
            style={{
              background: 'var(--surface)',
              borderRadius: 'var(--radius)',
              border: '1px solid var(--border)',
              padding: '1.5rem',
              marginBottom: '1.5rem',
            }}
          >
            <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--text-mid)', fontSize: '0.8rem', letterSpacing: '0.1em', marginBottom: '0.5rem' }}>
              MISSION PROGRESS
            </h3>
            <MissionTimeline
              createdAt={mission.createdAt}
              hasDecision={!!decision}
              decisionAt={decision?.createdAt}
              hasPurchase={!!purchaseRecord}
              purchasedAt={purchaseRecord?.purchasedAt}
              hasReview={!!purchaseRecord?.review}
            />
          </div>

          {/* Purchase section — visible in buying/done phases */}
          {(mission.phase === 'buying' || mission.phase === 'done') && (
            <div style={{ marginBottom: '1.5rem' }}>
              <PurchaseForm missionId={mission.id} existingRecord={purchaseRecord} />
            </div>
          )}

          {/* Lessons Learned — visible when done */}
          {mission.phase === 'done' && (
            <div style={{ marginBottom: '1.5rem' }}>
              <LessonsField
                missionSlug={mission.slug}
                value={mission.lessons}
                onSave={async (lessons) => { await updateMission({ lessons }) }}
              />
            </div>
          )}

          {/* Quick nav */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <button
              onClick={() => navigate(`/missions/${slug}/options`)}
              style={{
                padding: '1.5rem',
                borderRadius: 'var(--radius)',
                border: '1px solid var(--border)',
                background: 'var(--surface)',
                color: 'var(--text-mid)',
                cursor: 'pointer',
                fontFamily: 'var(--font-display)',
                fontSize: '0.9rem',
                letterSpacing: '0.08em',
                transition: 'all 0.15s',
                textAlign: 'left',
              }}
            >
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>🔍</div>
              <div style={{ color: 'var(--cyan)', fontWeight: 700, marginBottom: 4 }}>Options Explorer</div>
              <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>Compare research results</div>
            </button>
            <button
              onClick={() => navigate(`/missions/${slug}/shopping`)}
              style={{
                padding: '1.5rem',
                borderRadius: 'var(--radius)',
                border: '1px solid var(--border)',
                background: 'var(--surface)',
                color: 'var(--text-mid)',
                cursor: 'pointer',
                fontFamily: 'var(--font-display)',
                fontSize: '0.9rem',
                letterSpacing: '0.08em',
                transition: 'all 0.15s',
                textAlign: 'left',
              }}
            >
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>🛒</div>
              <div style={{ color: 'var(--gold)', fontWeight: 700, marginBottom: 4 }}>Shopping Tracker</div>
              <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>Track prices and merchants</div>
            </button>
          </div>
        </div>
      </main>
  )
}
