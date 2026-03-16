import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useFormatCurrency } from '../hooks/useFormatCurrency'
import { LoadingPulse } from '../components/scouter'
import { Topnav } from '../components/scouter/Topnav'
import { WishListItem } from '../components/wishlist'
import { useWishList, useCreateWishListItem, useDeleteWishListItem } from '../hooks/useWishList'
import { lookupProduct } from '../api/product'
import type { ProductInfo } from '../api/product'
import { copyShareLink } from '../utils/wishlistShare'
import { useWishlistPrioritizer } from '../hooks/useWishlistPrioritizer'
import styles from './WishListPage.module.css'

const CURRENCIES = ['EUR', 'USD', 'GBP'] as const

interface AddFormState {
  name: string
  url: string
  targetPrice: string
  currency: string
  notes: string
}

function makeEmptyForm(defaultCurrency: string): AddFormState {
  return {
    name: '',
    url: '',
    targetPrice: '',
    currency: defaultCurrency,
    notes: '',
  }
}

function WishlistPriorityCard() {
  const { t } = useTranslation()
  const { data, isLoading, isError } = useWishlistPrioritizer()
  if (isLoading) return <div style={{ padding: '12px', color: 'var(--text-dim)', fontSize: '0.85rem' }}>{t('wishlist.priorityLoading')}</div>
  if (isError || !data || data.items.length === 0) return null
  return (
    <div style={{
      background: 'var(--surface)',
      border: '1px solid var(--border)',
      borderRadius: '12px',
      padding: '16px',
      marginBottom: '24px',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
        <span style={{ fontSize: '1.1rem' }}>🏆</span>
        <span style={{ fontWeight: 600, color: 'var(--text)' }}>{t('wishlist.priorityTitle')}</span>
        {data.topPick && (
          <span style={{
            marginLeft: 'auto',
            fontSize: '0.75rem',
            background: 'var(--accent)',
            color: '#fff',
            borderRadius: '6px',
            padding: '2px 8px',
          }}>{t('wishlist.topPick', { name: data.topPick.name })}</span>
        )}
      </div>
      <p style={{ fontSize: '0.82rem', color: 'var(--text-dim)', marginBottom: '12px' }}>{data.summary}</p>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
        {data.items.slice(0, 5).map((item) => (
          <div key={item.id} style={{
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            padding: '8px 10px',
            background: 'var(--surface-alt)',
            borderRadius: '8px',
          }}>
            <span style={{
              minWidth: '22px',
              height: '22px',
              borderRadius: '50%',
              background: item.rank === 1 ? 'var(--accent)' : 'var(--border)',
              color: item.rank === 1 ? '#fff' : 'var(--text-dim)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '0.75rem',
              fontWeight: 700,
            }}>{item.rank}</span>
            <span style={{ flex: 1, fontSize: '0.85rem', color: 'var(--text)', fontWeight: 500 }}>{item.name}</span>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>{item.verdict}</span>
            <span style={{
              fontSize: '0.75rem',
              fontWeight: 700,
              color: item.score >= 75 ? 'var(--status-buy)' : item.score >= 50 ? 'var(--status-watch)' : 'var(--text-dim)',
            }}>{Math.round(item.score)}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function WishListPage() {
  const { t } = useTranslation()
  const { currency } = useFormatCurrency()
  const { items, isLoading } = useWishList()
  const createMutation = useCreateWishListItem()
  const deleteMutation = useDeleteWishListItem()

  const [form, setForm] = useState<AddFormState>(() => makeEmptyForm(currency))
  const [formError, setFormError] = useState<string | null>(null)
  const [shareCopied, setShareCopied] = useState(false)

  // Barcode lookup state
  const [barcode, setBarcode] = useState('')
  const [barcodeLoading, setBarcodeLoading] = useState(false)
  const [barcodeError, setBarcodeError] = useState<string | null>(null)
  const [foundProduct, setFoundProduct] = useState<ProductInfo | null>(null)

  function handleFieldChange(field: keyof AddFormState, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }))
    if (field === 'name') setFormError(null)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmedName = form.name.trim()
    if (!trimmedName) {
      setFormError(t('wishlist.name'))
      return
    }

    const payload: Parameters<typeof createMutation.mutateAsync>[0] = {
      name: trimmedName,
      currency: form.currency || currency,
    }
    if (form.url.trim()) payload.url = form.url.trim()
    if (form.targetPrice.trim()) {
      const parsed = parseFloat(form.targetPrice)
      if (!isNaN(parsed)) payload.targetPrice = parsed
    }
    if (form.notes.trim()) payload.notes = form.notes.trim()

    try {
      await createMutation.mutateAsync(payload)
      setForm(makeEmptyForm(currency))
      setFormError(null)
    } catch {
      setFormError(t('common.error'))
    }
  }

  function handleDelete(id: string) {
    deleteMutation.mutate(id)
  }

  async function handleShare() {
    try {
      await copyShareLink(items)
      setShareCopied(true)
      setTimeout(() => setShareCopied(false), 2000)
    } catch {
      // clipboard not available — silently ignore
    }
  }

  async function handleBarcodeLookup() {
    const trimmed = barcode.trim()
    if (!trimmed) return
    setBarcodeLoading(true)
    setBarcodeError(null)
    setFoundProduct(null)
    try {
      const product = await lookupProduct(trimmed)
      setFoundProduct(product)
    } catch (err) {
      if (err instanceof Error && err.message === 'NOT_FOUND') {
        setBarcodeError(t('product.notFound'))
      } else {
        setBarcodeError(t('common.error'))
      }
    } finally {
      setBarcodeLoading(false)
    }
  }

  function handleBarcodeKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      e.preventDefault()
      void handleBarcodeLookup()
    }
  }

  function handlePrefillForm(product: ProductInfo) {
    setForm((prev) => ({ ...prev, name: product.name }))
    setFoundProduct(null)
    setBarcode('')
  }

  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
      <div className={styles.header}>
        <div className={styles.headerRow}>
          <div>
            <h1 className={styles.title}>{t('wishlist.title')}</h1>
            <p className={styles.subtitle}>{t('wishlist.subtitle')}</p>
          </div>
          {items.length > 0 && (
            <button
              type="button"
              className={styles.shareBtn}
              onClick={() => void handleShare()}
            >
              {shareCopied ? t('wishlist.shareLinkCopied') : t('wishlist.shareMyList')}
            </button>
          )}
        </div>
      </div>

      {/* Barcode lookup */}
      <section className={styles.barcodeSection}>
        <h2 className={styles.sectionTitle}>{t('product.barcodeTitle')}</h2>
        <div className={styles.barcodeRow}>
          <input
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            className={styles.input}
            placeholder={t('product.barcodeInput')}
            value={barcode}
            onChange={(e) => {
              setBarcode(e.target.value)
              setBarcodeError(null)
              setFoundProduct(null)
            }}
            onKeyDown={handleBarcodeKeyDown}
            aria-label={t('product.barcodeInput')}
          />
          <button
            type="button"
            className={styles.lookupBtn}
            onClick={() => void handleBarcodeLookup()}
            disabled={barcodeLoading || !barcode.trim()}
          >
            {barcodeLoading ? t('product.looking') : t('product.lookup')}
          </button>
        </div>

        {barcodeError && (
          <p className={styles.error}>{barcodeError}</p>
        )}

        {foundProduct && (
          <div className={styles.productCard}>
            {foundProduct.imageUrl && (
              <img
                src={foundProduct.imageUrl}
                alt={foundProduct.name}
                className={styles.productImage}
              />
            )}
            <div className={styles.productInfo}>
              <p className={styles.productName}>{foundProduct.name}</p>
              {foundProduct.brand && (
                <p className={styles.productMeta}>{foundProduct.brand}</p>
              )}
              {foundProduct.quantity && (
                <p className={styles.productMeta}>{foundProduct.quantity}</p>
              )}
              {foundProduct.nutriScore && (
                <p className={styles.productMeta}>
                  {t('product.nutriScore')}: {foundProduct.nutriScore.toUpperCase()}
                </p>
              )}
              {foundProduct.ecoScore && (
                <p className={styles.productMeta}>
                  {t('product.ecoScore')}: {foundProduct.ecoScore.toUpperCase()}
                </p>
              )}
              {foundProduct.stores.length > 0 && (
                <p className={styles.productMeta}>
                  {t('product.stores')}: {foundProduct.stores.join(', ')}
                </p>
              )}
              <button
                type="button"
                className={styles.prefillBtn}
                onClick={() => handlePrefillForm(foundProduct)}
              >
                {t('product.addToWishList')}
              </button>
            </div>
          </div>
        )}
      </section>

      {/* Add form */}
      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        <div className={styles.formGrid}>
          <div className={styles.fieldGroup}>
            <label htmlFor="wl-name" className={styles.label}>
              {t('wishlist.name')} <span className={styles.required}>*</span>
            </label>
            <input
              id="wl-name"
              type="text"
              className={styles.input}
              placeholder={t('wishlist.name')}
              value={form.name}
              onChange={(e) => handleFieldChange('name', e.target.value)}
              required
            />
          </div>

          <div className={styles.fieldGroup}>
            <label htmlFor="wl-url" className={styles.label}>{t('wishlist.url')}</label>
            <input
              id="wl-url"
              type="url"
              className={styles.input}
              placeholder={t('wishlist.url')}
              value={form.url}
              onChange={(e) => handleFieldChange('url', e.target.value)}
            />
          </div>

          <div className={styles.fieldGroup}>
            <label htmlFor="wl-price" className={styles.label}>{t('wishlist.targetPrice')}</label>
            <input
              id="wl-price"
              type="number"
              min="0"
              step="0.01"
              className={styles.input}
              placeholder={t('wishlist.targetPrice')}
              value={form.targetPrice}
              onChange={(e) => handleFieldChange('targetPrice', e.target.value)}
            />
          </div>

          <div className={styles.fieldGroup}>
            <label htmlFor="wl-currency" className={styles.label}>{t('wishlist.currency')}</label>
            <select
              id="wl-currency"
              className={styles.select}
              value={form.currency}
              onChange={(e) => handleFieldChange('currency', e.target.value)}
            >
              {CURRENCIES.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>

          <div className={`${styles.fieldGroup} ${styles.notesGroup}`}>
            <label htmlFor="wl-notes" className={styles.label}>{t('wishlist.notes')}</label>
            <input
              id="wl-notes"
              type="text"
              className={styles.input}
              placeholder={t('wishlist.notes')}
              value={form.notes}
              onChange={(e) => handleFieldChange('notes', e.target.value)}
            />
          </div>

          <div className={styles.submitGroup}>
            <button
              type="submit"
              className={styles.addBtn}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? t('common.adding') : t('wishlist.addItem')}
            </button>
          </div>
        </div>
        {formError && <p className={styles.error}>{formError}</p>}
      </form>

      {/* Smart Priority Panel */}
      {items.length > 0 && <WishlistPriorityCard />}

      {/* List */}
      {isLoading ? (
        <LoadingPulse label={t('wishlist.loading')} />
      ) : items.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyIcon}>🎯</span>
          <p className={styles.emptyTitle}>{t('wishlist.empty')}</p>
          <p className={styles.emptyDesc}>{t('wishlist.emptyDesc')}</p>
        </div>
      ) : (
        <div className={styles.list}>
          {items.map((item) => (
            <WishListItem key={item.id} item={item} onDelete={handleDelete} />
          ))}
        </div>
      )}
    </main>
    </>
  )
}
