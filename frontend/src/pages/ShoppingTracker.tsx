import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Topnav, LoadingPulse, BudgetBar } from '../components/scouter'
import { ShoppingList } from '../components/shopping'
import { CostBreakdown } from '../components/shopping'
import { PriceHistoryChart } from '../components/shopping'
import { useMission, useShopping, usePriceIntel, usePriceSnapshots, useCreateShoppingItem } from '../hooks'
import type { ShoppingItem, ShoppingItemCreateRequest } from '../types'

export default function ShoppingTracker() {
  const { slug } = useParams<{ slug: string }>()
  const { t } = useTranslation()
  const { mission, isLoading: missionLoading } = useMission(slug!)
  const { items, isLoading: itemsLoading } = useShopping(mission?.id ?? '')
  const { triggerPricing, isPending: pricingPending } = usePriceIntel(mission?.id ?? '')
  const { createItem, isPending: createPending } = useCreateShoppingItem(mission?.id ?? '')
  const [historyItem, setHistoryItem] = useState<ShoppingItem | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [addForm, setAddForm] = useState<Partial<ShoppingItemCreateRequest>>({})

  const isLoading = missionLoading || itemsLoading

  const spent = items.reduce((sum, i) => sum + i.price, 0)

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
        <main style={{ flex: 1, padding: '2rem', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          <LoadingPulse label="Loading shopping list..." />
        </main>
      </div>
    )
  }

  return (
    <div className="page grid-bg scanlines">
      <Topnav missionSlug={slug} missionName={mission?.name} />
      <main style={{ flex: 1, padding: '2rem' }}>
        <div className="container">
          {/* Header */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: '2rem',
              flexWrap: 'wrap',
              gap: '1rem',
            }}
          >
            <div>
              <h1
                style={{
                  fontFamily: 'var(--font-display)',
                  color: 'var(--gold)',
                  letterSpacing: '0.1em',
                  fontSize: '1.8rem',
                }}
              >
                {t('nav.shopping')}
              </h1>
              <p style={{ color: 'var(--text-dim)', fontSize: '0.85rem', marginTop: 4 }}>
                {items.length} item{items.length !== 1 ? 's' : ''} tracked
              </p>
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
              <button
                onClick={() => setShowAddForm(true)}
                style={{
                  padding: '8px 18px',
                  borderRadius: 'var(--radius-sm)',
                  border: '1px solid var(--cyan)',
                  background: 'var(--cyan-dim)',
                  color: 'var(--cyan)',
                  cursor: 'pointer',
                  fontFamily: 'var(--font-display)',
                  fontSize: '0.8rem',
                  letterSpacing: '0.06em',
                }}
              >
                + Add Item
              </button>
              <button
                onClick={() => triggerPricing()}
                disabled={pricingPending}
                style={{
                  padding: '8px 18px',
                  borderRadius: 'var(--radius-sm)',
                  border: '1px solid var(--purple)',
                  background: 'rgba(139,92,246,0.08)',
                  color: 'var(--purple)',
                  cursor: pricingPending ? 'not-allowed' : 'pointer',
                  fontFamily: 'var(--font-display)',
                  fontSize: '0.8rem',
                  letterSpacing: '0.06em',
                  opacity: pricingPending ? 0.6 : 1,
                }}
              >
                {pricingPending ? '💰 Scouting...' : '💰 Price Intel'}
              </button>
            </div>
          </div>

          {/* Budget bar */}
          {mission && (
            <div
              style={{
                background: 'var(--surface)',
                borderRadius: 'var(--radius)',
                border: '1px solid var(--border)',
                padding: '1.25rem 1.5rem',
                marginBottom: '1.5rem',
              }}
            >
              <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />
            </div>
          )}

          {/* Add item form modal */}
          {showAddForm && (
            <div
              style={{
                position: 'fixed',
                inset: 0,
                background: 'rgba(5,8,16,0.85)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 200,
                padding: '1rem',
              }}
              onClick={(e) => e.target === e.currentTarget && setShowAddForm(false)}
            >
              <div
                style={{
                  background: 'var(--surface)',
                  borderRadius: 'var(--radius)',
                  border: '1px solid var(--border)',
                  padding: '2rem',
                  width: '100%',
                  maxWidth: 480,
                }}
              >
                <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--cyan)', marginBottom: '1.5rem', letterSpacing: '0.1em' }}>
                  ADD ITEM
                </h3>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {[
                    { key: 'name', label: 'Item Name', type: 'text', required: true },
                    { key: 'merchant', label: 'Merchant', type: 'text', required: true },
                    { key: 'costCategory', label: 'Category', type: 'text', required: false },
                    { key: 'price', label: 'Price', type: 'number', required: true },
                    { key: 'url', label: 'URL', type: 'url', required: false },
                    { key: 'note', label: 'Note', type: 'text', required: false },
                  ].map(({ key, label, type, required }) => (
                    <div key={key}>
                      <label style={{ display: 'block', color: 'var(--text-dim)', fontSize: '0.8rem', marginBottom: 4, fontFamily: 'var(--font-mono)' }}>
                        {label}{required && ' *'}
                      </label>
                      <input
                        type={type}
                        value={String(addForm[key as keyof typeof addForm] ?? '')}
                        onChange={(e) =>
                          setAddForm((prev) => ({
                            ...prev,
                            [key]: type === 'number' ? parseFloat(e.target.value) || 0 : e.target.value,
                          }))
                        }
                        style={{
                          width: '100%',
                          padding: '8px 12px',
                          background: 'var(--surface-raised)',
                          border: '1px solid var(--border)',
                          borderRadius: 'var(--radius-sm)',
                          color: 'var(--text)',
                          fontFamily: 'var(--font-mono)',
                          fontSize: '0.9rem',
                          boxSizing: 'border-box',
                        }}
                      />
                    </div>
                  ))}
                </div>
                <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem', justifyContent: 'flex-end' }}>
                  <button
                    onClick={() => setShowAddForm(false)}
                    style={{
                      padding: '8px 16px',
                      borderRadius: 'var(--radius-sm)',
                      border: '1px solid var(--border)',
                      background: 'transparent',
                      color: 'var(--text-dim)',
                      cursor: 'pointer',
                      fontFamily: 'var(--font-display)',
                      fontSize: '0.8rem',
                    }}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleAddItem}
                    disabled={createPending || !addForm.name || !addForm.merchant || !addForm.price}
                    style={{
                      padding: '8px 20px',
                      borderRadius: 'var(--radius-sm)',
                      border: '1px solid var(--cyan)',
                      background: 'var(--cyan-dim)',
                      color: 'var(--cyan)',
                      cursor: createPending ? 'not-allowed' : 'pointer',
                      fontFamily: 'var(--font-display)',
                      fontSize: '0.8rem',
                      fontWeight: 700,
                      opacity: createPending ? 0.6 : 1,
                    }}
                  >
                    {createPending ? 'Adding...' : 'Add Item'}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Content */}
          {items.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '5rem 2rem', color: 'var(--text-dim)', animation: 'fade-in 0.5s ease both' }}>
              <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.65rem', color: 'var(--border-glow)', letterSpacing: '0.15em', marginBottom: '1.5rem' }}>
                {'[ BUDGET TRACKING INACTIVE ]'}
              </div>
              <p style={{ fontSize: '0.95rem', marginBottom: '0.5rem', color: 'var(--text-mid)', fontFamily: 'var(--font-display)', letterSpacing: '0.08em' }}>
                NO ITEMS TRACKED
              </p>
              <p style={{ fontSize: '0.8rem', fontFamily: 'var(--font-mono)', color: 'var(--text-dim)' }}>
                Run Price Intel or add items manually
              </p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              {/* Shopping list */}
              <ShoppingList items={items} onPriceClick={setHistoryItem} />

              {/* Cost breakdown */}
              {items.length >= 2 && (
                <div
                  style={{
                    background: 'var(--surface)',
                    borderRadius: 'var(--radius)',
                    border: '1px solid var(--border)',
                    padding: '1.5rem',
                  }}
                >
                  <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--text-mid)', fontSize: '0.8rem', letterSpacing: '0.1em', marginBottom: '1rem' }}>
                    COST BREAKDOWN
                  </h3>
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
    </div>
  )
}

function PriceHistoryModal({ item, currency, onClose }: { item: ShoppingItem; currency: string; onClose: () => void }) {
  const { snapshots, isLoading } = usePriceSnapshots(item.missionId, item.id)

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(5,8,16,0.85)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 200,
        padding: '1rem',
      }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        style={{
          background: 'var(--surface)',
          borderRadius: 'var(--radius)',
          border: '1px solid var(--border)',
          padding: '2rem',
          width: '100%',
          maxWidth: 600,
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <div>
            <h3 style={{ fontFamily: 'var(--font-display)', color: 'var(--cyan)', letterSpacing: '0.1em' }}>
              PRICE HISTORY
            </h3>
            <p style={{ color: 'var(--text-dim)', fontSize: '0.85rem', marginTop: 4 }}>{item.name} @ {item.merchant}</p>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'transparent',
              border: 'none',
              color: 'var(--text-dim)',
              cursor: 'pointer',
              fontSize: '1.2rem',
              padding: 4,
            }}
          >
            ✕
          </button>
        </div>
        {isLoading ? (
          <LoadingPulse label="Loading history..." />
        ) : snapshots.length === 0 ? (
          <p style={{ color: 'var(--text-dim)', textAlign: 'center', padding: '2rem' }}>No price history yet</p>
        ) : (
          <PriceHistoryChart snapshots={snapshots} currency={currency} />
        )}
      </div>
    </div>
  )
}
