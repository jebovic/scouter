import { useState } from 'react'
import { useConditionPricing } from '../../hooks/useConditionPricing'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { ConditionPrice } from '../../api/conditionpricing'
import styles from './ConditionPricingCard.module.css'

interface ConditionPricingCardProps {
  missionId: string
  itemId: string
}

function GradeBadge({ grade }: { grade: string }) {
  const cls =
    grade === 'A+' ? styles.gradeAPlus
    : grade === 'A' ? styles.gradeA
    : grade === 'B' ? styles.gradeB
    : styles.gradeC
  return (
    <span className={`${styles.gradeBadge} ${cls}`} aria-label={`Grade ${grade}`}>
      {grade}
    </span>
  )
}

function AvailabilityDot({ availability }: { availability: string }) {
  const cls =
    availability === 'Disponible' ? styles.availGreen
    : availability === 'Stock limité' ? styles.availOrange
    : styles.availRed
  return (
    <span
      className={`${styles.availDot} ${cls}`}
      title={availability}
      aria-label={availability}
    />
  )
}

function ConditionRow({ cp, isBest, fmt }: { cp: ConditionPrice; isBest: boolean; fmt: (n: number) => string }) {
  return (
    <tr className={`${styles.condRow} ${isBest ? styles.bestRow : ''}`}>
      <td className={styles.condCell}>
        <AvailabilityDot availability={cp.availability} />
        <GradeBadge grade={cp.grade} />
        <span className={styles.condLabel}>{cp.condition}</span>
        {isBest && (
          <span className={styles.bestBadge} aria-label="Meilleure affaire">
            Meilleur deal
          </span>
        )}
      </td>
      <td className={styles.priceCell}>
        {fmt(cp.estPrice)}
      </td>
      <td className={styles.discountCell}>
        <span className={styles.discountBadge}>-{cp.discount.toFixed(0)}%</span>
      </td>
      <td className={styles.platformCell}>
        <a
          href={cp.searchUrl}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.platformLink}
        >
          {cp.platform}
        </a>
      </td>
      <td className={styles.warrantyCell}>
        {cp.warrantyMonths} mois
      </td>
    </tr>
  )
}

export function ConditionPricingCard({ missionId, itemId }: ConditionPricingCardProps) {
  const [enabled, setEnabled] = useState(false)
  const [open, setOpen] = useState(false)

  const { data, isFetching } = useConditionPricing(missionId, itemId, enabled)
  const { fmt } = useFormatCurrency()

  function handleToggle() {
    setEnabled(true)
    setOpen((v) => !v)
  }

  return (
    <div className={styles.wrapper}>
      <button
        className={styles.triggerBtn}
        onClick={handleToggle}
        aria-expanded={open}
        aria-label="Voir les prix reconditionnés"
        title="Comparer les prix reconditionnés (Back Market, Recommerce, Fnac)"
      >
        {isFetching ? (
          <span className={styles.spinner} aria-hidden="true" />
        ) : (
          '♻️ Reconditionné'
        )}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Prix reconditionnés">
          <h3 className={styles.heading}>♻️ Prix Reconditionnés</h3>

          {isFetching && !data ? (
            <p className={styles.loading}>Chargement des prix reconditionnés...</p>
          ) : data ? (
            <>
              {data.bestDeal && (
                <div className={styles.savingsBanner} role="status">
                  <span aria-hidden="true">💰</span>
                  <span>
                    Économisez{' '}
                    <strong>
                      {fmt(data.newPrice - data.bestDeal.estPrice)}
                    </strong>
                    {' '}en choisissant{' '}
                    <a
                      href={data.bestDeal.searchUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={styles.bestLink}
                    >
                      {data.bestDeal.platform}
                    </a>{' '}
                    (grade {data.bestDeal.grade})
                  </span>
                </div>
              )}

              <table className={styles.table} aria-label={`Prix reconditionnés pour ${data.itemName}`}>
                <thead>
                  <tr>
                    <th className={styles.thCond}>État</th>
                    <th className={styles.thPrice}>Prix</th>
                    <th className={styles.thDiscount}>Remise</th>
                    <th className={styles.thPlatform}>Plateforme</th>
                    <th className={styles.thWarranty}>Garantie</th>
                  </tr>
                </thead>
                <tbody>
                  {data.conditions.map((cp) => (
                    <ConditionRow
                      key={cp.grade}
                      cp={cp}
                      isBest={data.bestDeal?.grade === cp.grade}
                      fmt={fmt}
                    />
                  ))}
                </tbody>
              </table>

              <p className={styles.recommendation}>{data.recommendation}</p>
              <p className={styles.disclaimer}>
                Prix estimés à titre indicatif. Cliquez sur la plateforme pour vérifier la disponibilité réelle.
              </p>

              <button
                className={styles.closeBtn}
                onClick={() => setOpen(false)}
                aria-label="Fermer les prix reconditionnés"
              >
                Fermer
              </button>
            </>
          ) : null}
        </div>
      )}
    </div>
  )
}
