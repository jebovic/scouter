import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LoadingPulse } from '../components/scouter'
import { WishListItem } from '../components/wishlist'
import { useWishList, useCreateWishListItem, useDeleteWishListItem } from '../hooks/useWishList'
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

  return (
    <main className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>{t('wishlist.title')}</h1>
        <p className={styles.subtitle}>{t('wishlist.subtitle')}</p>
      </div>

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
