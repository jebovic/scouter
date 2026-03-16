import { useState, useMemo, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topnav, ScouterGrid, UsageWidget, EmptyState, SkeletonGrid, DaySummaryStrip, EtatDuJourCard, LastVisitCard, BadgeRow, BadgeToast, ShoppingPersonaCard } from '../components/scouter'
import { MissionCard, MissionForm, TemplateGallery, TemplatePreview, BudgetRollupWidget, MissionComparison, BudgetAlertBanner, PriceDigestWidget, ActivityFeedPanel, WeeklyDigestPanel, SaleCalendarWidget, HolidayAlertBanner } from '../components/mission'
import { PersonaCard } from '../components/persona'
import { useMissions, useCreateMission, useKeyboardShortcuts, useTemplates, useDealCalendar, useMissionComparison, useBadges } from '../hooks'
import { useStats } from '../hooks/usePurchase'
import { unlockIfEarned, ALL_BADGES } from '../utils/badges'
import type { Badge } from '../utils/badges'
import type { MissionCreateRequest, Template } from '../types'
import type { DealEvent } from '../api/dealCalendar'
import styles from './HQDashboard.module.css'

function findNextDeal(events: DealEvent[]): DealEvent | null {
  const now = Date.now()
  const upcoming = events
    .filter((e) => new Date(e.startDate).getTime() > now)
    .sort((a, b) => new Date(a.startDate).getTime() - new Date(b.startDate).getTime())
  return upcoming[0] ?? null
}

function daysUntil(isoDate: string): number {
  const ms = new Date(isoDate).getTime() - Date.now()
  return Math.max(0, Math.ceil(ms / (1000 * 60 * 60 * 24)))
}

function NextDealWidget() {
  const { t } = useTranslation()
  const { events, isLoading } = useDealCalendar({ upcoming: true })
  const next = findNextDeal(events)

  if (isLoading) {
    return <div className={styles.nextDeal} style={{ opacity: 0.4 }} />
  }

  if (!next) return null

  const days = daysUntil(next.startDate)
  const dateStr = new Intl.DateTimeFormat('fr-FR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(next.startDate))

  const levelColor =
    next.discountLevel === 'high'
      ? 'var(--coral)'
      : next.discountLevel === 'medium'
        ? 'var(--gold)'
        : 'var(--green)'

  return (
    <div className={styles.nextDeal}>
      <span className={styles.nextDealLabel}>{t('hq.nextDealLabel')}</span>
      <div className={styles.nextDealBody}>
        <span className={styles.nextDealName}>{next.name}</span>
        <span
          className={styles.nextDealCountdown}
          style={{ color: levelColor } as React.CSSProperties}
        >
          J-{days}
        </span>
      </div>
      <div className={styles.nextDealMeta}>
        {dateStr}
        <Link to="/deal-calendar" className={styles.nextDealLink}>
          {t('hq.viewCalendar')}
        </Link>
      </div>
    </div>
  )
}

export default function HQDashboard() {
  const { t } = useTranslation()
  const [showArchived, setShowArchived] = useState(false)
  const { missions, isLoading } = useMissions({ includeArchived: showArchived })
  const { createMission, isPending } = useCreateMission()
  const { templates, isLoading: templatesLoading } = useTemplates()
  const navigate = useNavigate()
  const [showForm, setShowForm] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [previewTemplate, setPreviewTemplate] = useState<Template | null>(null)
  const [formInitialValues, setFormInitialValues] = useState<Partial<MissionCreateRequest> | undefined>()
  const { data: stats } = useStats()
  const { badges, unlock } = useBadges()
  const [newBadge, setNewBadge] = useState<Badge | null>(null)

  const { selectedIds: compareIds, toggle: toggleCompare, clear: clearCompare, isSelected: isCompareSelected, canAdd: canAddToCompare } = useMissionComparison()

  const shortcuts = useMemo(() => ({
    n: () => setShowForm(true),
  }), [])
  useKeyboardShortcuts(shortcuts)

  useEffect(() => {
    if (!stats) return

    const missionCount = missions.length
    const savedTotal = Math.max(0, stats.savings)
    const researchCount = 0

    const earnedBadgeIds = unlockIfEarned(missionCount, savedTotal, researchCount)

    let firstNewBadge: string | null = null
    earnedBadgeIds.forEach((badgeId) => {
      if (!badges.some((b) => b.id === badgeId)) {
        unlock(badgeId)
        if (!firstNewBadge) {
          firstNewBadge = badgeId
        }
      }
    })

    if (firstNewBadge && firstNewBadge in ALL_BADGES) {
      const badgeKey = firstNewBadge as keyof typeof ALL_BADGES
      const def = ALL_BADGES[badgeKey]
      const badge: Badge = {
        id: def.id as any,
        name: def.name,
        description: def.description,
        icon: def.icon,
        unlockedAt: new Date(),
      }
      setNewBadge(badge)
    }
  }, [stats, missions.length, badges, unlock])

  async function handleCreate(req: MissionCreateRequest) {
    setCreateError(null)
    try {
      const mission = await createMission(req)
      setShowForm(false)
      setFormInitialValues(undefined)
      navigate(`/missions/${mission.slug}`)
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create mission')
    }
  }

  function handleApplyTemplate(template: Template) {
    setPreviewTemplate(null)
    setFormInitialValues({
      name: '',
      icon: template.icon,
      category: template.category,
      constraints: template.constraints,
      costCategories: template.costCategories,
      budget: template.suggestedBudget?.min,
    })
    setShowForm(true)
  }

  return (
    <div className="page grid-bg scanlines">
      <Topnav />
      <main className={styles.main}>
        <div className="container">
          <div className={styles.header}>
            <div>
              <h1 className={styles.title}>{t('nav.missionControl')}</h1>
              <p className={styles.subtitle}>
                {t('mission.activeCount', { count: missions.length })}
              </p>
            </div>
            <div className={styles.headerBtns}>
              <button
                className={showArchived ? styles.filterActive : styles.filterBtn}
                aria-pressed={showArchived}
                onClick={() => setShowArchived((s) => !s)}
              >
                {t('mission.actions.showArchived')}
              </button>
              <button className={styles.createBtn} onClick={() => setShowForm(true)}>
                + {t('mission.create')}
              </button>
            </div>
          </div>

          {showForm && (
            <div
              className={styles.overlay}
              onClick={(e) => e.target === e.currentTarget && setShowForm(false)}
            >
              <MissionForm
                key={formInitialValues?.icon ?? 'blank'}
                onSubmit={handleCreate}
                onCancel={() => { setShowForm(false); setFormInitialValues(undefined) }}
                loading={isPending}
                error={createError ?? undefined}
                initialValues={formInitialValues}
              />
            </div>
          )}

          {previewTemplate && (
            <TemplatePreview
              template={previewTemplate}
              onApply={handleApplyTemplate}
              onClose={() => setPreviewTemplate(null)}
            />
          )}

          <DaySummaryStrip />

          <EtatDuJourCard />

          <LastVisitCard />

          <HolidayAlertBanner />

          <BudgetAlertBanner />

          <PriceDigestWidget />

          <div className={styles.usageWrapper}>
            <UsageWidget />
          </div>

          <NextDealWidget />

          <WeeklyDigestPanel />

          <ActivityFeedPanel />

          <SaleCalendarWidget />

          <BudgetRollupWidget />

          {isLoading ? (
            <ScouterGrid cols={3}>
              <SkeletonGrid count={6} />
            </ScouterGrid>
          ) : missions.length === 0 ? (
            <EmptyState
              icon="🎯"
              title={t('hq.noMissions').toUpperCase()}
              description={t('hq.noMissionsDesc')}
              actionLabel={`+ ${t('hq.newMission').toUpperCase()}`}
              onAction={() => setShowForm(true)}
            />
          ) : (
            <ScouterGrid>
              {missions.map((mission) => (
                <MissionCard key={mission.id} mission={mission} />
              ))}
            </ScouterGrid>
          )}

          {missions.length >= 2 && (
            <div className={styles.compareSection}>
              <div className={styles.compareHeader}>
                <div>
                  <h2 className={styles.compareTitle}>{t('hq.compareMissions')}</h2>
                  <p className={styles.compareSubtitle}>
                    {t('hq.compareSubtitle')}
                  </p>
                </div>
                {compareIds.length > 0 && (
                  <button className={styles.clearBtn} onClick={clearCompare}>
                    {t('hq.compareClear')}
                  </button>
                )}
              </div>
              <div className={styles.compareChips}>
                {missions.map((mission) => {
                  const selected = isCompareSelected(mission.id)
                  const disabled = !canAddToCompare(mission.id)
                  return (
                    <button
                      key={mission.id}
                      className={styles.compareChip}
                      data-selected={selected}
                      data-disabled={disabled}
                      onClick={() => toggleCompare(mission.id)}
                      disabled={disabled}
                      title={disabled ? t('hq.compareMax') : undefined}
                    >
                      <span>{mission.icon}</span>
                      <span>{mission.name}</span>
                      {selected && <span className={styles.chipCheck}>✓</span>}
                    </button>
                  )
                })}
              </div>
              {compareIds.length >= 2 && (
                <MissionComparison missionIds={compareIds} />
              )}
            </div>
          )}

          <div className={styles.templatesSection}>
            <TemplateGallery
              templates={templates}
              onSelect={(t) => setPreviewTemplate(t)}
              isLoading={templatesLoading}
            />
          </div>

          <section className={styles.badgesSection}>
            <h2 className={styles.sectionTitle}>{t('hq.badges')}</h2>
            <BadgeRow badges={badges} />
          </section>

          <section className={styles.personaSection}>
            <ShoppingPersonaCard />
            <PersonaCard />
          </section>
        </div>
      </main>
      {newBadge && <BadgeToast badge={newBadge} onDismiss={() => setNewBadge(null)} />}
    </div>
  )
}
