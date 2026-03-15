import { useState, useRef } from 'react'
import Tesseract from 'tesseract.js'
import styles from './ReceiptScanner.module.css'

// ── Pure parsing helpers (exported for unit tests) ──────────────────────────

/**
 * Extracts the total amount from OCR text.
 * Looks for TOTAL, MONTANT, or TTC keywords followed by a numeric value.
 */
export function extractAmount(text: string): number | undefined {
  if (!text) return undefined

  // Match TOTAL (TTC)?, MONTANT, or standalone TTC followed by optional space and a number.
  // Uses negative lookbehind to exclude SOUS-TOTAL and similar prefixed variants.
  // Handles comma or dot as decimal separator, optional space-thousands separator.
  const pattern = /(?<![\w-])(?:TOTAL\s*(?:TTC)?|MONTANT|(?<!\w)TTC)\s+([\d\s]+[.,]\d{2}|\d+)/gi
  const matches = [...text.matchAll(pattern)]

  for (const match of matches) {
    const raw = match[1].replace(/\s/g, '').replace(',', '.')
    const value = parseFloat(raw)
    if (!isNaN(value)) return value
  }
  return undefined
}

/**
 * Extracts the merchant name — the first non-empty line of OCR text.
 */
export function extractMerchant(text: string): string | undefined {
  if (!text) return undefined
  const lines = text.split('\n')
  for (const line of lines) {
    const trimmed = line.trim()
    if (trimmed.length > 0) return trimmed
  }
  return undefined
}

/**
 * Extracts the first date matching DD/MM/YYYY or DD-MM-YYYY.
 */
export function extractDate(text: string): string | undefined {
  if (!text) return undefined
  const pattern = /\b(\d{2}[\/\-]\d{2}[\/\-]\d{4})\b/
  const match = text.match(pattern)
  return match ? match[1] : undefined
}

// ── Component ────────────────────────────────────────────────────────────────

export interface ReceiptScannerResult {
  merchant?: string
  amount?: number
  date?: string
}

interface Props {
  onResult: (data: ReceiptScannerResult) => void
  onClose: () => void
}

type Phase = 'idle' | 'scanning' | 'preview' | 'error'

interface PreviewData {
  merchant: string
  amount: string
  date: string
}

export function ReceiptScanner({ onResult, onClose }: Props) {
  const [phase, setPhase] = useState<Phase>('idle')
  const [progress, setProgress] = useState(0)
  const [preview, setPreview] = useState<PreviewData>({ merchant: '', amount: '', date: '' })
  const [errorMsg, setErrorMsg] = useState('')
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  async function processFile(file: File) {
    setPhase('scanning')
    setProgress(0)

    try {
      const result = await Tesseract.recognize(file, 'fra+eng', {
        logger: (m: { status: string; progress: number }) => {
          if (m.status === 'recognizing text') {
            setProgress(Math.round(m.progress * 100))
          }
        },
      })
      const text = result.data.text

      const merchant = extractMerchant(text) ?? ''
      const amount = extractAmount(text)
      const date = extractDate(text) ?? ''

      setPreview({
        merchant,
        amount: amount !== undefined ? String(amount) : '',
        date,
      })
      setPhase('preview')
    } catch {
      setErrorMsg('La lecture du reçu a échoué. Veuillez réessayer avec une image plus nette.')
      setPhase('error')
    }
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) processFile(file)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) processFile(file)
  }

  function handleConfirm() {
    onResult({
      merchant: preview.merchant || undefined,
      amount: preview.amount ? parseFloat(preview.amount) : undefined,
      date: preview.date || undefined,
    })
  }

  function handleRetry() {
    setPhase('idle')
    setProgress(0)
    setErrorMsg('')
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  return (
    <div className={styles.overlay}>
      <div className={styles.panel}>
        <div className={styles.header}>
          <h2 className={styles.title}>Scanner un reçu</h2>
          <button
            className={styles.closeBtn}
            onClick={onClose}
            aria-label="Annuler"
          >
            Annuler
          </button>
        </div>

        {phase === 'idle' && (
          <div
            className={`${styles.dropZone} ${dragOver ? styles.dropZoneOver : ''}`}
            onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => e.key === 'Enter' && fileInputRef.current?.click()}
          >
            <span className={styles.dropIcon}>📷</span>
            <p className={styles.dropLabel}>
              Glisser-déposer une photo de reçu ici
            </p>
            <p className={styles.dropHint}>ou cliquer pour sélectionner un fichier</p>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*,application/pdf"
              className={styles.fileInput}
              onChange={handleFileChange}
              aria-label="Sélectionner un reçu"
            />
          </div>
        )}

        {phase === 'scanning' && (
          <div className={styles.scanningState}>
            <p className={styles.scanningLabel}>Analyse du reçu en cours...</p>
            <div className={styles.progressBar}>
              <div
                className={styles.progressFill}
                style={{ '--progress': `${progress}%` } as React.CSSProperties}
              />
            </div>
            <p className={styles.progressPct}>{progress}%</p>
          </div>
        )}

        {phase === 'preview' && (
          <div className={styles.previewPanel}>
            <p className={styles.previewTitle}>Données extraites</p>

            <div className={styles.field}>
              <label className={styles.fieldLabel}>Enseigne</label>
              <input
                type="text"
                className={styles.fieldInput}
                value={preview.merchant}
                onChange={(e) => setPreview((p) => ({ ...p, merchant: e.target.value }))}
                placeholder="Nom du marchand"
              />
            </div>

            <div className={styles.field}>
              <label className={styles.fieldLabel}>Montant (€)</label>
              <input
                type="number"
                className={styles.fieldInput}
                value={preview.amount}
                onChange={(e) => setPreview((p) => ({ ...p, amount: e.target.value }))}
                step="0.01"
                min="0"
                placeholder="0.00"
              />
            </div>

            <div className={styles.field}>
              <label className={styles.fieldLabel}>Date</label>
              <input
                type="text"
                className={styles.fieldInput}
                value={preview.date}
                onChange={(e) => setPreview((p) => ({ ...p, date: e.target.value }))}
                placeholder="JJ/MM/AAAA"
              />
            </div>

            <div className={styles.actions}>
              <button className={styles.confirmBtn} onClick={handleConfirm}>
                Confirmer
              </button>
              <button className={styles.retryBtn} onClick={handleRetry}>
                Nouvelle image
              </button>
            </div>
          </div>
        )}

        {phase === 'error' && (
          <div className={styles.errorState}>
            <p className={styles.errorMsg}>{errorMsg}</p>
            <button className={styles.retryBtn} onClick={handleRetry} aria-label="Réessayer">
              Réessayer
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
