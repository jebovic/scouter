import { useState } from 'react'
import { useCreateInvite } from '../../hooks'
import styles from './InviteModal.module.css'

interface Props {
  missionId: string
  onClose: () => void
}

export function InviteModal({ missionId, onClose }: Props) {
  const [email, setEmail] = useState('')
  const [role, setRole] = useState<'viewer' | 'editor' | 'voter'>('voter')
  const [inviteUrl, setInviteUrl] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const { mutate: createInvite, isPending } = useCreateInvite(missionId)

  function handleCreate() {
    if (!email.trim()) return
    createInvite(
      { email: email.trim(), role },
      { onSuccess: (data) => setInviteUrl(data.inviteUrl) },
    )
  }

  function handleCopy() {
    if (!inviteUrl) return
    navigator.clipboard.writeText(inviteUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>INVITE COLLABORATOR</h2>
          <button className={styles.closeBtn} onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>

        {!inviteUrl ? (
          <div className={styles.form}>
            <label className={styles.label}>Email</label>
            <input
              type="email"
              className={styles.input}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="collaborator@example.com"
              onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            />

            <label className={styles.label}>Role</label>
            <div className={styles.roleOptions}>
              {(['viewer', 'editor', 'voter'] as const).map((r) => (
                <button
                  key={r}
                  className={`${styles.roleBtn} ${role === r ? styles.roleBtnActive : ''}`}
                  onClick={() => setRole(r)}
                >
                  {r}
                </button>
              ))}
            </div>

            <button
              className={styles.createBtn}
              onClick={handleCreate}
              disabled={!email.trim() || isPending}
            >
              {isPending ? 'Creating...' : 'Create Invite Link'}
            </button>
          </div>
        ) : (
          <div className={styles.result}>
            <p className={styles.successMsg}>Invite link created!</p>
            <div className={styles.linkBox}>
              <span className={styles.linkText}>{inviteUrl}</span>
              <button className={styles.copyBtn} onClick={handleCopy}>
                {copied ? '✓ Copied' : 'Copy'}
              </button>
            </div>
            <button className={styles.doneBtn} onClick={onClose}>
              Done
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
