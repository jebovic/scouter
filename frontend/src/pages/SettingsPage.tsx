import { useState } from 'react'
import { useSettings, useUpdateSettings, useDeleteAllData } from '../hooks'
import styles from './SettingsPage.module.css'

const CURRENCIES = ['EUR', 'USD', 'GBP', 'CHF', 'JPY', 'CAD']
const LOCALES = [
  { value: 'fr-FR', label: 'Français (France)' },
  { value: 'en-US', label: 'English (US)' },
  { value: 'en-GB', label: 'English (UK)' },
  { value: 'de-DE', label: 'Deutsch' },
]
const LLM_PROVIDERS = [
  { value: 'ollama', label: 'Ollama (Local)' },
  { value: 'anthropic', label: 'Anthropic Claude' },
]

export default function SettingsPage() {
  const { data: settings, isLoading } = useSettings()
  const updateSettings = useUpdateSettings()
  const deleteData = useDeleteAllData()
  const [deleteConfirm, setDeleteConfirm] = useState(false)
  const [deleteSuccess, setDeleteSuccess] = useState(false)

  function handleChange(key: string, value: string) {
    updateSettings.mutate({ [key]: value })
  }

  function handleDeleteAll() {
    if (!deleteConfirm) {
      setDeleteConfirm(true)
      return
    }
    deleteData.mutate(undefined, {
      onSuccess: () => {
        setDeleteSuccess(true)
        setDeleteConfirm(false)
      },
    })
  }

  if (isLoading) {
    return (
      <main className={styles.page}>
        <h1 className={styles.heading}>Settings</h1>
        <div className={styles.loading}>Loading...</div>
      </main>
    )
  }

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>Settings</h1>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Preferences</h2>
        <div className={styles.card}>
          <div className={styles.field}>
            <label className={styles.label}>Default Currency</label>
            <select
              className={styles.select}
              value={settings?.currency ?? 'EUR'}
              onChange={(e) => handleChange('currency', e.target.value)}
            >
              {CURRENCIES.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Locale</label>
            <select
              className={styles.select}
              value={settings?.locale ?? 'fr-FR'}
              onChange={(e) => handleChange('locale', e.target.value)}
            >
              {LOCALES.map((l) => (
                <option key={l.value} value={l.value}>
                  {l.label}
                </option>
              ))}
            </select>
          </div>
          <div className={styles.field}>
            <label className={styles.label}>LLM Provider</label>
            <select
              className={styles.select}
              value={settings?.llm_provider ?? 'ollama'}
              onChange={(e) => handleChange('llm_provider', e.target.value)}
            >
              {LLM_PROVIDERS.map((p) => (
                <option key={p.value} value={p.value}>
                  {p.label}
                </option>
              ))}
            </select>
          </div>
          {updateSettings.isSuccess && (
            <p className={styles.success}>Settings saved.</p>
          )}
        </div>
      </section>

      <section className={styles.section}>
        <h2 className={styles.dangerTitle}>Danger Zone</h2>
        <div className={`${styles.card} ${styles.dangerCard}`}>
          <div className={styles.dangerRow}>
            <div>
              <p className={styles.dangerLabel}>Delete All Data</p>
              <p className={styles.dangerDesc}>
                Permanently delete all missions, options, shopping items, and
                purchase records. This cannot be undone.
              </p>
            </div>
            <button
              className={`${styles.dangerBtn} ${
                deleteConfirm ? styles.dangerBtnConfirm : ''
              }`}
              onClick={handleDeleteAll}
              disabled={deleteData.isPending}
            >
              {deleteData.isPending
                ? 'Deleting...'
                : deleteConfirm
                  ? 'Confirm Delete All'
                  : 'Delete All Data'}
            </button>
          </div>
          {deleteSuccess && (
            <p className={styles.success}>All data deleted.</p>
          )}
          {deleteConfirm && !deleteData.isPending && (
            <p className={styles.dangerWarning}>
              ⚠️ Click again to confirm. This will permanently delete everything.
            </p>
          )}
        </div>
      </section>
    </main>
  )
}
