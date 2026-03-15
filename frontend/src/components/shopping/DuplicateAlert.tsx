import { useState } from 'react'
import { useDuplicates } from '../../hooks/useDuplicates'
import type { DuplicateGroup } from '../../api/duplicatedetector'
import styles from './DuplicateAlert.module.css'

interface Props {
  missionId: string
}

function GroupRow({
  group,
  onDismiss,
}: {
  group: DuplicateGroup
  onDismiss: () => void
}) {
  const pct = Math.round(group.similarity * 100)
  return (
    <div className={styles.group}>
      <div className={styles.groupHeader}>
        <span className={styles.groupReason}>{group.reason}</span>
        <span className={styles.groupSim}>{pct}% similaire</span>
        <button
          className={styles.dismissGroupBtn}
          onClick={onDismiss}
          aria-label="Ignorer ce doublon"
        >
          ✕
        </button>
      </div>
      <div className={styles.itemGrid}>
        {group.items.map((item) => {
          const isRecommended = item.itemId === group.recommended
          return (
            <div
              key={item.itemId}
              className={`${styles.itemCard} ${isRecommended ? styles.itemCardRecommended : ''}`}
            >
              {isRecommended && (
                <span className={styles.keepBadge}>Conserver</span>
              )}
              <p className={styles.itemName}>{item.itemName}</p>
              <p className={styles.itemMeta}>
                {item.merchant && (
                  <span className={styles.merchant}>{item.merchant}</span>
                )}
                <span className={styles.price}>{item.price.toFixed(2)} €</span>
              </p>
              <span className={styles.statusBadge} data-status={item.status}>
                {item.status}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export function DuplicateAlert({ missionId }: Props) {
  const { report, isLoading } = useDuplicates(missionId)
  const [dismissed, setDismissed] = useState<Set<number>>(new Set())
  const [collapsed, setCollapsed] = useState(false)

  if (isLoading || !report || report.groups.length === 0) return null

  const visibleGroups = report.groups.filter((_, i) => !dismissed.has(i))
  if (visibleGroups.length === 0) return null

  const totalVisible = visibleGroups.reduce((sum, g) => sum + g.items.length, 0)

  return (
    <div className={styles.banner} role="alert">
      <div className={styles.bannerHeader}>
        <span className={styles.bannerTitle}>
          ⚠️ {totalVisible} doublons potentiels détectés
        </span>
        <button
          className={styles.toggleBtn}
          onClick={() => setCollapsed((c) => !c)}
          aria-expanded={!collapsed}
        >
          {collapsed ? 'Voir' : 'Masquer'}
        </button>
      </div>

      {!collapsed && (
        <div className={styles.groupList}>
          {report.groups.map((group, i) => {
            if (dismissed.has(i)) return null
            return (
              <GroupRow
                key={i}
                group={group}
                onDismiss={() =>
                  setDismissed((prev) => new Set([...prev, i]))
                }
              />
            )
          })}
        </div>
      )}
    </div>
  )
}
