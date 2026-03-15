import { usePriceComparison } from '../../hooks/usePriceComparison'
import { formatCurrency } from '../../utils/format'
import type { PriceComparison } from '../../api/pricecomparison'
import styles from './PriceComparisonTable.module.css'

interface PriceComparisonTableProps {
  missionId: string
  itemId: string
  currency?: string
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

function ComparisonContent({ data, currency }: { data: PriceComparison; currency: string }) {
  const fmt = (n: number) => formatCurrency(n, currency)

  return (
    <>
      <table className={styles.table} aria-label={`Comparaison de prix pour ${data.itemName}`}>
        <thead>
          <tr>
            <th className={styles.thRetailer}>Marchand</th>
            <th className={styles.thPrice}>Prix</th>
            <th className={styles.thDiff}>Écart</th>
            <th className={styles.thLink}>Lien</th>
          </tr>
        </thead>
        <tbody>
          {data.retailers.map((r) => (
            <tr
              key={r.retailer}
              className={`${styles.row} ${r.isBestPrice ? styles.bestRow : ''}`}
            >
              <td>
                <div className={styles.retailerCell}>
                  <span className={r.isBestPrice ? styles.retailerNameBest : styles.retailerName}>
                    {r.retailer}
                  </span>
                  {r.isBestPrice && (
                    <span className={styles.bestBadge} aria-label="Meilleur prix">
                      🏆 Meilleur prix
                    </span>
                  )}
                </div>
              </td>
              <td className={styles.priceCell}>{fmt(r.price)}</td>
              <td className={styles.diffCell}>
                <DiffBadge pct={r.diffPercent} />
              </td>
              <td className={styles.linkCell}>
                <a
                  href={r.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className={styles.externalLink}
                  aria-label={`Voir sur ${r.retailer}`}
                >
                  🔗
                </a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {data.maxSaving > 0 && (
        <div className={styles.footer}>
          <span>Économie max :</span>
          <span className={styles.savingAmount}>{fmt(data.maxSaving)}</span>
        </div>
      )}

      <p className={styles.disclaimer}>
        Prix simulés à titre indicatif. Cliquez sur un lien pour vérifier le prix réel.
      </p>
    </>
  )
}

export function PriceComparisonTable({ missionId, itemId, currency = 'EUR' }: PriceComparisonTableProps) {
  const { data, isLoading } = usePriceComparison(missionId, itemId)

  return (
    <div className={styles.wrapper}>
      <div className={styles.title}>🏪 Comparaison prix français</div>

      {isLoading ? (
        <div className={styles.loading}>
          <span className={styles.spinner} aria-hidden="true" />
          Chargement des prix...
        </div>
      ) : data ? (
        <ComparisonContent data={data} currency={currency} />
      ) : null}
    </div>
  )
}
