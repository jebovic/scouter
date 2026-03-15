import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useGetInvite, useJoinMission } from '../hooks'
import { LoadingPulse } from '../components/scouter'
import styles from './JoinPage.module.css'

const COLLABORATOR_KEY = 'scouter_collaborator_id'

export default function JoinPage() {
  const { token } = useParams<{ token: string }>()
  const navigate = useNavigate()
  const [displayName, setDisplayName] = useState('')

  const { data: invite, isLoading, error } = useGetInvite(token!)
  const { mutate: join, isPending } = useJoinMission(token!)

  function handleJoin() {
    if (!displayName.trim()) return
    join(displayName.trim(), {
      onSuccess: (collaborator) => {
        localStorage.setItem(COLLABORATOR_KEY, collaborator.id)
        navigate(`/missions/${collaborator.missionId}`)
      },
    })
  }

  if (isLoading) {
    return (
      <main className={styles.main}>
        <LoadingPulse label="Loading invite..." />
      </main>
    )
  }

  if (error || !invite) {
    return (
      <main className={styles.main}>
        <div className={styles.errorCard}>
          <h1 className={styles.errorTitle}>Invalid Invite</h1>
          <p className={styles.errorMsg}>This invite link is invalid or has expired.</p>
          <button onClick={() => navigate('/')} className={styles.backBtn}>
            Go Home
          </button>
        </div>
      </main>
    )
  }

  return (
    <main className={styles.main}>
      <div className={styles.card}>
        <div className={styles.icon}>🤝</div>
        <h1 className={styles.title}>You've been invited</h1>
        <p className={styles.subtitle}>
          Join as <strong className={styles.role}>{invite.role}</strong>
        </p>
        <div className={styles.form}>
          <label className={styles.label}>Your display name</label>
          <input
            className={styles.input}
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="e.g. Alice"
            onKeyDown={(e) => e.key === 'Enter' && handleJoin()}
            autoFocus
          />
          <button
            className={styles.joinBtn}
            onClick={handleJoin}
            disabled={!displayName.trim() || isPending}
          >
            {isPending ? 'Joining...' : 'Join Mission'}
          </button>
        </div>
      </div>
    </main>
  )
}
