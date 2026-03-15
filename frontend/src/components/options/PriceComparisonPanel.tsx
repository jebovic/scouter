import { useTranslation } from 'react-i18next'
import { usePriceComparison, useRefreshPriceComparison } from '../../hooks/usePriceComp'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './PriceComparisonPanel.module.css'

interface Props {
  optionId: string
}

export function PriceComparisonPanel({ optionId }: Props) {
  const { t } = useTranslation()
  const { comparison, isLoading, error } = usePriceComparison(optionId)
  const { refresh, isPending } = useRefreshPriceComparison(optionId)
  const { fmt, locale } = useFormatCurrency()

  if (isLoading) return null
  if (error) return null
  if (!comparison) return null

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <span className={styles.title}>{t('pricecomp.title')}</span>
        <button
          className={styles.refreshBtn}
          disabled={isPending}
          onClick={() => refresh()}
        >
          {isPending ? t('pricecomp.refreshing') : t('pricecomp.refresh')}
        </button>
      </div>

      <div className={styles.rows}>
        {comparison.offers.map((offer, idx) => {
          const isBest = idx === comparison.lowestIdx
          const rowClass = [
            styles.row,
            isBest ? styles.rowBest : '',
            !offer.available ? styles.rowUnavailable : '',
          ].filter(Boolean).join(' ')

          return (
            <div key={offer.retailer} className={rowClass}>
              <span className={styles.retailerName}>{offer.retailer}</span>
              <span className={`${styles.price} ${isBest ? styles.priceBest : styles.priceNormal}`}>
                {fmt(offer.price)}
              </span>
              {isBest && offer.available && (
                <span className={styles.bestBadge}>{t('pricecomp.best')}</span>
              )}
              {!offer.available && (
                <span className={styles.unavailBadge}>{t('pricecomp.unavailable')}</span>
              )}
              {offer.available && offer.url && (
                <a
                  href={offer.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className={styles.buyLink}
                >
                  {t('pricecomp.buy')} →
                </a>
              )}
            </div>
          )
        })}
      </div>

      <div className={styles.meta}>
        {t('pricecomp.updated')}: {new Date(comparison.fetchedAt).toLocaleString(locale)}
      </div>
    </div>
  )
}
