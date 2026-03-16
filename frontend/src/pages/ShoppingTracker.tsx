import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { BudgetBar, EmptyState, Skeleton, FeedbackModal } from '../components/scouter'
import { ShoppingList, CostBreakdown, PriceHistoryModal, RetailerRadar, OptimizedPlanPanel, ReceiptAnalyzer, PurchaseTimeline, ShoppingOptimizerPanel, KanbanBoard, WishlistShareCard, BudgetPlannerPanel, MissionCSVExportButton } from '../components/shopping'
import { DuplicateAlert } from '../components/shopping/DuplicateAlert'
import { AgentRunHistory } from '../components/agentrun'
import {
  useMission,
  useShopping,
  usePinShoppingItem,
  useDeletePinnedShoppingItems,
  usePriceIntel,
  useCreateShoppingItem,
  useAgentRuns,
} from '../hooks'
import type { ShoppingItem, ShoppingItemCreateRequest } from '../types'
import { useFrenchBenchmark } from '../hooks/useFrenchBenchmark'
import { useTimelinePlanner } from '../hooks/useTimelinePlanner'
import { useFormatCurrency } from '../hooks/useFormatCurrency'
import styles from './ShoppingTracker.module.css'

function FrenchBenchmarkPanel({ missionId }: { missionId: string }) {
  const { data, isLoading } = useFrenchBenchmark(missionId)
  const { fmt } = useFormatCurrency()
  const { t } = useTranslation()
  if (isLoading) return <div style={{ padding: '10px', color: 'var(--text-dim)', fontSize: '0.82rem' }}>{t('shopping.frenchBenchmarkLoading')}</div>
  if (!data) return null
  const verdictColor = data.verdictCode === 'bon_prix'
    ? 'var(--status-buy)'
    : data.verdictCode === 'au_dessus_du_marche'
      ? 'var(--status-crisis)'
      : 'var(--text-dim)'
  return (
    <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: '12px', padding: '16px', marginBottom: '16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '10px' }}>
        <span style={{ fontSize: '1.1rem' }}>🇫🇷</span>
        <span style={{ fontWeight: 600, color: 'var(--text)', fontSize: '0.95rem' }}>{t('shopping.frenchBenchmarkTitle')}</span>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '12px', marginBottom: '12px' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text)' }}>{fmt(data.medianPrice)}</div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>{t('shopping.medianMarket')}</div>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--accent)' }}>{fmt(data.yourAvgPrice)}</div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>{t('shopping.yourAverage')}</div>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '0.85rem', fontWeight: 700, color: verdictColor }}>{data.verdictCode === 'bon_prix' ? t('shopping.verdictBonPrix') : data.verdictCode === 'au_dessus_du_marche' ? t('shopping.verdictEleve') : t('shopping.verdictMoyen')}</div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>{t('shopping.offersCount', { count: data.sampleSize })}</div>
        </div>
      </div>
      <p style={{ fontSize: '0.8rem', color: verdictColor, fontWeight: 500, marginBottom: '8px' }}>{data.verdict}</p>
      {data.tips.length > 0 && (
        <ul style={{ fontSize: '0.78rem', color: 'var(--text-dim)', paddingLeft: '16px', margin: 0 }}>
          {data.tips.map((tip, i) => <li key={i}>{tip}</li>)}
        </ul>
      )}
    </div>
  )
}

function PurchaseTimelineCard({ missionId }: { missionId: string }) {
  const { data, isLoading } = useTimelinePlanner(missionId)
  const { t } = useTranslation()
  if (isLoading) return null
  if (!data || data.totalItems === 0) return null
  return (
    <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: '12px', padding: '16px', marginBottom: '16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
        <span style={{ fontSize: '1.1rem' }}>📅</span>
        <span style={{ fontWeight: 600, color: 'var(--text)', fontSize: '0.95rem' }}>{t('shopping.timelineTitle')}</span>
        <span style={{ marginLeft: 'auto', fontSize: '0.75rem', color: 'var(--text-dim)' }}>{t('shopping.totalItems', { count: data.totalItems })}</span>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '8px' }}>
        {data.buckets.map((b) => (
          <div key={b.week} style={{
            background: 'var(--surface-alt)',
            borderRadius: '8px',
            padding: '10px',
            border: b.itemCount > 0 ? '1px solid var(--accent)' : '1px solid var(--border)',
          }}>
            <div style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-dim)', marginBottom: '4px' }}>{b.label}</div>
            <div style={{ fontSize: '0.7rem', color: 'var(--accent)', marginBottom: '6px' }}>{b.action}</div>
            {b.items.length > 0
              ? b.items.map((item, i) => (
                  <div key={i} style={{ fontSize: '0.75rem', color: 'var(--text)', padding: '2px 0' }}>• {item}</div>
                ))
              : <div style={{ fontSize: '0.72rem', color: 'var(--text-dim)', fontStyle: 'italic' }}>{t('shopping.noItemsInBucket')}</div>
            }
            <div style={{ marginTop: '6px', fontSize: '0.68rem', color: 'var(--text-dim)', fontStyle: 'italic' }}>{b.promoHint}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function ShoppingTracker() {
  const { slug } = useParams<{ slug: string }>()
  const { t } = useTranslation()
  const { mission, isLoading: missionLoading } = useMission(slug!)
  const { items, isLoading: itemsLoading } = useShopping(mission?.id ?? '')
  const { triggerPricing, isPending: pricingPending } = usePriceIntel(mission?.id ?? '')
  const { pinItem } = usePinShoppingItem(mission?.id ?? '')
  const { deletePinnedItems } = useDeletePinnedShoppingItems(mission?.id ?? '')
  const { runs, isLoading: runsLoading } = useAgentRuns(mission?.id ?? '', 'pricing')
  const { createItem, isPending: createPending } = useCreateShoppingItem(mission?.id ?? '')
  const [historyItem, setHistoryItem] = useState<ShoppingItem | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [showFeedback, setShowFeedback] = useState(false)
  const [addForm, setAddForm] = useState<Partial<ShoppingItemCreateRequest>>({})
  const [viewMode, setViewMode] = useState<'list' | 'kanban'>('list')
  const [showShareCard, setShowShareCard] = useState(false)
  const [showBudgetPlanner, setShowBudgetPlanner] = useState(false)

  const isLoading = missionLoading || itemsLoading

  const spent = items.reduce((sum, i) => sum + i.price, 0)
  const pinnedCount = items.filter((i) => i.pinned).length

  async function handleAddItem() {
    if (!addForm.name || !addForm.merchant || !addForm.price) return
    try {
      await createItem({
        name: addForm.name,
        merchant: addForm.merchant,
        costCategory: addForm.costCategory ?? 'other',
        price: addForm.price,
        status: 'watch',
        note: addForm.note,
        url: addForm.url,
      })
      setAddForm({})
      setShowAddForm(false)
    } catch {
      // error handled by mutation's onError toast
    }
  }

  if (isLoading) {
    return (
      <main className={styles.main}>
        <div className="container">
          <div className={styles.skeletonList}>
            <Skeleton variant="row" />
            <Skeleton variant="row" />
            <Skeleton variant="row" />
            <Skeleton variant="row" />
          </div>
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
              <h1 className={styles.title}>{t('nav.shopping')}</h1>
              <p className={styles.subtitle}>
                {t('shopping.itemCount', { count: items.length })}
              </p>
            </div>

            <div className={styles.actions}>
              <button className={styles.addBtn} onClick={() => setShowAddForm(true)}>
                + {t('shopping.addItem')}
              </button>
              {pinnedCount > 0 && (
                <button
                  className={styles.clearPinnedBtn}
                  onClick={() => deletePinnedItems()}
                >
                  {t('shopping.clearPinned', { count: pinnedCount })}
                </button>
              )}
              <button
                className={styles.priceIntelBtn}
                onClick={() => setShowFeedback(true)}
                disabled={pricingPending}
              >
                <span aria-hidden="true">💰</span>{pricingPending ? ` ${t('pricing.scouting')}` : ` ${t('pricing.priceIntel')}`}
              </button>
              <button
                className={styles.addBtn}
                onClick={() => setShowShareCard((v) => !v)}
                aria-pressed={showShareCard}
              >
                {showShareCard ? t('shopping.hideShareCard') : t('shopping.shareWishlist')}
              </button>
              {items.length > 0 && mission && (
                <button
                  className={styles.addBtn}
                  onClick={() => setShowBudgetPlanner((v) => !v)}
                  aria-pressed={showBudgetPlanner}
                >
                  📋 Plan
                </button>
              )}
              {mission && <MissionCSVExportButton missionId={mission.id} />}
              {/* View mode toggle */}
              <div className={styles.viewToggle} role="group" aria-label={t('shopping.viewModeLabel')}>
                <button
                  className={`${styles.viewToggleBtn} ${viewMode === 'list' ? styles.viewToggleActive : ''}`}
                  onClick={() => setViewMode('list')}
                  aria-pressed={viewMode === 'list'}
                  title={t('shopping.viewListTitle')}
                >
                  {t('shopping.viewList')}
                </button>
                <button
                  className={`${styles.viewToggleBtn} ${viewMode === 'kanban' ? styles.viewToggleActive : ''}`}
                  onClick={() => setViewMode('kanban')}
                  aria-pressed={viewMode === 'kanban'}
                  title={t('shopping.viewKanbanTitle')}
                >
                  {t('shopping.viewKanban')}
                </button>
              </div>
            </div>
          </div>

          {/* Budget bar */}
          {mission && (
            <div className={styles.budgetCard}>
              <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />
            </div>
          )}

          {/* Wishlist share card */}
          {showShareCard && mission && (
            <WishlistShareCard missionId={mission.id} currency={mission.currency} />
          )}

          {/* Add item form modal */}
          {showAddForm && (
            <div
              className={styles.overlay}
              onClick={(e) => e.target === e.currentTarget && setShowAddForm(false)}
            >
              <div
                className={styles.modal}
                role="dialog"
                aria-modal="true"
                aria-labelledby="add-item-modal-title"
              >
                <h3 id="add-item-modal-title" className={styles.modalTitle}>{t('shopping.addItemTitle').toUpperCase()}</h3>
                <div className={styles.formFields}>
                  {[
                    { key: 'name', label: t('shopping.itemName'), type: 'text', required: true },
                    { key: 'merchant', label: t('shopping.merchant'), type: 'text', required: true },
                    { key: 'costCategory', label: t('shopping.costCategory'), type: 'text', required: false },
                    { key: 'price', label: t('shopping.price'), type: 'number', required: true },
                    { key: 'url', label: t('shopping.url'), type: 'url', required: false },
                    { key: 'note', label: t('shopping.note'), type: 'text', required: false },
                  ].map(({ key, label, type, required }) => (
                    <div key={key}>
                      <label className={styles.fieldLabel}>
                        {label}{required && ' *'}
                      </label>
                      <input
                        type={type}
                        className={styles.fieldInput}
                        value={String(addForm[key as keyof typeof addForm] ?? '')}
                        onChange={(e) =>
                          setAddForm((prev) => ({
                            ...prev,
                            [key]: type === 'number'
                              ? (e.target.value === '' ? undefined : parseFloat(e.target.value))
                              : e.target.value,
                          }))
                        }
                      />
                    </div>
                  ))}
                </div>
                <div className={styles.modalActions}>
                  <button className={styles.cancelBtn} onClick={() => setShowAddForm(false)}>
                    {t('common.cancel')}
                  </button>
                  <button
                    className={styles.confirmBtn}
                    onClick={handleAddItem}
                    disabled={createPending || !addForm.name || !addForm.merchant || !addForm.price}
                  >
                    {createPending ? t('common.adding') : t('shopping.addItem')}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Agent run history */}
          {(runs.length > 0 || runsLoading) && (
            <div className={styles.historySection}>
              <AgentRunHistory runs={runs} isLoading={runsLoading} />
            </div>
          )}

          {/* Content */}
          {/* Smart Budget Planner */}
          {showBudgetPlanner && mission && (
            <BudgetPlannerPanel
              missionId={mission.id}
              currency={mission.currency}
              onClose={() => setShowBudgetPlanner(false)}
            />
          )}

          {/* Duplicate item detector */}
          {mission && <DuplicateAlert missionId={mission.id} />}

          {/* Smart Shopping List Optimizer */}
          {items.length > 0 && mission && (
            <ShoppingOptimizerPanel missionId={mission.id} />
          )}

          {items.length === 0 ? (
            <EmptyState
              icon="🛒"
              title={t('pricing.noPriceData').toUpperCase()}
              description={t('pricing.noPriceDataDesc')}
              actionLabel={t('pricing.runPriceIntel').toUpperCase()}
              onAction={() => setShowFeedback(true)}
            />
          ) : viewMode === 'kanban' ? (
            <KanbanBoard missionId={mission?.id ?? ''} currency={mission?.currency ?? 'USD'} />
          ) : (
            <div className={styles.content}>
              {/* Shopping list */}
              <ShoppingList items={items} missionId={mission?.id ?? ''} currency={mission?.currency ?? 'USD'} onPriceClick={setHistoryItem} onPin={(id) => pinItem(id)} />

              {/* Retailer Radar */}
              {mission && <RetailerRadar missionSlug={mission.id} />}

              {/* French Market Benchmark */}
              {mission && <FrenchBenchmarkPanel missionId={mission.id} />}

              {/* Purchase Timeline Planner (4-week plan) */}
              {mission && <PurchaseTimelineCard missionId={mission.id} />}

              {/* Purchase Timeline */}
              {items.length > 0 && <PurchaseTimeline items={items} />}

              {/* AI optimizer */}
              {items.length >= 2 && slug && (
                <OptimizedPlanPanel missionSlug={slug} />
              )}

              {/* Smart Receipt Analyzer */}
              {slug && <ReceiptAnalyzer missionSlug={slug} />}

              {/* Cost breakdown */}
              {items.length >= 2 && (
                <div className={styles.breakdownCard}>
                  <h3 className={styles.breakdownTitle}>{t('shopping.costBreakdown').toUpperCase()}</h3>
                  <CostBreakdown items={items} currency={mission?.currency ?? 'USD'} />
                </div>
              )}
            </div>
          )}
        </div>
      </main>

      {/* Price history modal */}
      {historyItem && (
        <PriceHistoryModal item={historyItem} currency={mission?.currency ?? 'USD'} onClose={() => setHistoryItem(null)} />
      )}

      {showFeedback && (
        <FeedbackModal
          title={t('feedbackModal.runPriceIntel')}
          placeholder={t('feedbackModal.runPriceIntelPlaceholder')}
          onConfirm={async (feedback) => {
            try {
              await triggerPricing(feedback || undefined)
            } finally {
              setShowFeedback(false)
            }
          }}
          onClose={() => setShowFeedback(false)}
          isPending={pricingPending}
        />
      )}
    </>
  )
}

