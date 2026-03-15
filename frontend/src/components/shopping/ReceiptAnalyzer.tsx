import { useState } from 'react'
import { useReceipts, useAnalyzeReceipt } from '../../hooks/useReceipt'
import type { ReceiptAnalysis } from '../../api/receipt'
import styles from './ReceiptAnalyzer.module.css'

interface Props {
  missionSlug: string
}

const CURRENCIES = ['EUR', 'USD', 'GBP']

function formatCurrency(value: number, currency: string): string {
  return value.toLocaleString('fr-FR', { style: 'currency', currency })
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString('fr-FR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function ItemsTable({ analysis }: { analysis: ReceiptAnalysis }) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th>Produit</th>
          <th className={styles.numericCell}>Qté</th>
          <th className={styles.numericCell}>Prix unit.</th>
          <th className={styles.numericCell}>Total</th>
        </tr>
      </thead>
      <tbody>
        {analysis.items.map((item, i) => (
          <tr key={i}>
            <td>{item.name}</td>
            <td className={styles.numericCell}>
              {item.quantity.toLocaleString('fr-FR')}
            </td>
            <td className={styles.numericCell}>
              {formatCurrency(item.unitPrice, analysis.currency)}
            </td>
            <td className={styles.numericCell}>
              {formatCurrency(item.total, analysis.currency)}
            </td>
          </tr>
        ))}
        <tr className={styles.totalRow}>
          <td colSpan={3}>Total</td>
          <td className={styles.numericCell}>
            {formatCurrency(analysis.total, analysis.currency)}
          </td>
        </tr>
      </tbody>
    </table>
  )
}

function HistoryCard({ analysis }: { analysis: ReceiptAnalysis }) {
  const [open, setOpen] = useState(false)

  return (
    <button
      className={styles.historyCard}
      aria-expanded={open}
      onClick={() => setOpen((v) => !v)}
      type="button"
    >
      <div className={styles.historyCardHeader}>
        <span className={styles.historyDate}>{formatDate(analysis.analyzedAt)}</span>
        <span className={styles.historyTotal}>
          {formatCurrency(analysis.total, analysis.currency)}
        </span>
        <span className={styles.historyChevron} aria-hidden="true">
          {open ? '▲' : '▼'}
        </span>
      </div>
      {open && (
        <div className={styles.historyBody}>
          <ItemsTable analysis={analysis} />
        </div>
      )}
    </button>
  )
}

export function ReceiptAnalyzer({ missionSlug }: Props) {
  const [rawText, setRawText] = useState('')
  const [currency, setCurrency] = useState('EUR')
  const [lastResult, setLastResult] = useState<ReceiptAnalysis | null>(null)

  const { data: history = [] } = useReceipts(missionSlug)
  const { mutateAsync: analyze, isPending } = useAnalyzeReceipt(missionSlug)

  async function handleAnalyze() {
    if (!rawText.trim()) return
    try {
      const result = await analyze({ rawText, currency })
      setLastResult(result)
      setRawText('')
    } catch {
      // errors surface via react-query; no silent swallow
    }
  }

  return (
    <section className={styles.container} aria-label="Analyser un ticket de caisse">
      <h3 className={styles.title}>TICKET DE CAISSE</h3>

      <div className={styles.form}>
        <textarea
          className={styles.textarea}
          placeholder="Collez ici le texte brut de votre ticket (e-mail, OCR, copier-coller…)"
          value={rawText}
          onChange={(e) => setRawText(e.target.value)}
          aria-label="Texte brut du ticket de caisse"
        />
        <div className={styles.controls}>
          <label className={styles.currencyLabel} htmlFor="receipt-currency">
            Devise
          </label>
          <select
            id="receipt-currency"
            className={styles.currencySelect}
            value={currency}
            onChange={(e) => setCurrency(e.target.value)}
          >
            {CURRENCIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
          <button
            type="button"
            className={styles.analyzeBtn}
            onClick={handleAnalyze}
            disabled={isPending || !rawText.trim()}
          >
            {isPending && <span className={styles.spinner} aria-hidden="true" />}
            {isPending ? 'Analyse…' : 'Analyser'}
          </button>
        </div>
      </div>

      {lastResult && lastResult.items.length > 0 && (
        <div className={styles.resultsSection}>
          <p className={styles.resultsTitle}>Résultats</p>
          <ItemsTable analysis={lastResult} />
        </div>
      )}

      {history.length > 0 && (
        <div className={styles.historySection}>
          <p className={styles.historyTitle}>Historique</p>
          <div className={styles.historyList}>
            {history.map((a) => (
              <HistoryCard key={a.id} analysis={a} />
            ))}
          </div>
        </div>
      )}
    </section>
  )
}
