import { useState } from 'react'
import { useLocalEnvelopes } from '../hooks/useLocalEnvelopes'
import { useFormatCurrency } from '../hooks/useFormatCurrency'
import { PRESET_COLORS, DEFAULT_EMOJIS, type Transaction } from '../utils/envelopes'
import styles from './EnvelopesPage.module.css'

// ── Formatting helper ─────────────────────────────────────────────────────────

function fmtDate(iso: string, locale: string): string {
  try {
    return new Intl.DateTimeFormat(locale, { day: '2-digit', month: 'short' }).format(
      new Date(iso),
    )
  } catch {
    return iso
  }
}

// ── New Envelope Form ─────────────────────────────────────────────────────────

interface NewEnvelopeFormProps {
  onSave: (fields: {
    name: string
    emoji: string
    budgetEur: number
    color: string
  }) => void
  onCancel: () => void
}

function NewEnvelopeForm({ onSave, onCancel }: NewEnvelopeFormProps) {
  const [name, setName] = useState('')
  const [emoji, setEmoji] = useState('💰')
  const [budget, setBudget] = useState('')
  const [color, setColor] = useState(PRESET_COLORS[0])
  const [emojiOpen, setEmojiOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = name.trim()
    if (!trimmed) { setError('Le nom est requis'); return }
    const amount = parseFloat(budget)
    if (isNaN(amount) || amount <= 0) { setError('Budget invalide'); return }
    onSave({ name: trimmed, emoji, budgetEur: amount, color })
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit} noValidate>
      <div className={styles.formTitle}>Nouvelle enveloppe</div>

      {/* Emoji picker */}
      <div className={styles.field}>
        <label className={styles.label}>Emoji</label>
        <div className={styles.emojiPickerWrap}>
          <button
            type="button"
            className={styles.emojiToggle}
            onClick={() => setEmojiOpen((v) => !v)}
            aria-label="Choisir un emoji"
          >
            {emoji}
          </button>
          {emojiOpen && (
            <div className={styles.emojiGrid}>
              {DEFAULT_EMOJIS.map((e) => (
                <button
                  key={e}
                  type="button"
                  className={styles.emojiOption}
                  onClick={() => { setEmoji(e); setEmojiOpen(false) }}
                >
                  {e}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className={styles.fieldRow}>
        <div className={styles.field}>
          <label className={styles.label}>Nom</label>
          <input
            className={styles.input}
            value={name}
            onChange={(e) => { setName(e.target.value); setError(null) }}
            placeholder="ex: Alimentation, Loisirs…"
            required
          />
        </div>
        <div className={styles.field} style={{ maxWidth: 120 }}>
          <label className={styles.label}>Budget (€)</label>
          <input
            className={styles.input}
            type="number"
            min={0.01}
            step={0.01}
            value={budget}
            onChange={(e) => { setBudget(e.target.value); setError(null) }}
            placeholder="300"
            required
          />
        </div>
      </div>

      {/* Color picker */}
      <div className={styles.field}>
        <label className={styles.label}>Couleur</label>
        <div className={styles.colorGrid}>
          {PRESET_COLORS.map((c) => (
            <button
              key={c}
              type="button"
              className={styles.colorDot}
              style={{ '--dot-color': c, outline: color === c ? `2px solid ${c}` : undefined } as React.CSSProperties}
              onClick={() => setColor(c)}
              aria-label={`Couleur ${c}`}
            />
          ))}
        </div>
      </div>

      {error && <p className={styles.formError}>{error}</p>}

      <div className={styles.formActions}>
        <button type="button" className={styles.cancelBtn} onClick={onCancel}>
          Annuler
        </button>
        <button type="submit" className={styles.saveBtn}>
          Créer
        </button>
      </div>
    </form>
  )
}

// ── Transaction Form ──────────────────────────────────────────────────────────

interface TransactionFormProps {
  envelopeId: string
  onSave: (fields: Omit<Transaction, 'id'>) => void
  onCancel: () => void
}

function TransactionForm({ envelopeId, onSave, onCancel }: TransactionFormProps) {
  const today = new Date().toISOString().slice(0, 10)
  const [label, setLabel] = useState('')
  const [amount, setAmount] = useState('')
  const [date, setDate] = useState(today)
  const [type, setType] = useState<'expense' | 'refill'>('expense')
  const [error, setError] = useState<string | null>(null)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmedLabel = label.trim()
    if (!trimmedLabel) { setError('Libellé requis'); return }
    const parsed = parseFloat(amount)
    if (isNaN(parsed) || parsed <= 0) { setError('Montant invalide'); return }
    const finalAmount = type === 'expense' ? -parsed : parsed
    onSave({ envelopeId, label: trimmedLabel, amountEur: finalAmount, date })
  }

  return (
    <form className={styles.txForm} onSubmit={handleSubmit} noValidate>
      <div className={styles.formTitle}>Ajouter une opération</div>

      <div className={styles.typeToggle}>
        <button
          type="button"
          className={type === 'expense' ? styles.typeActive : styles.typeInactive}
          onClick={() => setType('expense')}
        >
          Dépense
        </button>
        <button
          type="button"
          className={type === 'refill' ? styles.typeActiveGreen : styles.typeInactive}
          onClick={() => setType('refill')}
        >
          Rechargement
        </button>
      </div>

      <div className={styles.fieldRow}>
        <div className={styles.field}>
          <label className={styles.label}>Libellé</label>
          <input
            className={styles.input}
            value={label}
            onChange={(e) => { setLabel(e.target.value); setError(null) }}
            placeholder="ex: Supermarché, Netflix…"
            required
          />
        </div>
        <div className={styles.field} style={{ maxWidth: 110 }}>
          <label className={styles.label}>Montant (€)</label>
          <input
            className={styles.input}
            type="number"
            min={0.01}
            step={0.01}
            value={amount}
            onChange={(e) => { setAmount(e.target.value); setError(null) }}
            placeholder="25"
            required
          />
        </div>
        <div className={styles.field} style={{ maxWidth: 140 }}>
          <label className={styles.label}>Date</label>
          <input
            className={styles.input}
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </div>
      </div>

      {error && <p className={styles.formError}>{error}</p>}

      <div className={styles.formActions}>
        <button type="button" className={styles.cancelBtn} onClick={onCancel}>
          Annuler
        </button>
        <button type="submit" className={styles.saveBtn}>
          Enregistrer
        </button>
      </div>
    </form>
  )
}

// ── Envelope Card ─────────────────────────────────────────────────────────────

interface EnvelopeCardProps {
  envelope: {
    id: string
    name: string
    emoji: string
    budgetEur: number
    spentEur: number
    color: string
  }
  onDelete: (id: string) => void
  onAddTransaction: (id: string) => void
  fmt: (amount: number) => string
}

function EnvelopeCard({ envelope, onDelete, onAddTransaction, fmt }: EnvelopeCardProps) {
  const remaining = envelope.budgetEur - envelope.spentEur
  const pct = envelope.budgetEur > 0
    ? Math.min(100, (envelope.spentEur / envelope.budgetEur) * 100)
    : 0
  const barColor =
    pct < 75 ? 'var(--green)' : pct < 100 ? 'var(--orange)' : 'var(--coral)'

  return (
    <div
      className={styles.card}
      style={{ '--env-color': envelope.color } as React.CSSProperties}
    >
      {/* Card top accent */}
      <div className={styles.cardAccent} />

      <div className={styles.cardHeader}>
        <span className={styles.cardEmoji}>{envelope.emoji}</span>
        <span className={styles.cardName}>{envelope.name}</span>
        <button
          type="button"
          className={`${styles.iconBtn} ${styles.danger}`}
          onClick={() => onDelete(envelope.id)}
          aria-label={`Supprimer ${envelope.name}`}
          title="Supprimer"
        >
          ✕
        </button>
      </div>

      <div className={styles.cardAmounts}>
        <div>
          <div className={styles.amountLabel}>Dépensé</div>
          <div className={styles.amountValue} style={{ color: barColor }}>
            {fmt(envelope.spentEur)}
          </div>
        </div>
        <div className={styles.amountDivider}>/</div>
        <div>
          <div className={styles.amountLabel}>Budget</div>
          <div className={styles.amountBudget}>{fmt(envelope.budgetEur)}</div>
        </div>
        <div className={styles.remainingBadge} style={{ color: remaining >= 0 ? 'var(--green)' : 'var(--coral)' }}>
          {remaining >= 0 ? `${fmt(remaining)} restant` : `${fmt(-remaining)} dépassé`}
        </div>
      </div>

      {/* Progress bar */}
      <div className={styles.progressTrack}>
        <div
          className={styles.progressFill}
          style={{ width: `${pct}%`, background: barColor }}
        />
      </div>

      <button
        type="button"
        className={styles.addTxBtn}
        onClick={() => onAddTransaction(envelope.id)}
      >
        + Ajouter une dépense
      </button>
    </div>
  )
}

// ── Recent Transactions ───────────────────────────────────────────────────────

interface RecentTxProps {
  transactions: Transaction[]
  envelopeMap: Record<string, { name: string; emoji: string; color: string }>
  onDelete: (id: string) => void
  fmt: (amount: number) => string
  locale: string
}

function RecentTransactions({ transactions, envelopeMap, onDelete, fmt, locale }: RecentTxProps) {
  const sorted = [...transactions].sort((a, b) => b.date.localeCompare(a.date)).slice(0, 20)

  if (sorted.length === 0) return null

  return (
    <section className={styles.txSection}>
      <h2 className={styles.sectionTitle}>Transactions récentes</h2>
      <div className={styles.txList}>
        {sorted.map((tx) => {
          const env = envelopeMap[tx.envelopeId]
          const isExpense = tx.amountEur < 0
          return (
            <div key={tx.id} className={styles.txRow}>
              <span className={styles.txEmoji}>{env?.emoji ?? '?'}</span>
              <div className={styles.txInfo}>
                <span className={styles.txLabel}>{tx.label}</span>
                <span className={styles.txEnvName}>{env?.name ?? 'Inconnu'}</span>
              </div>
              <span className={styles.txDate}>{fmtDate(tx.date, locale)}</span>
              <span
                className={styles.txAmount}
                style={{ color: isExpense ? 'var(--coral)' : 'var(--green)' }}
              >
                {isExpense ? '-' : '+'}{fmt(Math.abs(tx.amountEur))}
              </span>
              <button
                type="button"
                className={styles.txDeleteBtn}
                onClick={() => onDelete(tx.id)}
                aria-label="Supprimer la transaction"
              >
                ✕
              </button>
            </div>
          )
        })}
      </div>
    </section>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function EnvelopesPage() {
  const {
    envelopes,
    transactions,
    addEnvelope,
    deleteEnvelope,
    addTransaction,
    deleteTransaction,
  } = useLocalEnvelopes()
  const { fmt, locale } = useFormatCurrency()

  const [showNewForm, setShowNewForm] = useState(false)
  const [activeTxEnvId, setActiveTxEnvId] = useState<string | null>(null)

  const totalBudget = envelopes.reduce((s, e) => s + e.budgetEur, 0)
  const totalSpent = envelopes.reduce((s, e) => s + e.spentEur, 0)
  const globalPct = totalBudget > 0 ? Math.min(100, (totalSpent / totalBudget) * 100) : 0
  const globalColor =
    globalPct < 75 ? 'var(--green)' : globalPct < 100 ? 'var(--orange)' : 'var(--coral)'

  const envelopeMap: Record<string, { name: string; emoji: string; color: string }> = {}
  for (const env of envelopes) {
    envelopeMap[env.id] = { name: env.name, emoji: env.emoji, color: env.color }
  }

  function handleAddEnvelope(fields: {
    name: string
    emoji: string
    budgetEur: number
    color: string
  }) {
    addEnvelope(fields)
    setShowNewForm(false)
  }

  function handleAddTransaction(fields: Omit<Transaction, 'id'>) {
    addTransaction(fields)
    setActiveTxEnvId(null)
  }

  return (
    <main className={styles.page}>
      {/* Header */}
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Budget Enveloppes</h1>
          <p className={styles.subtitle}>Gérez votre budget par catégorie</p>
        </div>
        {!showNewForm && (
          <button
            type="button"
            className={styles.addBtn}
            onClick={() => setShowNewForm(true)}
          >
            + Nouvelle enveloppe
          </button>
        )}
      </div>

      {/* Global summary bar */}
      {envelopes.length > 0 && (
        <div className={styles.health}>
          <div className={styles.healthMeta}>
            <span className={styles.healthLabel}>BUDGET GLOBAL</span>
            <div className={styles.healthAmounts}>
              <span style={{ color: globalColor }}>{fmt(totalSpent)}</span>
              <span className={styles.healthSep}>/</span>
              <span className={styles.healthBudget}>{fmt(totalBudget)}</span>
              <span className={styles.healthPct}>{globalPct.toFixed(0)}%</span>
            </div>
          </div>
          <div className={styles.healthBar}>
            <div
              className={styles.healthFill}
              style={{ width: `${globalPct}%`, background: globalColor }}
            />
          </div>
        </div>
      )}

      {/* New envelope form */}
      {showNewForm && (
        <div className={styles.formWrap}>
          <NewEnvelopeForm
            onSave={handleAddEnvelope}
            onCancel={() => setShowNewForm(false)}
          />
        </div>
      )}

      {/* Envelope grid */}
      {envelopes.length === 0 && !showNewForm ? (
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>💸</div>
          <div className={styles.emptyTitle}>Aucune enveloppe</div>
          <div className={styles.emptyDesc}>
            Créez des enveloppes pour organiser votre budget par catégorie.
          </div>
          <button
            type="button"
            className={styles.addBtn}
            onClick={() => setShowNewForm(true)}
            style={{ marginTop: '1rem' }}
          >
            Créer ma première enveloppe
          </button>
        </div>
      ) : (
        <div className={styles.grid}>
          {envelopes.map((env) => (
            <div key={env.id}>
              <EnvelopeCard
                envelope={env}
                onDelete={deleteEnvelope}
                onAddTransaction={(id) => setActiveTxEnvId(activeTxEnvId === id ? null : id)}
                fmt={fmt}
              />
              {activeTxEnvId === env.id && (
                <div className={styles.txFormWrap}>
                  <TransactionForm
                    envelopeId={env.id}
                    onSave={handleAddTransaction}
                    onCancel={() => setActiveTxEnvId(null)}
                  />
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Recent transactions */}
      <RecentTransactions
        transactions={transactions}
        envelopeMap={envelopeMap}
        onDelete={deleteTransaction}
        fmt={fmt}
        locale={locale}
      />
    </main>
  )
}
