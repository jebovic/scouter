import { useState } from 'react'
import { useCollaborators, useRemoveCollaborator } from '../../hooks'
import { InviteModal } from './InviteModal'
import styles from './CollaboratorsPanel.module.css'

interface Props {
  missionId: string
}

export function CollaboratorsPanel({ missionId }: Props) {
  const [showInvite, setShowInvite] = useState(false)
  const { data: collaborators = [], isLoading } = useCollaborators(missionId)
  const { mutate: remove } = useRemoveCollaborator(missionId)

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <h3 className={styles.title}>COLLABORATORS</h3>
        <button className={styles.inviteBtn} onClick={() => setShowInvite(true)}>
          + Invite
        </button>
      </div>

      {isLoading && <div className={styles.loading}>Loading...</div>}

      {!isLoading && collaborators.length === 0 && (
        <p className={styles.empty}>
          No collaborators yet. Invite someone to collaborate on this mission.
        </p>
      )}

      <div className={styles.list}>
        {collaborators.map((c) => (
          <div key={c.id} className={styles.collaborator}>
            <div className={styles.avatar}>
              {(c.displayName ?? c.email).charAt(0).toUpperCase()}
            </div>
            <div className={styles.info}>
              <span className={styles.name}>{c.displayName ?? c.email}</span>
              <span className={`${styles.role} ${styles[`role_${c.role}`]}`}>{c.role}</span>
              {!c.joinedAt && <span className={styles.pending}>pending</span>}
            </div>
            <button
              className={styles.removeBtn}
              onClick={() => remove(c.id)}
              aria-label={`Remove ${c.displayName ?? c.email}`}
            >
              ×
            </button>
          </div>
        ))}
      </div>

      {showInvite && (
        <InviteModal missionId={missionId} onClose={() => setShowInvite(false)} />
      )}
    </div>
  )
}
