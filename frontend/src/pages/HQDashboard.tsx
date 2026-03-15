import { useState, useMemo } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topnav, ScouterGrid, UsageWidget, EmptyState, SkeletonGrid } from '../components/scouter'
import { MissionCard, MissionForm, TemplateGallery, TemplatePreview } from '../components/mission'
import { useMissions, useCreateMission, useKeyboardShortcuts, useTemplates, useDealCalendar } from '../hooks'
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
      <span className={styles.nextDealLabel}>PROCHAIN BON PLAN</span>
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
          Voir le calendrier →
        </Link>
      </div>
    </div>
  )
}

export default function HQDashboard() {
  const { t } = useTranslation()
  const { missions, isLoading } = useMissions()
  const { createMission, isPending } = useCreateMission()
  const { templates, isLoading: templatesLoading } = useTemplates()
  const navigate = useNavigate()
  const [showForm, setShowForm] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [previewTemplate, setPreviewTemplate] = useState<Template | null>(null)
  const [formInitialValues, setFormInitialValues] = useState<Partial<MissionCreateRequest> | undefined>()

  const shortcuts = useMemo(() => ({
    n: () => setShowForm(true),
  }), [])
  useKeyboardShortcuts(shortcuts)

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
            <button className={styles.createBtn} onClick={() => setShowForm(true)}>
              + {t('mission.create')}
            </button>
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

          <div className={styles.usageWrapper}>
            <UsageWidget />
          </div>

          <NextDealWidget />

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

          <div className={styles.templatesSection}>
            <TemplateGallery
              templates={templates}
              onSelect={(t) => setPreviewTemplate(t)}
              isLoading={templatesLoading}
            />
          </div>
        </div>
      </main>
    </div>
  )
}
