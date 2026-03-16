import { useState, useEffect } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { useSettings, useUpdateSettings, useDeleteAllData } from '../hooks'
import { CurrencyConverter } from '../components/scouter'
import { Topnav } from '../components/scouter/Topnav'
import { syncLanguageFromSettings } from '../i18n'
import { usePerformanceMetrics } from '../hooks/usePerformanceMetrics'
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

const DELETE_CONFIRM_WORD = 'DELETE'

function formatMs(value: number | null): { value: string; unit: string } {
  if (value === null) return { value: '—', unit: '' }
  if (value >= 1000) return { value: (value / 1000).toFixed(2), unit: 's' }
  return { value: Math.round(value).toString(), unit: 'ms' }
}

function formatBytes(value: number | null): { value: string; unit: string } {
  if (value === null) return { value: '—', unit: '' }
  const mb = value / (1024 * 1024)
  return { value: mb.toFixed(1), unit: 'MB' }
}

function msColorClass(value: number | null, greenThreshold: number, goldThreshold: number): string {
  if (value === null) return styles.perfCardNeutral
  if (value <= greenThreshold) return styles.perfCardGreen
  if (value <= goldThreshold) return styles.perfCardGold
  return styles.perfCardCoral
}

function msValueClass(value: number | null, greenThreshold: number, goldThreshold: number): string {
  if (value === null) return ''
  if (value <= greenThreshold) return styles.perfMetricValueGreen
  if (value <= goldThreshold) return styles.perfMetricValueGold
  return styles.perfMetricValueCoral
}

function PerformancePanel() {
  const { t } = useTranslation()
  const { metrics, refresh } = usePerformanceMetrics()

  const metricCards = [
    {
      name: t('settings.perfTTFB'),
      label: t('settings.perfTTFBLabel'),
      formatted: formatMs(metrics.ttfb),
      cardClass: msColorClass(metrics.ttfb, 200, 600),
      valueClass: msValueClass(metrics.ttfb, 200, 600),
    },
    {
      name: t('settings.perfDOMLoaded'),
      label: t('settings.perfDOMLabel'),
      formatted: formatMs(metrics.domContentLoaded),
      cardClass: msColorClass(metrics.domContentLoaded, 1500, 3000),
      valueClass: msValueClass(metrics.domContentLoaded, 1500, 3000),
    },
    {
      name: t('settings.perfLoadTime'),
      label: t('settings.perfLoadLabel'),
      formatted: formatMs(metrics.loadTime),
      cardClass: msColorClass(metrics.loadTime, 2500, 5000),
      valueClass: msValueClass(metrics.loadTime, 2500, 5000),
    },
    {
      name: t('settings.perfHeapUsed'),
      label: t('settings.perfHeapLabel'),
      formatted: formatBytes(metrics.usedJSHeapSize),
      cardClass: styles.perfCardNeutral,
      valueClass: '',
    },
    {
      name: t('settings.perfConnection'),
      label: t('settings.perfConnectionLabel'),
      formatted: { value: metrics.connectionType ?? '—', unit: '' },
      cardClass: styles.perfCardNeutral,
      valueClass: '',
    },
    {
      name: t('settings.perfTotalHeap'),
      label: t('settings.perfTotalHeapLabel'),
      formatted: formatBytes(metrics.totalJSHeapSize),
      cardClass: styles.perfCardNeutral,
      valueClass: '',
    },
  ]

  return (
    <details>
      <summary className={styles.perfSummary}>
        {t('settings.performance')}
      </summary>
      <div className={styles.perfBody}>
        <div className={styles.perfHeader}>
          <p className={styles.perfDesc}>{t('settings.performanceDesc')}</p>
          <button className={styles.perfRefreshBtn} onClick={refresh} type="button">
            {t('settings.perfRefresh')}
          </button>
        </div>
        <div className={styles.perfGrid}>
          {metricCards.map((card) => (
            <div key={card.name} className={`${styles.perfCard} ${card.cardClass}`}>
              <p className={styles.perfMetricName}>{card.name}</p>
              <p className={`${styles.perfMetricValue} ${card.valueClass}`}>
                {card.formatted.value}
                {card.formatted.unit && (
                  <span className={styles.perfMetricUnit}> {card.formatted.unit}</span>
                )}
              </p>
              <p className={styles.perfMetricLabel}>{card.label}</p>
            </div>
          ))}
        </div>
        <p className={styles.perfFooter}>{t('settings.perfDisclaimer')} — {t('settings.perfNote')}</p>
      </div>
    </details>
  )
}

export default function SettingsPage() {
  const { t } = useTranslation()
  const { data: settings, isLoading } = useSettings()
  const updateSettings = useUpdateSettings()
  const deleteData = useDeleteAllData()
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deleteInput, setDeleteInput] = useState('')
  const [deleteSuccess, setDeleteSuccess] = useState(false)
  const [simpleMode, setSimpleMode] = useState(false)

  useEffect(() => {
    try { setSimpleMode(localStorage.getItem('simpleMode') === 'true') } catch {}
  }, [])

  function toggleSimpleMode() {
    const next = !simpleMode
    setSimpleMode(next)
    try { localStorage.setItem('simpleMode', String(next)) } catch {}
    // Notify NavRail via storage event
    window.dispatchEvent(new StorageEvent('storage', { key: 'simpleMode', newValue: String(next) }))
  }

  function handleChange(key: string, value: string) {
    updateSettings.mutate({ [key]: value })
    if (key === 'locale') {
      syncLanguageFromSettings(value)
    }
  }

  function handleDeleteConfirm() {
    deleteData.mutate(undefined, {
      onSuccess: () => {
        setDeleteSuccess(true)
        setShowDeleteModal(false)
        setDeleteInput('')
      },
    })
  }

  if (isLoading) {
    return (
      <>
        <Topnav />
        <main className={`page ${styles.page}`}>
          <h1 className={styles.heading}>{t('settings.title')}</h1>
          <div className={styles.loading}>{t('common.loading')}</div>
        </main>
      </>
    )
  }

  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
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
          <div className={styles.field}>
            <label className={styles.label}>{t('settings.simpleMode')}</label>
            <div className={styles.toggleRow}>
              <span className={styles.toggleDesc}>
                {t('settings.simpleModeDesc')}
              </span>
              <button
                className={styles.toggle}
                data-active={simpleMode}
                onClick={toggleSimpleMode}
                aria-pressed={simpleMode}
              >
                {simpleMode ? t('settings.enabled') : t('settings.disabled')}
              </button>
            </div>
          </div>
          {updateSettings.isSuccess && (
            <p className={styles.success}>{t('settings.saved')}</p>
          )}
        </div>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>{t('settings.tools')}</h2>
        <CurrencyConverter />
        <div className={`${styles.card} ${styles.toolCard}`}>
          <PerformancePanel />
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
              className={styles.dangerBtn}
              onClick={() => { setShowDeleteModal(true); setDeleteInput('') }}
            >
              {t('settings.deleteAll')}
            </button>
          </div>
          {deleteSuccess && (
            <p className={styles.success}>{t('settings.allDeleted')}</p>
          )}
        </div>
      </section>

      {showDeleteModal && (
        <div className={styles.modalBackdrop} role="dialog" aria-modal="true" aria-label={t('settings.deleteAllData')}>
          <div className={styles.modal}>
            <h3 className={styles.modalTitle}>⚠️ {t('settings.dangerZone')}</h3>
            <p className={styles.dangerWarning}>{t('settings.confirmWarning')}</p>
            <p className={styles.modalPrompt}>
              <Trans i18nKey="settings.deletePrompt" values={{ word: DELETE_CONFIRM_WORD }} components={{ bold: <strong /> }} />
            </p>
            <input
              className={styles.modalInput}
              type="text"
              value={deleteInput}
              onChange={(e) => setDeleteInput(e.target.value)}
              autoFocus
              aria-label={t('settings.deletePrompt', { word: DELETE_CONFIRM_WORD })}
            />
            <div className={styles.modalActions}>
              <button
                className={styles.modalCancelBtn}
                onClick={() => { setShowDeleteModal(false); setDeleteInput('') }}
              >
                {t('common.cancel')}
              </button>
              <button
                className={styles.dangerBtnConfirm}
                onClick={handleDeleteConfirm}
                disabled={deleteInput !== DELETE_CONFIRM_WORD || deleteData.isPending}
              >
                {deleteData.isPending ? t('settings.deleting') : t('settings.confirmDelete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
    </>
  )
}
