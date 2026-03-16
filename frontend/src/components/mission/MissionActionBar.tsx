import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import styles from './MissionActionBar.module.css'

interface Mission {
  id: string
  slug: string
  name: string
}

interface MissionActionBarProps {
  mission: Mission
  onEdit: () => void
  onArchive: () => void
  onDelete: () => void
}

type Dialog = 'none' | 'archive' | 'delete'

export function MissionActionBar({ mission, onEdit, onArchive, onDelete }: MissionActionBarProps) {
  const { t } = useTranslation()
  const [dialog, setDialog] = useState<Dialog>('none')
  const [deleteConfirmName, setDeleteConfirmName] = useState('')

  function handleArchiveConfirm() {
    onArchive()
    setDialog('none')
  }

  function handleDeleteConfirm() {
    if (deleteConfirmName !== mission.name) return
    onDelete()
    setDialog('none')
    setDeleteConfirmName('')
  }

  function handleClose() {
    setDialog('none')
    setDeleteConfirmName('')
  }

  return (
    <>
      <div className={styles.bar}>
        <button className={styles.editBtn} onClick={onEdit}>
          {t('mission.actions.edit')}
        </button>
        <button className={styles.archiveBtn} onClick={() => setDialog('archive')}>
          {t('mission.actions.archive')}
        </button>
        <button className={styles.deleteBtn} onClick={() => setDialog('delete')}>
          {t('mission.actions.delete')}
        </button>
      </div>

      {dialog === 'archive' && (
        <div className={styles.overlay} onClick={handleClose}>
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <p>{t('mission.actions.archiveConfirm')}</p>
            <div className={styles.dialogActions}>
              <button onClick={handleClose}>{t('common.cancel')}</button>
              <button className={styles.confirmBtn} onClick={handleArchiveConfirm}>
                {t('common.confirm')}
              </button>
            </div>
          </div>
        </div>
      )}

      {dialog === 'delete' && (
        <div className={styles.overlay} onClick={handleClose}>
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <p>{t('mission.actions.deleteConfirm')}</p>
            <p className={styles.deleteHint}>{t('mission.actions.deleteTypeToConfirm')}</p>
            <input
              className={styles.deleteInput}
              placeholder={mission.name}
              value={deleteConfirmName}
              onChange={(e) => setDeleteConfirmName(e.target.value)}
              autoFocus
            />
            <div className={styles.dialogActions}>
              <button onClick={handleClose}>{t('common.cancel')}</button>
              <button
                aria-label="confirm delete"
                className={styles.deleteConfirmBtn}
                onClick={handleDeleteConfirm}
                disabled={deleteConfirmName !== mission.name}
              >
                {t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
