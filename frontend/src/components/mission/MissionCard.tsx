import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Card, StatusBadge, BudgetBar } from '../scouter'
import { CategoryBadge } from './CategoryBadge'
import { useSwipeGesture } from '../../hooks/useSwipeGesture'
import { useDuplicateMission, useCloneMission, useArchiveMission, useUnarchiveMission, useDeleteMission } from '../../hooks/useMission'
import type { Mission, ShoppingItem } from '../../types'
import styles from './MissionCard.module.css'

interface MissionCardProps {
  mission: Mission
  items?: ShoppingItem[]
  onArchive?: (missionId: string) => void
}

export function MissionCard({ mission, items = [], onArchive }: MissionCardProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { duplicateMission, isPending: isDuplicating } = useDuplicateMission()
  const { cloneMission, isPending: isCloning } = useCloneMission()
  const { archiveMission } = useArchiveMission()
  const { unarchiveMission } = useUnarchiveMission()
  const { deleteMission } = useDeleteMission()
  const [menuOpen, setMenuOpen] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [deleteConfirmName, setDeleteConfirmName] = useState('')
  const isArchived = !!mission.archivedAt
  const spent = items.reduce((sum, item) => sum + item.price, 0)

  useEffect(() => {
    if (!menuOpen) return
    function handleOutside(e: MouseEvent) {
      setMenuOpen(false)
    }
    document.addEventListener('mousedown', handleOutside)
    return () => document.removeEventListener('mousedown', handleOutside)
  }, [menuOpen])

  const { swipeX, handlers } = useSwipeGesture({
    threshold: 80,
    onSwipeLeft: () => onArchive?.(mission.id),
    onSwipeRight: () => navigate(`/missions/${mission.slug}?research=1`),
  })

  const isSwipingLeft = swipeX < -20
  const isSwipingRight = swipeX > 20

  return (
    <div
      className={styles.swipeWrapper}
      {...handlers}
    >
      {/* Archive action hint (revealed on swipe-left) */}
      <div
        className={`${styles.swipeAction} ${styles.swipeActionLeft} ${isSwipingLeft ? styles.swipeActionVisible : ''}`}
        aria-hidden="true"
      >
        <span className={styles.swipeActionIcon}>📦</span>
        <span className={styles.swipeActionLabel}>Archive</span>
      </div>

      {/* Research action hint (revealed on swipe-right) */}
      <div
        className={`${styles.swipeAction} ${styles.swipeActionRight} ${isSwipingRight ? styles.swipeActionVisible : ''}`}
        aria-hidden="true"
      >
        <span className={styles.swipeActionIcon}>⚡</span>
        <span className={styles.swipeActionLabel}>Research</span>
      </div>

      <div
        className={styles.swipeCard}
        style={{ '--swipe-x': `${swipeX}px` } as React.CSSProperties}
      >
        <Card
          onClick={() => navigate(`/missions/${mission.slug}`)}
          className={`card-enter ${styles.card}`}
        >
          <div className={styles.header}>
            <div className={styles.identity}>
              <span className={styles.icon}>{mission.icon}</span>
              <div>
                <div className={styles.name}>{mission.name}</div>
                <div className={styles.category}>{mission.category}</div>
                {mission.autotagCategory && (
                  <div className={styles.autotagBadge}>
                    <CategoryBadge category={mission.autotagCategory} size="sm" />
                  </div>
                )}
              </div>
            </div>
            <div className={styles.headerActions}>
              {isArchived && (
                <span className={styles.archivedBadge}>{t('mission.actions.archivedBadge')}</span>
              )}
              <StatusBadge phase={mission.phase} />
              <button
                className={styles.duplicateBtn}
                title="Duplicate mission"
                aria-label="Duplicate mission"
                disabled={isDuplicating}
                onClick={async (e) => {
                  e.stopPropagation()
                  const copy = await duplicateMission(mission.slug)
                  navigate(`/missions/${copy.slug}`)
                }}
              >
                {isDuplicating ? '…' : '📋'}
              </button>
              <button
                className={styles.cloneBtn}
                title="Dupliquer avec options et achats"
                aria-label="Dupliquer la mission avec tous les éléments"
                disabled={isCloning}
                onClick={async (e) => {
                  e.stopPropagation()
                  try {
                    const cloned = await cloneMission(mission.slug)
                    navigate(`/missions/${cloned.slug}`)
                  } catch {
                    // error toast handled by useCloneMission onError
                  }
                }}
              >
                {isCloning ? 'Copie en cours...' : 'Dupliquer'}
              </button>
              <div className={styles.menuContainer}>
                <button
                  className={styles.menuBtn}
                  aria-label="more options"
                  onClick={(e) => { e.stopPropagation(); setMenuOpen((o) => !o) }}
                >
                  ⋯
                </button>
                {menuOpen && (
                  <div className={styles.menu} onClick={(e) => e.stopPropagation()}>
                    {isArchived ? (
                      <button onClick={() => { unarchiveMission(mission.id); setMenuOpen(false) }}>
                        {t('mission.actions.unarchive')}
                      </button>
                    ) : (
                      <button onClick={() => { archiveMission(mission.id); setMenuOpen(false) }}>
                        {t('mission.actions.archive')}
                      </button>
                    )}
                    <button
                      className={styles.menuDeleteBtn}
                      onClick={() => { setConfirmDelete(true); setMenuOpen(false) }}
                    >
                      {t('common.delete')}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          <BudgetBar spent={spent} budget={mission.budget} currency={mission.currency} />

          {mission.constraints.length > 0 && (
            <div className={styles.constraintList}>
              {mission.constraints.slice(0, 3).map((c) => (
                <span
                  key={c.key}
                  className={`${styles.constraintTag} ${c.type === 'hard' ? styles.constraintHard : styles.constraintSoft}`}
                >
                  {c.label}
                </span>
              ))}
              {mission.constraints.length > 3 && (
                <span className={styles.constraintOverflow}>
                  +{mission.constraints.length - 3}
                </span>
              )}
            </div>
          )}
        </Card>
      </div>

      {confirmDelete && (
        <div className={styles.overlay} onClick={() => setConfirmDelete(false)}>
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <p>{t('mission.actions.deleteConfirm')}</p>
            <p className={styles.deleteHint}>{t('mission.actions.deleteTypeToConfirm')}</p>
            <input
              placeholder={mission.name}
              value={deleteConfirmName}
              onChange={(e) => setDeleteConfirmName(e.target.value)}
              autoFocus
            />
            <div className={styles.dialogActions}>
              <button onClick={() => { setConfirmDelete(false); setDeleteConfirmName('') }}>{t('common.cancel')}</button>
              <button
                disabled={deleteConfirmName !== mission.name}
                onClick={() => {
                  deleteMission(mission.slug)
                  setConfirmDelete(false)
                  setDeleteConfirmName('')
                }}
              >
                {t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
