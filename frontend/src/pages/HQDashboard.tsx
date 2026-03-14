import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Topnav, ScouterGrid, UsageWidget, EmptyState, SkeletonGrid } from '../components/scouter'
import { MissionCard, MissionForm, TemplateGallery, TemplatePreview } from '../components/mission'
import { useMissions, useCreateMission, useKeyboardShortcuts, useTemplates } from '../hooks'
import { useNavigate } from 'react-router-dom'
import type { MissionCreateRequest, Template } from '../types'
import styles from './HQDashboard.module.css'

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
                {missions.length} active mission{missions.length !== 1 ? 's' : ''}
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

          {isLoading ? (
            <ScouterGrid cols={3}>
              <SkeletonGrid count={6} />
            </ScouterGrid>
          ) : missions.length === 0 ? (
            <EmptyState
              icon="🎯"
              title="NO ACTIVE MISSIONS"
              description="Create your first mission to start researching a purchase"
              actionLabel="+ NEW MISSION"
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
