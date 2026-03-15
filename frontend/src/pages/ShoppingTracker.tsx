import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { BudgetBar, EmptyState, Skeleton, FeedbackModal } from '../components/scouter'
import { ShoppingList, CostBreakdown, PriceHistoryModal, RetailerRadar } from '../components/shopping'
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
import styles from './ShoppingTracker.module.css'

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
            </div>
          </div>

          {/* Budget bar */}
          {mission && (
            <div className={styles.budgetCard}>
              <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />
            </div>
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
          {items.length === 0 ? (
            <EmptyState
              icon="🛒"
              title={t('pricing.noPriceData').toUpperCase()}
              description={t('pricing.noPriceDataDesc')}
              actionLabel={t('pricing.runPriceIntel').toUpperCase()}
              onAction={() => setShowFeedback(true)}
            />
          ) : (
            <div className={styles.content}>
              {/* Shopping list */}
              <ShoppingList items={items} missionId={mission?.id ?? ''} currency={mission?.currency ?? 'USD'} onPriceClick={setHistoryItem} onPin={(id) => pinItem(id)} />

              {/* Retailer Radar */}
              {mission && <RetailerRadar missionSlug={mission.id} />}

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

