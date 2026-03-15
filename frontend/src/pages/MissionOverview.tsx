import { useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LoadingPulse, BudgetBar, StatusBadge } from '../components/scouter'
import { CategoryTemplate, DecisionPanel, MissionTimeline, PurchaseForm, LessonsField } from '../components/mission'
import { useMission, useShopping, useResearch, usePriceIntel, useUpdateMission, useKeyboardShortcuts, usePurchaseRecord, useDecision } from '../hooks'
import type { MissionPhase } from '../types'
import styles from './MissionOverview.module.css'

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
      <main className={styles.loading}>
        <LoadingPulse label="Loading mission..." />
      </main>
    )
  }

  if (!mission) {
    return (
      <main className={styles.notFound}>
        <p>Mission not found.</p>
      </main>
    )
  }

  const locale = mission.locale
  const currency = mission.currency
  const fmt = (v: number) => new Intl.NumberFormat(locale, { style: 'currency', currency }).format(v)

  return (
    <main className={styles.main}>
      <div className="container">
        {/* Header */}
        <div className={styles.header}>
          <span className={styles.headerIcon}>{mission.icon}</span>
          <div>
            <h1 className={styles.missionName}>{mission.name}</h1>
            <div className={styles.headerMeta}>
              <StatusBadge phase={mission.phase} />
              <span className={styles.categoryLabel}>{mission.category}</span>
            </div>
          </div>
        </div>

        {/* Main grid */}
        <div className={styles.cardGrid}>
          {/* Budget card */}
          <div className={styles.card}>
            <h3 className={styles.cardLabel}>{t('mission.budget')}</h3>
            <BudgetBar spent={spent} budget={mission.budget} currency={currency} />
            <div className={styles.budgetMeta}>
              <span>{fmt(spent)} spent</span>
              <span>{fmt(Math.max(0, mission.budget - spent))} remaining</span>
            </div>
          </div>

          {/* Phase switcher */}
          <div className={styles.card}>
            <h3 className={styles.cardLabel}>PHASE</h3>
            <div className={styles.phaseButtons}>
              {PHASES.map((p) => (
                <button
                  key={p}
                  onClick={() => handlePhase(p)}
                  disabled={updatePending}
                  className={`${styles.phaseBtn} ${mission.phase === p ? styles.phaseBtnActive : ''}`}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Agent action buttons */}
        <div className={styles.agentActions}>
          <button
            onClick={handleResearch}
            disabled={researchPending}
            className={`${styles.agentBtn} ${styles.researchBtn}`}
          >
            {researchPending ? '⚡ Researching...' : '⚡ Run Research Agent'}
          </button>
          <button
            onClick={handlePricing}
            disabled={pricingPending}
            className={`${styles.agentBtn} ${styles.pricingBtn}`}
          >
            {pricingPending ? '💰 Scouting prices...' : '💰 Run Price Intel'}
          </button>
        </div>

        {/* Category template */}
        <div className={styles.section}>
          <CategoryTemplate category={mission.category} />
        </div>

        {/* Constraints */}
        {mission.constraints.length > 0 && (
          <div className={`${styles.card} ${styles.section}`}>
            <h3 className={styles.cardLabel}>CONSTRAINTS</h3>
            <div className={styles.constraintList}>
              {mission.constraints.map((c) => (
                <div
                  key={c.key}
                  className={`${styles.constraint} ${c.type === 'hard' ? styles.constraintHard : styles.constraintSoft}`}
                >
                  {c.label}: {String(c.value)}
                  <span className={styles.constraintType}>{c.type}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Decision Engine */}
        <div className={styles.section}>
          <DecisionPanel mission={mission} />
        </div>

        {/* Mission Timeline */}
        <div className={`${styles.card} ${styles.section}`}>
          <h3 className={styles.cardLabel}>MISSION PROGRESS</h3>
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
          <div className={styles.section}>
            <PurchaseForm missionId={mission.id} existingRecord={purchaseRecord} />
          </div>
        )}

        {/* Lessons Learned — visible when done */}
        {mission.phase === 'done' && (
          <div className={styles.section}>
            <LessonsField
              missionSlug={mission.slug}
              value={mission.lessons}
              onSave={async (lessons) => { await updateMission({ lessons }) }}
            />
          </div>
        )}

        {/* Quick nav */}
        <div className={styles.quickNav}>
          <button
            onClick={() => navigate(`/missions/${slug}/options`)}
            className={styles.quickNavBtn}
          >
            <div className={styles.quickNavIcon}>🔍</div>
            <div className={`${styles.quickNavTitle} ${styles.quickNavTitleOptions}`}>Options Explorer</div>
            <div className={styles.quickNavDesc}>Compare research results</div>
          </button>
          <button
            onClick={() => navigate(`/missions/${slug}/shopping`)}
            className={styles.quickNavBtn}
          >
            <div className={styles.quickNavIcon}>🛒</div>
            <div className={`${styles.quickNavTitle} ${styles.quickNavTitleShopping}`}>Shopping Tracker</div>
            <div className={styles.quickNavDesc}>Track prices and merchants</div>
          </button>
        </div>
      </div>
    </main>
  )
}
