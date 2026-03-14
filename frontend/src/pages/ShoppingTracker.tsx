import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topnav, LoadingPulse, BudgetBar, EmptyState, Skeleton, FeedbackModal } from '../components/scouter'
import { ShoppingList } from '../components/shopping'
import { CostBreakdown } from '../components/shopping'
import { PriceHistoryChart } from '../components/shopping'
import { AgentRunHistory } from '../components/agentrun'
import {
  useMission,
  useShopping,
  usePinShoppingItem,
  useDeletePinnedShoppingItems,
  usePriceIntel,
  usePriceSnapshots,
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
  }

  if (isLoading) {
    return (
      <div className="page grid-bg scanlines">
        <Topnav missionSlug={slug} />
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
              <h1 className={styles.title}>{t('nav.shopping')}</h1>
              <p className={styles.subtitle}>
                {items.length} item{items.length !== 1 ? 's' : ''} tracked
              </p>
            </div>

            <div className={styles.actions}>
              <button className={styles.addBtn} onClick={() => setShowAddForm(true)}>
                + Add Item
              </button>
              {pinnedCount > 0 && (
                <button
                  className={styles.clearPinnedBtn}
                  onClick={() => deletePinnedItems()}
                >
                  Clear Pinned ({pinnedCount})
                </button>
              )}
              <button
                className={styles.priceIntelBtn}
                onClick={() => setShowFeedback(true)}
                disabled={pricingPending}
              >
                {pricingPending ? '💰 Scouting...' : '💰 Price Intel'}
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
              <div className={styles.modal}>
                <h3 className={styles.modalTitle}>ADD ITEM</h3>
                <div className={styles.formFields}>
                  {[
                    { key: 'name', label: 'Item Name', type: 'text', required: true },
                    { key: 'merchant', label: 'Merchant', type: 'text', required: true },
                    { key: 'costCategory', label: 'Category', type: 'text', required: false },
                    { key: 'price', label: 'Price', type: 'number', required: true },
                    { key: 'url', label: 'URL', type: 'url', required: false },
                    { key: 'note', label: 'Note', type: 'text', required: false },
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
                            [key]: type === 'number' ? parseFloat(e.target.value) || 0 : e.target.value,
                          }))
                        }
                      />
                    </div>
                  ))}
                </div>
                <div className={styles.modalActions}>
                  <button className={styles.cancelBtn} onClick={() => setShowAddForm(false)}>
                    Cancel
                  </button>
                  <button
                    className={styles.confirmBtn}
                    onClick={handleAddItem}
                    disabled={createPending || !addForm.name || !addForm.merchant || !addForm.price}
                  >
                    {createPending ? 'Adding...' : 'Add Item'}
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
              title="NO PRICE DATA"
              description="Run Price Intel to track prices for this mission"
              actionLabel="RUN PRICE INTEL"
              onAction={() => setShowFeedback(true)}
            />
          ) : (
            <div className={styles.content}>
              {/* Shopping list */}
              <ShoppingList items={items} onPriceClick={setHistoryItem} onPin={(id) => pinItem(id)} />

              {/* Cost breakdown */}
              {items.length >= 2 && (
                <div className={styles.breakdownCard}>
                  <h3 className={styles.breakdownTitle}>COST BREAKDOWN</h3>
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
          title="RUN PRICE INTEL"
          placeholder='Optional: guide the agent (e.g. "check for Black Friday deals on electronics")'
          onConfirm={(feedback) => {
            triggerPricing(feedback || undefined)
            setShowFeedback(false)
          }}
          onClose={() => setShowFeedback(false)}
          isPending={pricingPending}
        />
      )}
    </div>
  )
}

function PriceHistoryModal({ item, currency, onClose }: { item: ShoppingItem; currency: string; onClose: () => void }) {
  const { snapshots, isLoading } = usePriceSnapshots(item.missionId, item.id)

  return (
    <div
      className={styles.overlay}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className={styles.historyModal}>
        <div className={styles.historyModalHeader}>
          <div>
            <h3 className={styles.historyTitle}>PRICE HISTORY</h3>
            <p className={styles.historySubtitle}>{item.name} @ {item.merchant}</p>
          </div>
          <button className={styles.closeBtn} onClick={onClose}>✕</button>
        </div>
        {isLoading ? (
          <LoadingPulse label="Loading history..." />
        ) : snapshots.length === 0 ? (
          <p className={styles.noHistory}>No price history yet</p>
        ) : (
          <PriceHistoryChart snapshots={snapshots} currency={currency} />
        )}
      </div>
    </div>
  )
}
