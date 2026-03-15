import { useState } from 'react'
import { StarRating } from '../scouter/StarRating'
import { useRecordPurchase, useUpdatePurchase } from '../../hooks/usePurchase'
import { useNegotiation } from '../../hooks/useNegotiation'
import { NegotiationCoach } from './NegotiationCoach'
import type { PurchaseRecord } from '../../api/purchase'
import styles from './PurchaseForm.module.css'

interface PurchaseFormProps {
  missionId: string
  selectedOptionId?: string | null
  existingRecord?: PurchaseRecord | null
}

export function PurchaseForm({ missionId, selectedOptionId, existingRecord }: PurchaseFormProps) {
  const today = new Date().toISOString().split('T')[0]
  const [purchasedAt, setPurchasedAt] = useState(existingRecord?.purchasedAt?.split('T')[0] ?? today)
  const [finalPrice, setFinalPrice] = useState(String(existingRecord?.finalPrice ?? ''))
  const [merchant, setMerchant] = useState(existingRecord?.merchant ?? '')
  const [satisfaction, setSatisfaction] = useState<number | null>(existingRecord?.satisfaction ?? null)
  const [review, setReview] = useState(existingRecord?.review ?? '')
  const [success, setSuccess] = useState(false)
  const [showCoach, setShowCoach] = useState(false)

  const record = useRecordPurchase(missionId)
  const update = useUpdatePurchase(missionId)
  const negotiation = useNegotiation(selectedOptionId ?? '')

  const isEditing = !!existingRecord

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSuccess(false)
    const price = parseFloat(finalPrice)
    if (!purchasedAt || isNaN(price) || !merchant.trim()) return

    if (isEditing) {
      update.mutate(
        {
          merchant: merchant || undefined,
          finalPrice: price,
          satisfaction: satisfaction ?? undefined,
          review: review || undefined,
        },
        { onSuccess: () => setSuccess(true) },
      )
    } else {
      record.mutate(
        {
          purchasedAt,
          finalPrice: price,
          merchant,
          satisfaction: satisfaction ?? undefined,
          review: review || undefined,
        },
        { onSuccess: () => setSuccess(true) },
      )
    }
  }

  const isPending = record.isPending || update.isPending

  return (
    <div className={styles.wrap}>
      <h3 className={styles.title}>{isEditing ? 'Edit Purchase' : 'Record Purchase'}</h3>
      <form className={styles.form} onSubmit={handleSubmit}>
        {!isEditing && (
          <div className={styles.field}>
            <label className={styles.label}>Purchase Date</label>
            <input
              type="date"
              className={styles.input}
              value={purchasedAt}
              onChange={(e) => setPurchasedAt(e.target.value)}
              required
            />
          </div>
        )}
        <div className={styles.field}>
          <label className={styles.label}>Merchant</label>
          <input
            type="text"
            className={styles.input}
            value={merchant}
            onChange={(e) => setMerchant(e.target.value)}
            placeholder="e.g. Amazon, Fnac, Darty"
            required
          />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>Final Price (€)</label>
          <input
            type="number"
            className={styles.input}
            value={finalPrice}
            onChange={(e) => setFinalPrice(e.target.value)}
            step="0.01"
            min="0"
            required
          />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>Satisfaction</label>
          <StarRating value={satisfaction} onChange={setSatisfaction} />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>Review (optional)</label>
          <textarea
            className={styles.textarea}
            value={review}
            onChange={(e) => setReview(e.target.value)}
            placeholder="What went well? What would you do differently?"
            rows={3}
          />
        </div>
        {success && <p className={styles.success}>Purchase recorded successfully!</p>}
        {(record.isError || update.isError) && (
          <p className={styles.error}>Failed to save. Please try again.</p>
        )}
        <div className={styles.actions}>
          <button type="submit" className={styles.submit} disabled={isPending}>
            {isPending ? 'Saving...' : isEditing ? 'Update Purchase' : 'Record Purchase'}
          </button>
          {selectedOptionId && (
            <button
              type="button"
              className={styles.coachBtn}
              disabled={negotiation.isPending}
              onClick={() => {
                negotiation.reset()
                setShowCoach(false)
                negotiation.coach(undefined, {
                  onSuccess: () => setShowCoach(true),
                })
              }}
            >
              {negotiation.isPending ? 'Analyzing...' : 'Coach Me'}
            </button>
          )}
        </div>
        {negotiation.error && (
          <p className={styles.error}>Failed to load negotiation script. Please try again.</p>
        )}
      </form>
      {showCoach && negotiation.script && (
        <div className={styles.coachPanel}>
          <NegotiationCoach
            script={negotiation.script}
            onClose={() => {
              setShowCoach(false)
              negotiation.reset()
            }}
          />
        </div>
      )}
    </div>
  )
}
