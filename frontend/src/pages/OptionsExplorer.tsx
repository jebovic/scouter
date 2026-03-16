import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ScouterGrid, EmptyState, SkeletonGrid, FeedbackModal } from '../components/scouter'
import {
  OptionCard,
  ComparisonTable,
  RadarChart,
  ConstraintChecker,
  LivePricesPanel,
  CompareBar,
  ComparisonPanel,
  ExportButton,
} from '../components/options'
import { ComparisonMatrix } from '../components/comparison'
import { AgentRunHistory } from '../components/agentrun'
import {
  useMission,
  useOptions,
  usePinOption,
  useRejectOption,
  useUnrejectOption,
  useDeletePinnedOptions,
  useResearch,
  useDecision,
  useAgentRuns,
  useComparisonMode,
} from '../hooks'
import type { OptionBadge } from '../types'
import styles from './OptionsExplorer.module.css'

type ViewMode = 'grid' | 'compare'

const BADGE_FILTER_VALUES: (OptionBadge | 'all')[] = ['all', 'recommended', 'alternative', 'watch', 'rejected']

export default function OptionsExplorer() {
  const { slug } = useParams<{ slug: string }>()
  const { t } = useTranslation()
  const { mission, isLoading: missionLoading } = useMission(slug!)
  const { options, isLoading: optionsLoading } = useOptions(mission?.id ?? '')
  const { triggerResearch, isPending: researchPending } = useResearch(mission?.id ?? '')
  const { pinOption } = usePinOption(mission?.id ?? '')
  const { rejectOption } = useRejectOption(mission?.id ?? '')
  const { unrejectOption } = useUnrejectOption(mission?.id ?? '')
  const { deletePinnedOptions } = useDeletePinnedOptions(mission?.id ?? '')
  const { decision } = useDecision(mission?.id ?? '')
  const { runs, isLoading: runsLoading } = useAgentRuns(mission?.id ?? '', 'research')
  const [viewMode, setViewMode] = useState<ViewMode>('grid')
  const [badgeFilter, setBadgeFilter] = useState<OptionBadge | 'all'>('all')
  const [showFeedback, setShowFeedback] = useState(false)
  const [rejectTarget, setRejectTarget] = useState<string | null>(null)
  const [showComparisonPanel, setShowComparisonPanel] = useState(false)
  const { isCompareMode, setIsCompareMode, selectedIds, toggle, clearSelection } = useComparisonMode()

  const isLoading = missionLoading || optionsLoading

  const filtered = badgeFilter === 'all' ? options : options.filter((o) => o.badge === badgeFilter)
  const pinnedCount = options.filter((o) => o.pinned).length
  const selectionFull = selectedIds.size >= 4
  const selectedOptions = options.filter((o) => selectedIds.has(o.id))

  if (isLoading) {
    return (
      <main className={styles.main}>
        <div className="container">
          <ScouterGrid cols={2}>
            <SkeletonGrid count={4} />
          </ScouterGrid>
        </div>
      </main>
    )
  }

  return (
    <>
    <main className={styles.main}>
        <div className="container">
          {/* Header */}
          <div className={styles.header}>
            <div>
              <h1 className={styles.title}>{t('nav.options')}</h1>
              <p className={styles.subtitle}>
                {t('options.optionCount', { count: filtered.length })}
                {badgeFilter !== 'all' ? ` (${t(`badges.${badgeFilter}`, { defaultValue: badgeFilter })})` : ''}
              </p>
            </div>

            <div className={styles.actions}>
              {/* Compare mode toggle */}
              <button
                onClick={() => {
                  setIsCompareMode(!isCompareMode)
                  if (isCompareMode) {
                    clearSelection()
                    setShowComparisonPanel(false)
                  }
                }}
                className={`${styles.viewToggleBtn}${isCompareMode ? ` ${styles.active}` : ''}`}
                style={{ border: '1px solid var(--border)', borderRadius: 'var(--radius-sm)', padding: '6px 14px' }}
              >
                <span aria-hidden="true">⊞↔</span> {t('options.compareMode', { defaultValue: 'Mode Comparaison' })}
              </button>

              {/* View mode toggle */}
              <div className={styles.viewToggle}>
                {(['grid', 'compare'] as ViewMode[]).map((mode) => (
                  <button
                    key={mode}
                    onClick={() => setViewMode(mode)}
                    className={`${styles.viewToggleBtn}${viewMode === mode ? ` ${styles.active}` : ''}`}
                  >
                    {mode === 'grid'
                      ? <><span aria-hidden="true">⊞</span> {t('options.grid')}</>
                      : <><span aria-hidden="true">⊟</span> {t('options.compare')}</>}
                  </button>
                ))}
              </div>

              {/* Clear pinned */}
              {pinnedCount > 0 && (
                <button
                  onClick={() => deletePinnedOptions()}
                  className={styles.clearPinnedBtn}
                >
                  {t('options.clearPinned', { count: pinnedCount })}
                </button>
              )}

              {/* Export CSV */}
              {mission && <ExportButton slug={mission.slug} />}

              {/* Research button */}
              <button
                onClick={() => setShowFeedback(true)}
                disabled={researchPending}
                className={styles.researchBtn}
              >
                <span aria-hidden="true">⚡</span>{researchPending ? ` ${t('options.running')}` : ` ${t('options.rerunResearch')}`}
              </button>
            </div>
          </div>

          {/* Badge filters */}
          <div className={styles.filterBar}>
            {BADGE_FILTER_VALUES.map((value) => (
              <button
                key={value}
                onClick={() => setBadgeFilter(value)}
                className={`${styles.filterBtn}${badgeFilter === value ? ` ${styles.active}` : ''}`}
              >
                {value === 'all' ? t('options.all') : t(`badges.${value}`, { defaultValue: value })}
              </button>
            ))}
          </div>

          {/* Agent run history */}
          {(runs.length > 0 || runsLoading) && (
            <div className={styles.historySection}>
              <AgentRunHistory runs={runs} isLoading={runsLoading} />
            </div>
          )}

          {/* Empty state */}
          {options.length === 0 ? (
            <EmptyState
              icon="🔍"
              title={t('research.noOptions').toUpperCase()}
              description={t('research.noOptionsDesc')}
              actionLabel={t('research.runResearch').toUpperCase()}
              onAction={() => setShowFeedback(true)}
            />
          ) : viewMode === 'compare' ? (
            <div className={styles.compareSection}>
              <ComparisonTable options={filtered} />
              {mission && (
                <div className={styles.constraintList}>
                  {filtered.map((o) => (
                    <div key={o.id} className={styles.constraintItem}>
                      <div className={styles.constraintLabel}>{o.name}</div>
                      <ConstraintChecker option={o} constraints={mission.constraints} />
                    </div>
                  ))}
                </div>
              )}
              <div className={styles.radarWrapper}>
                <RadarChart options={filtered} />
              </div>
            </div>
          ) : (
            <div>
              {/* Radar chart above grid when ≥3 options */}
              {filtered.length >= 3 && (
                <div className={styles.radarWrapperAbove}>
                  <RadarChart options={filtered} />
                </div>
              )}
              {/* Comparison panel (inline, above grid when open) */}
              {isCompareMode && showComparisonPanel && selectedOptions.length >= 2 && (
                <ComparisonPanel
                  options={selectedOptions}
                  onClose={() => setShowComparisonPanel(false)}
                />
              )}

              <div className={styles.optionsGrid}>
                {filtered.map((option) => (
                  <div key={option.id}>
                    <OptionCard
                      option={option}
                      missionId={mission?.id ?? ''}
                      score={decision?.scores.find((s) => s.optionId === option.id)?.score}
                      onPin={(id) => pinOption(id)}
                      onReject={(id) => setRejectTarget(id)}
                      onUnreject={(id) => unrejectOption(id)}
                      compareMode={isCompareMode}
                      selected={selectedIds.has(option.id)}
                      onToggleSelect={toggle}
                      selectionFull={selectionFull}
                    />
                    {mission && (
                      <LivePricesPanel
                        missionId={mission.id}
                        optionId={option.id}
                        missionCategory={mission.category}
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Weighted Comparison Matrix — always shown when mission is loaded */}
          {mission && (
            <div className={styles.matrixSection}>
              <ComparisonMatrix missionId={mission.id} />
            </div>
          )}
        </div>
      </main>

      {/* Sticky compare bar */}
      <CompareBar
        isCompareMode={isCompareMode}
        selectedCount={selectedIds.size}
        onCompare={() => setShowComparisonPanel(true)}
        onClear={() => {
          clearSelection()
          setShowComparisonPanel(false)
        }}
      />

      {showFeedback && (
        <FeedbackModal
          title={t('feedbackModal.rerunResearch')}
          placeholder={t('feedbackModal.rerunResearchPlaceholder')}
          onConfirm={async (feedback) => {
            try {
              await triggerResearch(feedback || undefined)
            } finally {
              setShowFeedback(false)
            }
          }}
          onClose={() => setShowFeedback(false)}
          isPending={researchPending}
        />
      )}

      {rejectTarget && (
        <FeedbackModal
          title={t('feedbackModal.rejectOption')}
          placeholder={t('feedbackModal.rejectPlaceholder')}
          onConfirm={(reason) => {
            rejectOption({ optionId: rejectTarget, reason: reason || 'Rejected by user' })
            setRejectTarget(null)
          }}
          onClose={() => setRejectTarget(null)}
        />
      )}
    </>
  )
}
