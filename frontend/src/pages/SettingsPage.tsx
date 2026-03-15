import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSettings, useUpdateSettings, useDeleteAllData } from '../hooks'
import { CurrencyConverter } from '../components/scouter'
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
  const { t } = useTranslation()
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
        <h1 className={styles.heading}>{t('settings.title')}</h1>
        <div className={styles.loading}>{t('common.loading')}</div>
      </main>
    )
  }

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>{t('settings.title')}</h1>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>{t('settings.preferences')}</h2>
        <div className={styles.card}>
          <div className={styles.field}>
            <label className={styles.label}>{t('settings.defaultCurrency')}</label>
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
            <label className={styles.label}>{t('settings.locale')}</label>
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
            <label className={styles.label}>{t('settings.llmProvider')}</label>
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
            <p className={styles.success}>{t('settings.saved')}</p>
          )}
        </div>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Outils</h2>
        <div className={styles.card}>
          <CurrencyConverter />
        </div>
      </section>

      <section className={styles.section}>
        <h2 className={styles.dangerTitle}>{t('settings.dangerZone')}</h2>
        <div className={`${styles.card} ${styles.dangerCard}`}>
          <div className={styles.dangerRow}>
            <div>
              <p className={styles.dangerLabel}>{t('settings.deleteAllData')}</p>
              <p className={styles.dangerDesc}>
                {t('settings.deleteAllDesc')}
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
                ? t('settings.deleting')
                : deleteConfirm
                  ? t('settings.confirmDelete')
                  : t('settings.deleteAll')}
            </button>
          </div>
          {deleteSuccess && (
            <p className={styles.success}>{t('settings.allDeleted')}</p>
          )}
          {deleteConfirm && !deleteData.isPending && (
            <p className={styles.dangerWarning}>
              ⚠️ {t('settings.confirmWarning')}
            </p>
          )}
        </div>
      </section>
    </main>
  )
}
