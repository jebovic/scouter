import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Topnav } from '../components/scouter/Topnav'
import { BudgetRebalancer } from '../components/scouter/BudgetRebalancer'
import { useEnvelopes, useCreateEnvelope, useUpdateEnvelope, useDeleteEnvelope, useEnvelopeSummary } from '../hooks'
import type { EnvelopeDTO } from '../api/envelope'
import styles from './EnvelopesPage.module.css'

const CURRENCIES = ['EUR', 'USD', 'GBP', 'CHF', 'CAD']

interface FormState {
  name: string
  monthlyAmount: string
  currency: string
}

const defaultForm = (): FormState => ({ name: '', monthlyAmount: '', currency: 'EUR' })

function EnvelopeForm({
  initial,
  title,
  isPending,
  onSave,
  onCancel,
}: {
  initial?: FormState
  title: string
  isPending: boolean
  onSave: (f: FormState) => void
  onCancel: () => void
}) {
  const [form, setForm] = useState<FormState>(initial ?? defaultForm())

  function set(k: keyof FormState, v: string) {
    setForm((p) => ({ ...p, [k]: v }))
  }

  function submit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name.trim() || !form.monthlyAmount) return
    onSave(form)
  }

  return (
    <form className={styles.form} onSubmit={submit}>
      <div className={styles.formTitle}>{title}</div>
      <div className={styles.field}>
        <label className={styles.label}>Name</label>
        <input
          className={styles.input}
          value={form.name}
          onChange={(e) => set('name', e.target.value)}
          placeholder="e.g. Alimentation, Loisirs…"
          required
        />
      </div>
      <div className={styles.fieldRow}>
        <div className={styles.field}>
          <label className={styles.label}>Monthly budget</label>
          <input
            className={styles.input}
            type="number"
            min={0}
            step={0.01}
            value={form.monthlyAmount}
            onChange={(e) => set('monthlyAmount', e.target.value)}
            placeholder="300"
            required
          />
        </div>
        <div className={styles.field} style={{ maxWidth: 90 }}>
          <label className={styles.label}>Currency</label>
          <select
            className={styles.input}
            value={form.currency}
            onChange={(e) => set('currency', e.target.value)}
          >
            {CURRENCIES.map((c) => (
              <option key={c}>{c}</option>
            ))}
          </select>
        </div>
      </div>
      <div className={styles.formActions}>
        <button type="button" className={styles.cancelBtn} onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className={styles.saveBtn} disabled={isPending}>
          {isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </form>
  )
}

function EnvelopeSummaryBadge({ envelopeId, monthlyAmount, currency }: { envelopeId: string; monthlyAmount: number; currency: string }) {
  const { summary, isLoading } = useEnvelopeSummary(envelopeId)

  const fmt = new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  })

  if (isLoading) {
    return <div className={styles.summaryLoading} aria-label="Loading summary" />
  }

  if (!summary) return null

  const fillPct = monthlyAmount > 0 ? Math.min(100, (summary.totalSpent / monthlyAmount) * 100) : 0
  const fillColor =
    fillPct < 80 ? 'var(--green)' : fillPct <= 100 ? 'var(--gold)' : 'var(--coral)'

  return (
    <div className={styles.summary}>
      <div className={styles.summaryMeta}>
        <span className={styles.summaryCount}>
          {summary.missionCount} mission{summary.missionCount !== 1 ? 's' : ''} linked
        </span>
      </div>
      <div className={styles.summaryAmounts}>
        <span>Budget: {fmt.format(summary.totalBudget)}</span>
        <span className={styles.summaryDivider}>/</span>
        <span>Spent: {fmt.format(summary.totalSpent)}</span>
      </div>
      <div className={styles.summaryBar}>
        <div
          className={styles.summaryFill}
          style={{ width: `${fillPct}%`, background: fillColor } as React.CSSProperties}
        />
      </div>
    </div>
  )
}

function EnvelopeCard({ env }: { env: EnvelopeDTO }) {
  const [editing, setEditing] = useState(false)
  const { update, isPending: updating } = useUpdateEnvelope()
  const { remove, isPending: deleting } = useDeleteEnvelope()

  const fmt = new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency: env.currency,
    maximumFractionDigits: 0,
  })

  if (editing) {
    return (
      <EnvelopeForm
        title="Edit envelope"
        initial={{
          name: env.name,
          monthlyAmount: String(env.monthlyAmount),
          currency: env.currency,
        }}
        isPending={updating}
        onSave={async (f) => {
          await update({
            id: env.id,
            payload: {
              name: f.name,
              monthlyAmount: parseFloat(f.monthlyAmount),
              currency: f.currency,
            },
          })
          setEditing(false)
        }}
        onCancel={() => setEditing(false)}
      />
    )
  }

  return (
    <div className={styles.card}>
      <div className={styles.cardHeader}>
        <span className={styles.cardName}>{env.name}</span>
        <div className={styles.cardActions}>
          <button
            className={styles.iconBtn}
            onClick={() => setEditing(true)}
            aria-label="Edit envelope"
            title="Edit"
          >
            ✎
          </button>
          <button
            className={`${styles.iconBtn} ${styles.danger}`}
            onClick={() => remove(env.id)}
            disabled={deleting}
            aria-label="Delete envelope"
            title="Delete"
          >
            ✕
          </button>
        </div>
      </div>
      <div>
        <span className={styles.amount}>{fmt.format(env.monthlyAmount)}</span>
        <span className={styles.period}> / month</span>
      </div>
      <EnvelopeSummaryBadge
        envelopeId={env.id}
        monthlyAmount={env.monthlyAmount}
        currency={env.currency}
      />
    </div>
  )
}

export default function EnvelopesPage() {
  const { t } = useTranslation()
  const { envelopes, isLoading } = useEnvelopes()
  const { create, isPending: creating } = useCreateEnvelope()
  const [showForm, setShowForm] = useState(false)

  const total = envelopes.reduce((s, e) => s + e.monthlyAmount, 0)
  // Arbitrary cap to show a progress-style health bar: 5 000 € / month reference
  const HEALTH_CAP = 5000
  const healthPct = Math.min(100, (total / HEALTH_CAP) * 100)
  const healthColor =
    healthPct < 60 ? 'var(--green)' : healthPct < 85 ? 'var(--gold)' : 'var(--coral)'

  const fmtEur = new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency: 'EUR',
    maximumFractionDigits: 0,
  })

  return (
    <>
      <Topnav />
      <main className={styles.page}>
        <div className={styles.header}>
          <div>
            <div className={styles.title}>Envelopes</div>
            <div className={styles.subtitle}>
              {t('envelopes.subtitle', 'Named monthly budget allocations')}
            </div>
          </div>
          {!showForm && (
            <button className={styles.addBtn} onClick={() => setShowForm(true)}>
              + New envelope
            </button>
          )}
        </div>

        {/* Budget health summary */}
        {envelopes.length > 0 && (
          <div className={styles.health}>
            <span className={styles.healthLabel}>MONTHLY BUDGET</span>
            <div className={styles.healthBar}>
              <div
                className={styles.healthFill}
                style={{ width: `${healthPct}%`, background: healthColor }}
              />
            </div>
            <div>
              <div className={styles.healthTotal}>{fmtEur.format(total)}</div>
              <div className={styles.healthSub}>
                {envelopes.length} envelope{envelopes.length !== 1 ? 's' : ''}
              </div>
            </div>
          </div>
        )}

        {showForm && (
          <div style={{ marginBottom: '1.5rem' }}>
            <EnvelopeForm
              title="New envelope"
              isPending={creating}
              onSave={async (f) => {
                await create({
                  name: f.name,
                  monthlyAmount: parseFloat(f.monthlyAmount),
                  currency: f.currency,
                })
                setShowForm(false)
              }}
              onCancel={() => setShowForm(false)}
            />
          </div>
        )}

        {isLoading ? (
          <div className={styles.grid}>
            {[1, 2, 3].map((k) => (
              <div key={k} className={styles.card} style={{ minHeight: 120, opacity: 0.4 }} />
            ))}
          </div>
        ) : (
          <div className={styles.grid}>
            {envelopes.length === 0 && !showForm ? (
              <div className={styles.empty}>
                <div className={styles.emptyIcon}>💰</div>
                <div className={styles.emptyTitle}>No envelopes yet</div>
                <div className={styles.emptyDesc}>
                  Create budget envelopes to allocate your monthly spending.
                </div>
              </div>
            ) : (
              envelopes.map((env) => <EnvelopeCard key={env.id} env={env} />)
            )}
          </div>
        )}

        <BudgetRebalancer />
      </main>
    </>
  )
}
