import { useState, useEffect } from 'react'
import { useVotes, useRecordVote } from '../../hooks'
import styles from './VoteBar.module.css'

const COLLABORATOR_KEY = 'scouter_collaborator_id'

interface VoteBarProps {
  optionId: string
  onVote?: (vote: number) => void
}

export function VoteBar({ optionId, onVote }: VoteBarProps) {
  const collaboratorId = localStorage.getItem(COLLABORATOR_KEY)
  const { data: votes, isLoading } = useVotes(optionId)
  const { mutate: recordVote, isPending } = useRecordVote(optionId)
  const [updatingUpvote, setUpdatingUpvote] = useState(false)
  const [updatingDownvote, setUpdatingDownvote] = useState(false)

  // Extract user's current vote
  const userVote = collaboratorId ? (votes?.byCollaborator[collaboratorId] ?? null) : null

  // Animate count updates
  useEffect(() => {
    if (!isPending) {
      setUpdatingUpvote(false)
      setUpdatingDownvote(false)
    }
  }, [isPending])

  function handleVote(voteValue: number) {
    if (!collaboratorId || isPending) return

    // If clicking the same vote button, remove the vote
    const newVote = userVote === voteValue ? 0 : voteValue

    if (voteValue === 1) {
      setUpdatingUpvote(true)
    } else if (voteValue === -1) {
      setUpdatingDownvote(true)
    }

    recordVote(
      { collaboratorId, vote: newVote },
      {
        onSuccess: () => {
          onVote?.(newVote)
        },
      }
    )
  }

  // Don't show vote bar if no collaborator is logged in
  if (!collaboratorId) {
    return null
  }

  if (isLoading) {
    return (
      <div className={styles.container}>
        <div className={styles.label}>Votes…</div>
      </div>
    )
  }

  const upvotes = votes?.upvotes ?? 0
  const downvotes = votes?.downvotes ?? 0
  const score = (votes?.score ?? 0)

  return (
    <div className={styles.container}>
      <div className={styles.voteGroup}>
        <button
          onClick={() => handleVote(1)}
          className={`${styles.voteBtn} ${styles.upvoteBtn} ${userVote === 1 ? styles.active : ''}`}
          title={userVote === 1 ? 'Remove upvote' : 'Upvote this option'}
          aria-label={userVote === 1 ? 'Remove upvote' : 'Upvote this option'}
          disabled={isPending}
        >
          👍
          <span className={`${styles.count} ${updatingUpvote ? styles.updating : ''}`}>
            {upvotes}
          </span>
        </button>
        <button
          onClick={() => handleVote(-1)}
          className={`${styles.voteBtn} ${styles.downvoteBtn} ${userVote === -1 ? styles.active : ''}`}
          title={userVote === -1 ? 'Remove downvote' : 'Downvote this option'}
          aria-label={userVote === -1 ? 'Remove downvote' : 'Downvote this option'}
          disabled={isPending}
        >
          👎
          <span className={`${styles.count} ${updatingDownvote ? styles.updating : ''}`}>
            {downvotes}
          </span>
        </button>
      </div>
      {score !== 0 && (
        <div className={styles.score} title={`Net vote score: ${score}`}>
          {score > 0 ? '+' : ''}{score}
        </div>
      )}
      <div className={styles.label}>votes</div>
    </div>
  )
}
