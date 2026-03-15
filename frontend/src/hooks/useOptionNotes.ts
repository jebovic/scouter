import { useState, useCallback, useEffect } from 'react'

export interface OptionNote {
  optionId: string
  content: string
  updatedAt: string // ISO string
}

interface UseOptionNotesResult {
  note: string
  save: (content: string) => void
  clear: () => void
  hasNote: boolean
}

export function useOptionNotes(optionId: string): UseOptionNotesResult {
  const [note, setNote] = useState<string>('')
  const [hasNote, setHasNote] = useState<boolean>(false)

  // Load note from localStorage on mount
  useEffect(() => {
    try {
      const storageKey = `scouter_note_${optionId}`
      const stored = localStorage.getItem(storageKey)
      if (stored) {
        const parsed = JSON.parse(stored) as OptionNote
        setNote(parsed.content)
        setHasNote(true)
      } else {
        setNote('')
        setHasNote(false)
      }
    } catch {
      // Gracefully handle localStorage errors
      setNote('')
      setHasNote(false)
    }
  }, [optionId])

  const save = useCallback(
    (content: string) => {
      try {
        const storageKey = `scouter_note_${optionId}`
        const noteData: OptionNote = {
          optionId,
          content,
          updatedAt: new Date().toISOString(),
        }
        localStorage.setItem(storageKey, JSON.stringify(noteData))
        setNote(content)
        setHasNote(content.length > 0)
      } catch {
        // Gracefully handle localStorage errors
        console.error('Failed to save note to localStorage')
      }
    },
    [optionId],
  )

  const clear = useCallback(() => {
    try {
      const storageKey = `scouter_note_${optionId}`
      localStorage.removeItem(storageKey)
      setNote('')
      setHasNote(false)
    } catch {
      // Gracefully handle localStorage errors
      console.error('Failed to clear note from localStorage')
    }
  }, [optionId])

  return {
    note,
    save,
    clear,
    hasNote,
  }
}
