import { useState } from 'react'
import { useAlertRules } from '../../hooks/useAlertRules'
import type { AlertRule } from '../../api/pricealertrule'
import styles from './PriceAlertRules.module.css'

interface PriceAlertRulesProps {
  itemId: string
  currentPrice: number
}

const RULE_LABELS: Record<string, string> = {
  drop_percent: '↓%',
  fixed_below:  '<',
  weekly_low:   '7j bas',
}

function ruleSummary(rule: AlertRule): string {
  switch (rule.type) {
    case 'drop_percent':
      return `↓${rule.value}%`
    case 'fixed_below':
      return `< ${rule.value.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })}`
    case 'weekly_low':
      return '7j bas'
    default:
      return rule.type
  }
}

type RuleType = 'drop_percent' | 'fixed_below' | 'weekly_low'

const RULE_TYPES: { value: RuleType; label: string; hasValue: boolean }[] = [
  { value: 'drop_percent', label: 'Baisse de X%',         hasValue: true },
  { value: 'fixed_below',  label: 'Sous un seuil fixe',   hasValue: true },
  { value: 'weekly_low',   label: 'Plus bas de la semaine', hasValue: false },
]

export function PriceAlertRules({ itemId, currentPrice }: PriceAlertRulesProps) {
  const { rules, add, remove } = useAlertRules(itemId)
  const [showForm, setShowForm] = useState(false)
  const [ruleType, setRuleType] = useState<RuleType>('drop_percent')
  const [value, setValue] = useState('')

  const selectedDef = RULE_TYPES.find((rt) => rt.value === ruleType)!

  function handleSubmit() {
    const numVal = selectedDef.hasValue ? parseFloat(value) : 0
    if (selectedDef.hasValue && (isNaN(numVal) || numVal <= 0)) return

    add.mutate(
      { type: ruleType, value: numVal, basePrice: currentPrice },
      { onSuccess: () => { setShowForm(false); setValue('') } },
    )
  }

  return (
    <div className={styles.wrapper}>
      {rules.length > 0 && (
        <div className={styles.rulesRow}>
          {rules.map((rule) => (
            <span
              key={rule.id}
              className={`${styles.pill} ${rule.triggered ? styles.pillTriggered : ''}`}
              title={`${RULE_LABELS[rule.type] ?? rule.type} — ${ruleSummary(rule)}${rule.triggered ? ' — DÉCLENCHÉ' : ''}`}
            >
              {rule.triggered && <span className={styles.triggeredDot} aria-label="triggered" />}
              {ruleSummary(rule)}
              <button
                className={styles.deleteBtn}
                onClick={() => remove.mutate(rule.id)}
                aria-label={`Supprimer la règle ${ruleSummary(rule)}`}
                disabled={remove.isPending}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}

      {showForm ? (
        <div className={styles.form}>
          <select
            className={styles.select}
            value={ruleType}
            onChange={(e) => { setRuleType(e.target.value as RuleType); setValue('') }}
          >
            {RULE_TYPES.map((rt) => (
              <option key={rt.value} value={rt.value}>{rt.label}</option>
            ))}
          </select>

          {selectedDef.hasValue && (
            <input
              className={styles.input}
              type="number"
              min="0"
              step="any"
              placeholder={ruleType === 'drop_percent' ? '% ex: 10' : 'Prix ex: 89'}
              value={value}
              onChange={(e) => setValue(e.target.value)}
              autoFocus
            />
          )}

          <button
            className={styles.submitBtn}
            onClick={handleSubmit}
            disabled={add.isPending || (selectedDef.hasValue && (!value || parseFloat(value) <= 0))}
          >
            {add.isPending ? '…' : '+ Ajouter'}
          </button>
          <button className={styles.cancelBtn} onClick={() => { setShowForm(false); setValue('') }}>
            Annuler
          </button>
        </div>
      ) : (
        <button className={styles.addBtn} onClick={() => setShowForm(true)}>
          + alerte prix
        </button>
      )}
    </div>
  )
}
