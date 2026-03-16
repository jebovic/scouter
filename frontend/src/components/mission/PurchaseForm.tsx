import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { StarRating } from '../scouter/StarRating'
import { useRecordPurchase, useUpdatePurchase } from '../../hooks/usePurchase'
import { useNegotiation } from '../../hooks/useNegotiation'
import { NegotiationCoach } from './NegotiationCoach'
import type { PurchaseRecord } from '../../api/purchase'
import styles from './PurchaseForm.module.css'

export interface PurchaseFormPrefill {
  merchant?: string
  finalPrice?: number
  purchasedAt?: string
  itemName?: string
}

interface PurchaseFormProps {
  missionId: string
  selectedOptionId?: string | null
  existingRecord?: PurchaseRecord | null
  prefill?: PurchaseFormPrefill
}

export function PurchaseForm({ missionId, selectedOptionId, existingRecord, prefill }: PurchaseFormProps) {
  const { t } = useTranslation()
  const today = new Date().toISOString().split('T')[0]
  const [purchasedAt, setPurchasedAt] = useState(existingRecord?.purchasedAt?.split('T')[0] ?? prefill?.purchasedAt ?? today)
  const [finalPrice, setFinalPrice] = useState(String(existingRecord?.finalPrice ?? prefill?.finalPrice ?? ''))
  const [merchant, setMerchant] = useState(existingRecord?.merchant ?? prefill?.merchant ?? '')
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
      <h3 className={styles.title}>{isEditing ? t('purchase.edit') : t('purchase.record')}</h3>
      <form className={styles.form} onSubmit={handleSubmit}>
        {prefill?.itemName && !isEditing && (
          <div className={styles.field}>
            <label className={styles.label}>{t('shopping.itemName')}</label>
            <input
              type="text"
              className={styles.input}
              value={prefill.itemName}
              readOnly
              aria-readonly="true"
            />
          </div>
        )}
        {!isEditing && (
          <div className={styles.field}>
            <label className={styles.label}>{t('purchase.purchaseDate')}</label>
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
          <label className={styles.label}>{t('purchase.merchant')}</label>
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
          <label className={styles.label}>{t('purchase.finalPrice')}</label>
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
          <label className={styles.label}>{t('purchase.satisfaction')}</label>
          <StarRating value={satisfaction} onChange={setSatisfaction} />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>{t('purchase.review')}</label>
          <textarea
            className={styles.textarea}
            value={review}
            onChange={(e) => setReview(e.target.value)}
            placeholder="What went well? What would you do differently?"
            rows={3}
          />
        </div>
        {success && <p className={styles.success}>{t('purchase.success')}</p>}
        {(record.isError || update.isError) && (
          <p className={styles.error}>{t('purchase.error')}</p>
        )}
        <div className={styles.actions}>
          <button type="submit" className={styles.submit} disabled={isPending}>
            {isPending ? t('purchase.saving') : isEditing ? t('purchase.update') : t('purchase.record')}
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
              {negotiation.isPending ? t('negotiation.analyzing') : t('negotiation.coachMe')}
            </button>
          )}
        </div>
        {negotiation.error && (
          <p className={styles.error}>{t('negotiation.error')}</p>
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
