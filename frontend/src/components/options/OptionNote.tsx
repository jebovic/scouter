import { useState, useCallback, useRef, useEffect } from 'react'
import { useOptionNotes } from '../../hooks/useOptionNotes'
import styles from './OptionNote.module.css'

interface OptionNoteProps {
  optionId: string
}

export function OptionNote({ optionId }: OptionNoteProps) {
  const { note, save, clear, hasNote } = useOptionNotes(optionId)
  const [isExpanded, setIsExpanded] = useState(false)
  const [draftContent, setDraftContent] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const isClickingClearRef = useRef(false)

  // Sync draft content with note when it changes
  useEffect(() => {
    setDraftContent(note)
  }, [note])

  const handleExpand = useCallback(() => {
    setDraftContent(note)
    setIsExpanded(true)
    // Focus textarea after state update
    setTimeout(() => {
      textareaRef.current?.focus()
    }, 0)
  }, [note])

  const handleCollapse = useCallback(() => {
    setIsExpanded(false)
  }, [])

  const handleBlur = useCallback(() => {
    // If we just clicked the clear button, don't auto-collapse
    if (isClickingClearRef.current) {
      isClickingClearRef.current = false
      return
    }
    if (draftContent !== note) {
      save(draftContent)
    }
    handleCollapse()
  }, [draftContent, note, save, handleCollapse])

  const handleClear = useCallback(() => {
    isClickingClearRef.current = true
    clear()
    setDraftContent('')
    handleCollapse()
  }, [clear, handleCollapse])

  const charCount = draftContent.length
  const truncatedPreview = hasNote && note.length > 60 ? note.slice(0, 60) + '...' : note

  return (
    <div className={styles.container}>
      {!isExpanded ? (
        <button
          className={styles.collapseButton}
          onClick={handleExpand}
          aria-label={hasNote ? 'Modifier une note' : 'Ajouter une note'}
        >
          {hasNote ? <span className={styles.notePreview}>{truncatedPreview}</span> : '✎ Ajouter une note'}
          {hasNote && <span className={styles.indicator} title="Note exists" />}
        </button>
      ) : (
        <div className={styles.expandedContainer}>
          <textarea
            ref={textareaRef}
            className={styles.textarea}
            value={draftContent}
            onChange={(e) => setDraftContent(e.currentTarget.value.slice(0, 500))}
            onBlur={handleBlur}
            placeholder="Entrez votre note ici..."
            rows={3}
          />
          <div className={styles.actionsRow}>
            <span className={styles.charCount}>
              {charCount}/500
            </span>
            <button
              className={styles.clearButton}
              onClick={handleClear}
              aria-label="Effacer la note"
            >
              Effacer
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
