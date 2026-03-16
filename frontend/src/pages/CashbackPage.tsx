import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Topnav } from '../components/scouter/Topnav'
import {
  useCashbackEntries,
  useCashbackSummary,
  useCreateCashback,
  useDeleteCashback,
} from '../hooks/useCashback'
import { useFormatCurrency } from '../hooks/useFormatCurrency'
import styles from './CashbackPage.module.css'

const COLLABORATOR_KEY = 'scouter_collaborator_id'

function getUserRef(): string {
  if (typeof window === 'undefined') return ''
  let ref = localStorage.getItem(COLLABORATOR_KEY)
  if (!ref) {
    ref = crypto.randomUUID()
    localStorage.setItem(COLLABORATOR_KEY, ref)
  }
  return ref
}

const STATUS_OPTIONS = ['pending', 'confirmed', 'paid'] as const
type Status = (typeof STATUS_OPTIONS)[number]

export default function CashbackPage() {
  const { t } = useTranslation()
  const userRef = getUserRef()
  const { fmt } = useFormatCurrency()

  const { data: entries = [], isLoading: loadingEntries } = useCashbackEntries(userRef)
  const { data: summary } = useCashbackSummary(userRef)
  const createMutation = useCreateCashback(userRef)
  const deleteMutation = useDeleteCashback(userRef)

  const [form, setForm] = useState({
    itemName: '',
    merchant: '',
    amount: '',
    percentRate: '',
    status: 'pending' as Status,
    notes: '',
  })

  function handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
    const { name, value } = e.target
    setForm((prev) => ({ ...prev, [name]: value }))
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.itemName.trim() || !form.merchant.trim()) return
    createMutation.mutate(
      {
        itemName: form.itemName.trim(),
        merchant: form.merchant.trim(),
        amount: parseFloat(form.amount) || 0,
        percentRate: parseFloat(form.percentRate) || 0,
        status: form.status,
        notes: form.notes.trim(),
      },
      {
        onSuccess: () =>
          setForm({ itemName: '', merchant: '', amount: '', percentRate: '', status: 'pending', notes: '' }),
      },
    )
  }

  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
        <div className={styles.header}>
          <h1 className={styles.title}>{t('cashback.title')}</h1>
          <p className={styles.subtitle}>{t('cashback.subtitle')}</p>
        </div>

        {/* Summary cards */}
        <div className={styles.summaryGrid}>
          <div className={styles.summaryCard}>
            <span className={styles.summaryLabel}>{t('cashback.pending')}</span>
            <span className={`${styles.summaryAmount} ${styles.pending}`}>
              {summary ? fmt(summary.totalPending) : '—'}
            </span>
          </div>
          <div className={styles.summaryCard}>
            <span className={styles.summaryLabel}>{t('cashback.confirmed')}</span>
            <span className={`${styles.summaryAmount} ${styles.confirmed}`}>
              {summary ? fmt(summary.totalConfirmed) : '—'}
            </span>
          </div>
          <div className={styles.summaryCard}>
            <span className={styles.summaryLabel}>{t('cashback.received')}</span>
            <span className={`${styles.summaryAmount} ${styles.paid}`}>
              {summary ? fmt(summary.totalPaid) : '—'}
            </span>
          </div>
        </div>

        {/* Form */}
        <div className={styles.formCard}>
          <p className={styles.formTitle}>{t('cashback.addTitle')}</p>
          <form onSubmit={handleSubmit}>
            <div className={styles.formGrid}>
              <div className={styles.formGroup}>
                <label className={styles.label} htmlFor="itemName">{t('cashback.item')}</label>
                <input
                  id="itemName"
                  name="itemName"
                  className={styles.input}
                  value={form.itemName}
                  onChange={handleChange}
                  placeholder={t('cashback.itemPlaceholder')}
                  required
                />
              </div>
              <div className={styles.formGroup}>
                <label className={styles.label} htmlFor="merchant">{t('cashback.merchant')}</label>
                <input
                  id="merchant"
                  name="merchant"
                  className={styles.input}
                  value={form.merchant}
                  onChange={handleChange}
                  placeholder={t('cashback.merchantPlaceholder')}
                  required
                />
              </div>
              <div className={styles.formGroup}>
                <label className={styles.label} htmlFor="amount">{t('cashback.amount')}</label>
                <input
                  id="amount"
                  name="amount"
                  type="number"
                  min="0"
                  step="0.01"
                  className={styles.input}
                  value={form.amount}
                  onChange={handleChange}
                  placeholder="45.00"
                />
              </div>
              <div className={styles.formGroup}>
                <label className={styles.label} htmlFor="percentRate">{t('cashback.rate')}</label>
                <input
                  id="percentRate"
                  name="percentRate"
                  type="number"
                  min="0"
                  step="0.1"
                  className={styles.input}
                  value={form.percentRate}
                  onChange={handleChange}
                  placeholder="5.0"
                />
              </div>
              <div className={styles.formGroup}>
                <label className={styles.label} htmlFor="status">{t('cashback.status')}</label>
                <select
                  id="status"
                  name="status"
                  className={styles.select}
                  value={form.status}
                  onChange={handleChange}
                >
                  {STATUS_OPTIONS.map((s) => (
                    <option key={s} value={s}>
                      {s === 'pending' ? t('cashback.pending') : s === 'confirmed' ? t('cashback.confirmed') : t('cashback.received')}
                    </option>
                  ))}
                </select>
              </div>
              <div className={styles.formGroup}>
                <label className={styles.label} htmlFor="notes">{t('cashback.notes')}</label>
                <input
                  id="notes"
                  name="notes"
                  className={styles.input}
                  value={form.notes}
                  onChange={handleChange}
                  placeholder={t('cashback.optional')}
                />
              </div>
            </div>
            <button
              type="submit"
              className={styles.submitBtn}
              disabled={createMutation.isPending || !form.itemName.trim() || !form.merchant.trim()}
            >
              {createMutation.isPending ? t('cashback.adding') : t('common.add')}
            </button>
          </form>
        </div>

        {/* Entries */}
        <p className={styles.listHeader}>
          {t('cashback.myCashbacks')} {summary ? `(${summary.entryCount})` : ''}
        </p>

        {loadingEntries ? (
          <div className={styles.empty}>
            <p>{t('common.loading')}</p>
          </div>
        ) : entries.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyIcon}>🐷</div>
            <p className={styles.emptyTitle}>{t('cashback.none')}</p>
            <p className={styles.emptyDesc}>{t('cashback.noneDesc')}</p>
          </div>
        ) : (
          <div className={styles.entryList}>
            {entries.map((entry) => (
              <div key={entry.id} className={styles.entryCard}>
                <div className={styles.entryInfo}>
                  <div className={styles.entryName}>{entry.itemName}</div>
                  <div className={styles.entryMeta}>
                    {entry.merchant}
                    {entry.percentRate > 0 && ` · ${entry.percentRate}%`}
                    {entry.notes && ` · ${entry.notes}`}
                  </div>
                </div>
                <span className={styles.entryAmount}>{fmt(entry.amount)}</span>
                <span className={`${styles.statusPill} ${styles[entry.status]}`}>
                  {entry.status === 'pending' ? t('cashback.pending') : entry.status === 'confirmed' ? t('cashback.confirmed') : t('cashback.received')}
                </span>
                <button
                  className={styles.deleteBtn}
                  onClick={() => deleteMutation.mutate(entry.id)}
                  disabled={deleteMutation.isPending}
                  aria-label={t('cashback.deleteAriaLabel', { name: entry.itemName })}
                >
                  {t('cashback.delete')}
                </button>
              </div>
            ))}
          </div>
        )}
      </main>
    </>
  )
}
