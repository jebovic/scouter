import { useState } from 'react'
import { useCompetitorPrices } from '../../hooks/useCompetitorPrices'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { CompetitorOffer } from '../../api/competitorprice'
import styles from './CompetitorPriceTable.module.css'

interface CompetitorPriceTableProps {
  missionId: string
  itemId: string
}

function DiffBadge({ pct }: { pct: number }) {
  const cheaper = pct < 0
  const label = `${cheaper ? '' : '+'}${pct.toFixed(1)}%`
  return (
    <span
      className={`${styles.diffBadge} ${cheaper ? styles.diffCheaper : styles.diffExpensive}`}
      aria-label={`${cheaper ? 'moins cher' : 'plus cher'} de ${Math.abs(pct).toFixed(1)}%`}
    >
      {label}
    </span>
  )
}

function AvailabilityDot({ availability }: { availability: string }) {
  const inStock = availability === 'En stock'
  return (
    <span
      className={`${styles.availDot} ${inStock ? styles.availInStock : styles.availOrder}`}
      title={availability}
      aria-label={availability}
    />
  )
}

function OfferRow({ offer, fmt }: { offer: CompetitorOffer; fmt: (n: number) => string }) {
  return (
    <tr className={`${styles.offerRow} ${offer.isBest ? styles.bestRow : ''}`}>
      <td className={styles.platformCell}>
        <AvailabilityDot availability={offer.availability} />
        <a
          href={offer.url}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.platformLink}
        >
          {offer.platform}
        </a>
        {offer.isBest && (
          <span className={styles.bestBadge} aria-label="Meilleure offre">
            Meilleur prix
          </span>
        )}
      </td>
      <td className={styles.priceCell}>
        {fmt(offer.price)}
      </td>
      <td className={styles.diffCell}>
        <DiffBadge pct={offer.priceDiffPct} />
      </td>
      <td className={styles.availCell}>{offer.availability}</td>
    </tr>
  )
}

export function CompetitorPriceTable({ missionId, itemId }: CompetitorPriceTableProps) {
  const [enabled, setEnabled] = useState(false)
  const [open, setOpen] = useState(false)

  const { data, isFetching } = useCompetitorPrices(missionId, itemId, enabled)
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
        aria-label="Comparer les prix concurrents"
        title="Comparer les prix sur les sites français"
      >
        {isFetching ? <span className={styles.spinner} aria-hidden="true" /> : '🏪 Comparer'}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Comparaison des prix concurrents">
          {isFetching && !data ? (
            <p className={styles.loading}>Chargement des prix...</p>
          ) : data ? (
            <>
              {data.savings > 0 && (
                <div className={styles.savingsBanner} role="status">
                  <span className={styles.savingsIcon} aria-hidden="true">💰</span>
                  <span>
                    Économie possible&nbsp;:{' '}
                    <strong>
                      {fmt(data.savings)}
                    </strong>
                    {data.bestOffer && (
                      <> chez <a href={data.bestOffer.url} target="_blank" rel="noopener noreferrer" className={styles.bestLink}>{data.bestOffer.platform}</a></>
                    )}
                  </span>
                </div>
              )}

              <table className={styles.table} aria-label={`Prix pour ${data.itemName}`}>
                <thead>
                  <tr>
                    <th className={styles.thPlatform}>Site</th>
                    <th className={styles.thPrice}>Prix</th>
                    <th className={styles.thDiff}>Écart</th>
                    <th className={styles.thAvail}>Dispo</th>
                  </tr>
                </thead>
                <tbody>
                  {data.competitors.map((offer) => (
                    <OfferRow key={offer.platform} offer={offer} fmt={fmt} />
                  ))}
                </tbody>
              </table>

              <p className={styles.disclaimer}>
                Prix simulés à titre indicatif. Cliquez sur un site pour vérifier le prix réel.
              </p>

              <button
                className={styles.closeBtn}
                onClick={() => setOpen(false)}
                aria-label="Fermer la comparaison"
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
