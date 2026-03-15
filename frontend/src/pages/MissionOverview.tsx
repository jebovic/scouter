import { useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LoadingPulse, BudgetBar, StatusBadge } from '../components/scouter'
import { CategoryTemplate, DecisionPanel, MissionTimeline, PurchaseForm, LessonsField, CollaboratorsPanel, TravelSearchWidget, TimingAdvisorCard, ExportPanel } from '../components/mission'
import { ForecastPanel } from '../components/forecast'
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
        <LoadingPulse label={t('mission.loadingMission')} />
      </main>
    )
  }

  if (!mission) {
    return (
      <main className={styles.notFound}>
        <p>{t('mission.notFound')}</p>
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
              <span>{fmt(spent)} {t('mission.spent')}</span>
              <span>{fmt(Math.max(0, mission.budget - spent))} {t('mission.remaining')}</span>
            </div>
          </div>

          {/* Phase switcher */}
          <div className={styles.card}>
            <h3 className={styles.cardLabel}>{t('mission.phase').toUpperCase()}</h3>
            <div className={styles.phaseButtons}>
              {PHASES.map((p) => (
                <button
                  key={p}
                  onClick={() => handlePhase(p)}
                  disabled={updatePending}
                  className={`${styles.phaseBtn} ${mission.phase === p ? styles.phaseBtnActive : ''}`}
                >
                  {t(`mission.phases.${p}`)}
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
            {researchPending ? `⚡ ${t('mission.researching')}` : `⚡ ${t('mission.runResearchAgent')}`}
          </button>
          <button
            onClick={handlePricing}
            disabled={pricingPending}
            className={`${styles.agentBtn} ${styles.pricingBtn}`}
          >
            {pricingPending ? `💰 ${t('mission.scoutingPrices')}` : `💰 ${t('mission.runPriceIntel')}`}
          </button>
        </div>

        {/* Category template */}
        <div className={styles.section}>
          <CategoryTemplate category={mission.category} />
        </div>

        {/* Travel search — shown only for travel missions */}
        {mission.category === 'travel' && (
          <div className={styles.section}>
            <TravelSearchWidget />
          </div>
        )}

        {/* Collaborators */}
        <div className={styles.section}>
          <CollaboratorsPanel missionId={mission.id} />
        </div>

        {/* Constraints */}
        {mission.constraints.length > 0 && (
          <div className={`${styles.card} ${styles.section}`}>
            <h3 className={styles.cardLabel}>{t('mission.constraints').toUpperCase()}</h3>
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

        {/* Smart Budget Forecast */}
        <div className={styles.section}>
          <ForecastPanel missionId={mission.id} />
        </div>

        {/* AI Purchase Timing */}
        <div className={styles.section}>
          <TimingAdvisorCard missionId={mission.id} />
        </div>

        {/* Export Mission */}
        <div className={styles.section}>
          <ExportPanel missionId={mission.id} />
        </div>

        {/* Mission Timeline */}
        <div className={`${styles.card} ${styles.section}`}>
          <h3 className={styles.cardLabel}>{t('mission.missionProgress').toUpperCase()}</h3>
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
            <div className={`${styles.quickNavTitle} ${styles.quickNavTitleOptions}`}>{t('quickNav.optionsExplorer')}</div>
            <div className={styles.quickNavDesc}>{t('quickNav.optionsDesc')}</div>
          </button>
          <button
            onClick={() => navigate(`/missions/${slug}/shopping`)}
            className={styles.quickNavBtn}
          >
            <div className={styles.quickNavIcon}>🛒</div>
            <div className={`${styles.quickNavTitle} ${styles.quickNavTitleShopping}`}>{t('quickNav.shoppingTracker')}</div>
            <div className={styles.quickNavDesc}>{t('quickNav.shoppingDesc')}</div>
          </button>
        </div>
      </div>
    </main>
  )
}
