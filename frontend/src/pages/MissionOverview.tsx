import { useMemo, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LoadingPulse, BudgetBar, StatusBadge, ToastContainer, useToasts } from '../components/scouter'
import { useBudgetAlerts } from '../hooks/useBudgetAlerts'
import { CategoryTemplate, DecisionPanel, MissionTimeline, PurchaseForm, LessonsField, CollaboratorsPanel, TravelSearchWidget, TimingAdvisorCard, ExportPanel, ReceiptScanner, SummaryReport, CoachPanel, HealthScoreCard, MissionSummaryCard, CommentThread, CategoryBadge, MissionGoalTracker, BudgetRecommendations, SalesCalendar, EcoScorePanel, MissionProgressWidget, GiftFinderWidget, LoyaltySummaryPanel, MissionROICard, InflationTrackerPanel, DecisionMatrixTable, SmartAlertsPanel, VoteSummaryPanel, MissionReportButton, ReorderSuggestionsPanel, NegotiationOutcomePanel, BundleDealsPanel, BurnRateCard, RegretAnalyzerCard, ListOptimizerPanel, CashbackSummaryPanel, SeasonalCalendarPanel } from '../components/mission'
import { ForecastPanel } from '../components/forecast'
import { useMission, useShopping, useResearch, usePriceIntel, useUpdateMission, useKeyboardShortcuts, usePurchaseRecord, useSuggestCategory } from '../hooks'
import type { MissionPhase } from '../types'
import type { PurchaseFormPrefill } from '../components/mission/PurchaseForm'
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
  const [showScanner, setShowScanner] = useState(false)
  const [scanPrefill, setScanPrefill] = useState<PurchaseFormPrefill | undefined>(undefined)
  const [scanKey, setScanKey] = useState(0)
  const { mutate: suggestCategory, isPending: isSuggesting, data: categorySuggestion } = useSuggestCategory(slug ?? '')

  const spent = items.reduce((sum, i) => sum + i.price, 0)

  const { toasts, addToast, removeToast } = useToasts()
  const budget = mission?.budget ?? 0
  useBudgetAlerts(spent, budget, items.length, addToast)

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
              {(mission.autotagCategory || categorySuggestion) && (
                <CategoryBadge category={categorySuggestion?.category ?? mission.autotagCategory ?? undefined} />
              )}
              <button
                className={styles.autotagBtn}
                onClick={() => suggestCategory()}
                disabled={isSuggesting}
                title="Suggérer une catégorie"
                aria-label="Suggérer une catégorie avec l'IA"
              >
                {isSuggesting ? '⏳' : '🏷️'}
              </button>
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

        {/* Budget Burn Rate Tracker (Phase 156) */}
        <div className={styles.section}>
          <BurnRateCard missionId={mission.id} currency={mission.currency} />
        </div>

        {/* Smart Shopping List Optimizer (Phase 159) */}
        <div className={styles.section}>
          <ListOptimizerPanel missionId={mission.id} />
        </div>

        {/* French Cashback & Reward Tracker (Phase 160) */}
        <div className={styles.section}>
          <CashbackSummaryPanel missionId={mission.id} currency={mission.currency} />
        </div>

        {/* Purchase Regret Analyzer (Phase 158) */}
        <div className={styles.section}>
          <RegretAnalyzerCard missionId={mission.id} />
        </div>

        {/* Smart Notification Rules Engine (Phase 143) */}
        <div className={styles.section}>
          <SmartAlertsPanel missionId={mission.id} />
        </div>

        {/* Smart Reorder & Repurchase Suggestions (Phase 148) */}
        <ReorderSuggestionsPanel missionId={mission.id} currency={mission.currency} />

        {/* Mission Progress Widget (Phase 117) */}
        <div className={styles.section}>
          <MissionProgressWidget missionId={mission.id} />
        </div>

        {/* Mission ROI Calculator (Phase 140) */}
        <div className={styles.section}>
          <MissionROICard missionId={mission.id} />
        </div>

        {/* French Inflation Impact Tracker (Phase 141) */}
        <div className={styles.section}>
          <InflationTrackerPanel missionId={mission.id} />
        </div>

        {/* Purchase Decision Matrix (Phase 142) */}
        <div className={styles.section}>
          <h3 className={styles.cardLabel}>📊 Matrice de décision</h3>
          <DecisionMatrixTable missionId={mission.id} />
        </div>

        {/* Collaborative Wishlist Voting (Phase 144) */}
        <div className={styles.section}>
          <VoteSummaryPanel missionId={mission.id} />
        </div>

        {/* Negotiation Outcome Tracker (Phase 151) */}
        <div className={styles.section}>
          <NegotiationOutcomePanel missionId={mission.id} />
        </div>

        {/* Reorder Suggestions (Phase 148) */}
        <div className={styles.section}>
          <ReorderSuggestionsPanel missionId={mission.id} />
        </div>

        {/* Smart Bundle Deal Detector (Phase 152) */}
        <div className={styles.section}>
          <BundleDealsPanel missionId={mission.id} />
        </div>

        {/* Gift Finder Assistant (Phase 122) */}
        <div className={styles.section}>
          <h3 className={styles.cardLabel}>🎁 IDÉES CADEAUX</h3>
          <GiftFinderWidget missionId={mission.id} budget={mission.budget} />
        </div>

        {/* Merchant Loyalty Tracker (Phase 135) */}
        <div className={styles.section}>
          <LoyaltySummaryPanel missionId={mission.id} />
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
          <MissionReportButton missionId={mission.id} />
        </div>

        {/* Mission Goal Tracker (Phase 80) */}
        <div className={styles.section}>
          <MissionGoalTracker
            mission={{
              budget: mission.budget,
              currency: mission.currency,
              createdAt: mission.createdAt,
            }}
            items={items.map(item => ({
              currentPrice: item.price,
              targetPrice: item.targetPrice,
            }))}
          />
        </div>

        {/* Smart Budget Recommendations (Phase 96) */}
        <div className={styles.section}>
          <BudgetRecommendations missionId={mission.id} />
        </div>

        {/* French Sales Calendar (Phase 99) */}
        <div className={styles.section}>
          <SalesCalendar category={mission.category} />
        </div>

        {/* Seasonal Buying Calendar (Phase 162) */}
        <div className={styles.section}>
          <SeasonalCalendarPanel missionId={mission.id} />
        </div>

        {/* Eco-Score & Sustainability Dashboard (Phase 109) */}
        <div className={styles.section}>
          <EcoScorePanel missionId={mission.id} />
        </div>

        {/* Mission Health Score (Phase 57) */}
        <div className={styles.section}>
          <HealthScoreCard missionSlug={mission.slug} />
        </div>

        {/* AI Mission Summary Card (Phase 72) */}
        <div className={styles.section}>
          <MissionSummaryCard missionSlug={mission.slug} />
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

        {/* AI Shopping Summary Report (Phase 43) */}
        <div className={styles.section}>
          <SummaryReport missionId={mission.id} />
        </div>

        {/* AI Coach (Phase 47) */}
        <div className={styles.section}>
          <CoachPanel missionSlug={mission.slug} />
        </div>

        {/* Activity Timeline */}
        <div className={`${styles.card} ${styles.section}`}>
          <h3 className={styles.cardLabel}>{t('mission.activityFeed').toUpperCase()}</h3>
          <MissionTimeline
            missionId={mission.id}
            missionSlug={mission.slug}
          />
        </div>

        {/* Purchase section — visible in buying/done phases */}
        {(mission.phase === 'buying' || mission.phase === 'done') && (
          <div className={styles.section}>
            {!purchaseRecord && (
              <div className={styles.scannerTrigger}>
                <button
                  className={styles.scannerBtn}
                  onClick={() => setShowScanner(true)}
                  title="Scanner un reçu pour pré-remplir le formulaire"
                >
                  📷 Scanner un reçu
                </button>
              </div>
            )}
            <PurchaseForm
              key={scanKey}
              missionId={mission.id}
              existingRecord={purchaseRecord}
              prefill={scanPrefill}
            />
          </div>
        )}

        {/* Receipt scanner modal */}
        {showScanner && (
          <ReceiptScanner
            onResult={(data) => {
              setScanPrefill({
                merchant: data.merchant,
                finalPrice: data.amount,
                purchasedAt: data.date
                  ? (() => {
                      // Convert DD/MM/YYYY or DD-MM-YYYY → YYYY-MM-DD for date input
                      const parts = data.date.split(/[\/\-]/)
                      return parts.length === 3 ? `${parts[2]}-${parts[1]}-${parts[0]}` : undefined
                    })()
                  : undefined,
              })
              setScanKey((k) => k + 1)
              setShowScanner(false)
            }}
            onClose={() => setShowScanner(false)}
          />
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

        {/* Comment threads */}
        <div className={styles.section}>
          <CommentThread missionSlug={mission.slug} />
        </div>

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
      <ToastContainer toasts={toasts} onRemove={removeToast} />
    </main>
  )
}
