import { useState } from 'react'
import { useListOptimizer } from '../../hooks/useListOptimizer'
import { LoadingPulse } from '../scouter'
import styles from './ListOptimizerPanel.module.css'

interface ListOptimizerPanelProps {
  missionId: string
}

export function ListOptimizerPanel({ missionId }: ListOptimizerPanelProps) {
  const { data, isLoading } = useListOptimizer(missionId)
  const [isCollapsed, setIsCollapsed] = useState(false)

  if (isLoading) {
    return (
      <div className={styles.container}>
        <LoadingPulse label="Optimisation en cours..." />
      </div>
    )
  }

  if (!data) {
    return null
  }

  return (
    <div className={styles.container}>
      <button
        className={styles.header}
        onClick={() => setIsCollapsed(!isCollapsed)}
        aria-expanded={!isCollapsed}
      >
        <span className={styles.title}>🛒 Liste Optimisée</span>
        <span className={styles.toggleIcon}>{isCollapsed ? '▶' : '▼'}</span>
      </button>

      {!isCollapsed && (
        <div className={styles.content}>
          <div className={styles.summary}>{data.summary}</div>

          <div className={styles.merchantGroups}>
            {data.merchantGroups.map((group) => (
              <div key={group.merchant} className={styles.merchantGroup}>
                <div className={styles.merchantHeader}>
                  <h4 className={styles.merchantName}>
                    {group.merchant.charAt(0).toUpperCase() + group.merchant.slice(1)}
                  </h4>
                  <div className={styles.merchantStats}>
                    <span className={styles.itemCount}>{group.items.length} article(s)</span>
                    <span className={styles.price}>{group.totalPrice.toFixed(2)}€</span>
                    {group.tripSaving > 0 && (
                      <span className={styles.saving}>Économies : {group.tripSaving.toFixed(2)}€</span>
                    )}
                  </div>
                </div>

                <div className={styles.itemsList}>
                  {group.items.map((item) => (
                    <div key={item.itemId} className={styles.itemRow}>
                      <span className={styles.priority}>#{item.priority}</span>
                      <div className={styles.itemDetails}>
                        <span className={styles.itemName}>{item.name}</span>
                        <span className={styles.itemStatus}>{item.status}</span>
                      </div>
                      <span className={styles.itemPrice}>{item.price.toFixed(2)}€</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>

          <div className={styles.footer}>
            <div className={styles.totalRow}>
              <span className={styles.label}>Total des articles</span>
              <span className={styles.value}>{data.totalItems}</span>
            </div>
            <div className={styles.totalRow}>
              <span className={styles.label}>Prix total</span>
              <span className={styles.value}>{data.totalPrice.toFixed(2)}€</span>
            </div>
            <div className={styles.totalRow}>
              <span className={styles.label}>Trajets estimés</span>
              <span className={styles.value}>{data.estimatedTrips}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
