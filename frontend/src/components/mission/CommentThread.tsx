import { useState } from 'react'
import { useComments, useCreateComment, useDeleteComment } from '../../hooks/useComments'
import styles from './CommentThread.module.css'

interface CommentThreadProps {
  missionSlug: string
  optionId?: string
  compact?: boolean
}

const COMPACT_LIMIT = 3

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diff / 60_000)
  if (minutes < 1) return "il y a quelques secondes"
  if (minutes < 60) return `il y a ${minutes} min`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `il y a ${hours} h`
  const days = Math.floor(hours / 24)
  return `il y a ${days} j`
}

export function CommentThread({ missionSlug, optionId, compact = false }: CommentThreadProps) {
  const { data: comments = [], isLoading } = useComments(missionSlug, optionId)
  const { mutate: create, isPending: creating } = useCreateComment(missionSlug)
  const { mutate: remove } = useDeleteComment(missionSlug)

  const [author, setAuthor] = useState('Vous')
  const [body, setBody] = useState('')
  const [expanded, setExpanded] = useState(false)

  const displayedComments =
    compact && !expanded && comments.length > COMPACT_LIMIT
      ? comments.slice(0, COMPACT_LIMIT)
      : comments

  const hiddenCount = comments.length - COMPACT_LIMIT

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmedBody = body.trim()
    const trimmedAuthor = author.trim() || 'Vous'
    if (!trimmedBody || trimmedBody.length > 1000) return
    create(
      { optionId, author: trimmedAuthor, body: trimmedBody },
      {
        onSuccess: () => {
          setBody('')
        },
      },
    )
  }

  return (
    <div className={styles.thread}>
      <h4 className={styles.title}>Commentaires</h4>

      <div className={styles.list}>
        {isLoading && (
          <p className={styles.empty}>Chargement...</p>
        )}

        {!isLoading && comments.length === 0 && (
          <p className={styles.empty}>Aucun commentaire pour l&apos;instant</p>
        )}

        {displayedComments.map((c) => (
          <div key={c.id} className={styles.comment}>
            <div className={styles.commentMeta}>
              <span className={styles.author}>{c.author}</span>
              <span className={styles.time}>{timeAgo(c.createdAt)}</span>
            </div>
            <p className={styles.body}>{c.body}</p>
            <button
              className={styles.deleteBtn}
              onClick={() => remove(c.id)}
              aria-label="Supprimer ce commentaire"
              title="Supprimer"
            >
              ✕
            </button>
          </div>
        ))}

        {compact && comments.length > COMPACT_LIMIT && !expanded && (
          <button className={styles.toggleBtn} onClick={() => setExpanded(true)}>
            Voir {hiddenCount} commentaire{hiddenCount > 1 ? 's' : ''}
          </button>
        )}

        {compact && expanded && comments.length > COMPACT_LIMIT && (
          <button className={styles.toggleBtn} onClick={() => setExpanded(false)}>
            Réduire
          </button>
        )}
      </div>

      <form className={styles.form} onSubmit={handleSubmit}>
        <div className={styles.formRow}>
          <input
            className={styles.authorInput}
            type="text"
            value={author}
            onChange={(e) => setAuthor(e.target.value)}
            placeholder="Votre nom"
            maxLength={50}
            aria-label="Nom de l'auteur"
          />
          <textarea
            className={styles.textarea}
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="Écrire un commentaire..."
            maxLength={1000}
            aria-label="Corps du commentaire"
          />
        </div>
        <div className={styles.formFooter}>
          <span className={`${styles.charCount} ${body.length > 900 ? styles.charCountWarn : ''}`}>
            {body.length} / 1000
          </span>
          <button
            type="submit"
            className={styles.submitBtn}
            disabled={creating || !body.trim() || body.length > 1000}
          >
            Envoyer
          </button>
        </div>
      </form>
    </div>
  )
}
