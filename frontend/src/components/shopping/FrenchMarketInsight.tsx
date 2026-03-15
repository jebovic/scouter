import { useState } from 'react'
import { useFrenchMarket } from '../../hooks/useFrenchMarket'
import styles from './FrenchMarketInsight.module.css'

interface FrenchMarketInsightProps {
  itemName: string
  category?: string
}

function trendArrow(trend: string): string {
  const lower = trend.toLowerCase()
  if (lower.includes('hausse')) return '↑'
  if (lower.includes('baisse')) return '↓'
  return '→'
}

export function FrenchMarketInsight({ itemName, category = '' }: FrenchMarketInsightProps) {
  const [open, setOpen] = useState(false)
  const { data, isLoading, isError } = useFrenchMarket(itemName, category)

  if (isLoading) return null
  if (isError || !data) return null

  const topRetailers = data.bestRetailers.slice(0, 3)
  const topTips = data.tipsForFrenchBuyers.slice(0, 2)

  return (
    <div className={styles.wrap}>
      <button
        className={styles.toggle}
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label={open ? 'Masquer le marché français' : 'Afficher les infos marché français'}
      >
        🇫🇷 Marché FR
      </button>

      {open && (
        <div className={styles.card}>
          {/* Best month */}
          <div className={styles.row}>
            <span className={styles.icon}>📅</span>
            <span>
              <span className={styles.label}>Meilleur mois :</span>
              <span className={styles.value}>{data.bestMonthToBuy}</span>
            </span>
          </div>

          {/* Market trend */}
          <div className={styles.row}>
            <span className={styles.icon}>📈</span>
            <span>
              <span className={styles.label}>Tendance :</span>
              <span className={`${styles.trend}`}>
                {trendArrow(data.marketTrend)} {data.marketTrend}
              </span>
            </span>
          </div>

          {/* Top retailers */}
          {topRetailers.length > 0 && (
            <div className={styles.row}>
              <span className={styles.icon}>🛒</span>
              <div>
                <span className={styles.label}>Meilleurs sites :</span>
                <div className={styles.retailers}>
                  {topRetailers.map((r) => (
                    <a
                      key={r.name}
                      href={r.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={styles.chip}
                      title={`${r.name} — remise typique : ${r.typicalDiscount}`}
                    >
                      {r.name}
                    </a>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* French buyer tips */}
          {topTips.length > 0 && (
            <div className={styles.row}>
              <span className={styles.icon}>💡</span>
              <ul className={styles.tips}>
                {topTips.map((tip, i) => (
                  <li key={i}>{tip}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
