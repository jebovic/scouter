import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topnav, ScouterGrid, EmptyState, SkeletonGrid, FeedbackModal } from '../components/scouter'
import { OptionCard } from '../components/options'
import { ComparisonTable } from '../components/options'
import { RadarChart } from '../components/options'
import { ConstraintChecker } from '../components/options'
import { AgentRunHistory } from '../components/agentrun'
import {
  useMission,
  useOptions,
  usePinOption,
  useRejectOption,
  useDeletePinnedOptions,
  useResearch,
  useDecision,
  useAgentRuns,
} from '../hooks'
import type { OptionBadge } from '../types'
import styles from './OptionsExplorer.module.css'

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
  const { pinOption } = usePinOption(mission?.id ?? '')
  const { rejectOption } = useRejectOption(mission?.id ?? '')
  const { deletePinnedOptions } = useDeletePinnedOptions(mission?.id ?? '')
  const { decision } = useDecision(mission?.id ?? '')
  const { runs, isLoading: runsLoading } = useAgentRuns(mission?.id ?? '', 'research')
  const [viewMode, setViewMode] = useState<ViewMode>('grid')
  const [badgeFilter, setBadgeFilter] = useState<OptionBadge | 'all'>('all')
  const [showFeedback, setShowFeedback] = useState(false)
  const [rejectTarget, setRejectTarget] = useState<string | null>(null)

  const isLoading = missionLoading || optionsLoading

  const filtered = badgeFilter === 'all' ? options : options.filter((o) => o.badge === badgeFilter)
  const pinnedCount = options.filter((o) => o.pinned).length

  if (isLoading) {
    return (
      <div className="page grid-bg scanlines">
        <Topnav missionSlug={slug} />
        <main className={styles.main}>
          <div className="container">
            <ScouterGrid cols={2}>
              <SkeletonGrid count={4} />
            </ScouterGrid>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="page grid-bg scanlines">
      <Topnav missionSlug={slug} missionName={mission?.name} />
      <main className={styles.main}>
        <div className="container">
          {/* Header */}
          <div className={styles.header}>
            <div>
              <h1 className={styles.title}>{t('nav.options')}</h1>
              <p className={styles.subtitle}>
                {filtered.length} option{filtered.length !== 1 ? 's' : ''}
                {badgeFilter !== 'all' ? ` (${badgeFilter})` : ''}
              </p>
            </div>

            <div className={styles.actions}>
              {/* View mode toggle */}
              <div className={styles.viewToggle}>
                {(['grid', 'compare'] as ViewMode[]).map((mode) => (
                  <button
                    key={mode}
                    onClick={() => setViewMode(mode)}
                    className={`${styles.viewToggleBtn}${viewMode === mode ? ` ${styles.active}` : ''}`}
                  >
                    {mode === 'grid' ? '⊞ Grid' : '⊟ Compare'}
                  </button>
                ))}
              </div>

              {/* Clear pinned */}
              {pinnedCount > 0 && (
                <button
                  onClick={() => deletePinnedOptions()}
                  className={styles.clearPinnedBtn}
                >
                  Clear Pinned ({pinnedCount})
                </button>
              )}

              {/* Research button */}
              <button
                onClick={() => setShowFeedback(true)}
                disabled={researchPending}
                className={styles.researchBtn}
              >
                {researchPending ? '⚡ Running...' : '⚡ Re-run Research'}
              </button>
            </div>
          </div>

          {/* Badge filters */}
          <div className={styles.filterBar}>
            {BADGE_FILTERS.map((f) => (
              <button
                key={f.value}
                onClick={() => setBadgeFilter(f.value)}
                className={`${styles.filterBtn}${badgeFilter === f.value ? ` ${styles.active}` : ''}`}
              >
                {f.label}
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
              title="NO OPTIONS FOUND"
              description="Run the Research Agent to discover options for this mission"
              actionLabel="RUN RESEARCH"
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
              <div className={styles.optionsGrid}>
                {filtered.map((option) => (
                  <OptionCard
                    key={option.id}
                    option={option}
                    score={decision?.scores.find((s) => s.optionId === option.id)?.score}
                    onPin={(id) => pinOption(id)}
                    onReject={(id) => setRejectTarget(id)}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </main>

      {showFeedback && (
        <FeedbackModal
          title="RE-RUN RESEARCH"
          placeholder='Optional: guide the agent (e.g. "focus on options under $800 with good warranty")'
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
          title="REJECT OPTION"
          placeholder="Reason for rejection (e.g. 'too expensive', 'not available in my region')"
          onConfirm={(reason) => {
            rejectOption({ optionId: rejectTarget, reason: reason || 'Rejected by user' })
            setRejectTarget(null)
          }}
          onClose={() => setRejectTarget(null)}
        />
      )}
    </div>
  )
}
