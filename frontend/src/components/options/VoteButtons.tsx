import { useVotes, useRecordVote } from '../../hooks'
import styles from './VoteButtons.module.css'

const COLLABORATOR_KEY = 'scouter_collaborator_id'

interface Props {
  optionId: string
}

export function VoteButtons({ optionId }: Props) {
  const collaboratorId = localStorage.getItem(COLLABORATOR_KEY)
  const { data: votes } = useVotes(optionId)
  const { mutate: recordVote, isPending } = useRecordVote(optionId)

  if (!collaboratorId) return null

  const myVote = votes?.byCollaborator[collaboratorId] ?? 0

  function handleVote(v: number) {
    if (!collaboratorId || isPending) return
    const newVote = myVote === v ? 0 : v
    recordVote({ collaboratorId, vote: newVote })
  }

  const score = votes?.score ?? 0

  return (
    <div className={styles.container}>
      <button
        className={`${styles.btn} ${myVote === 1 ? styles.upActive : ''}`}
        onClick={() => handleVote(1)}
        disabled={isPending}
        aria-label="Upvote"
      >
        ↑
      </button>
      <span
        className={`${styles.score} ${score > 0 ? styles.positive : score < 0 ? styles.negative : ''}`}
      >
        {score > 0 ? `+${score}` : score}
      </span>
      <button
        className={`${styles.btn} ${myVote === -1 ? styles.downActive : ''}`}
        onClick={() => handleVote(-1)}
        disabled={isPending}
        aria-label="Downvote"
      >
        ↓
      </button>
    </div>
  )
}
