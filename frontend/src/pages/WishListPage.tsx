import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LoadingPulse } from '../components/scouter'
import { WishListItem } from '../components/wishlist'
import { useWishList, useCreateWishListItem, useDeleteWishListItem } from '../hooks/useWishList'
import { lookupProduct } from '../api/product'
import type { ProductInfo } from '../api/product'
import styles from './WishListPage.module.css'

const CURRENCIES = ['EUR', 'USD', 'GBP'] as const

interface AddFormState {
  name: string
  url: string
  targetPrice: string
  currency: string
  notes: string
}

const emptyForm: AddFormState = {
  name: '',
  url: '',
  targetPrice: '',
  currency: 'EUR',
  notes: '',
}

export default function WishListPage() {
  const { t } = useTranslation()
  const { items, isLoading } = useWishList()
  const createMutation = useCreateWishListItem()
  const deleteMutation = useDeleteWishListItem()

  const [form, setForm] = useState<AddFormState>(emptyForm)
  const [formError, setFormError] = useState<string | null>(null)

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
      currency: form.currency || 'EUR',
    }
    if (form.url.trim()) payload.url = form.url.trim()
    if (form.targetPrice.trim()) {
      const parsed = parseFloat(form.targetPrice)
      if (!isNaN(parsed)) payload.targetPrice = parsed
    }
    if (form.notes.trim()) payload.notes = form.notes.trim()

    try {
      await createMutation.mutateAsync(payload)
      setForm(emptyForm)
      setFormError(null)
    } catch {
      setFormError(t('common.error'))
    }
  }

  function handleDelete(id: string) {
    deleteMutation.mutate(id)
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
    <main className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>{t('wishlist.title')}</h1>
        <p className={styles.subtitle}>{t('wishlist.subtitle')}</p>
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
  )
}
