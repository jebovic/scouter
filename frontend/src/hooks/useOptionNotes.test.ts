import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useOptionNotes } from './useOptionNotes'

const OPTION_ID = 'test-option-123'

describe('useOptionNotes', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns empty note when no note exists', () => {
    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    expect(result.current.note).toBe('')
    expect(result.current.hasNote).toBe(false)
  })

  it('loads note from localStorage on mount', () => {
    const noteContent = 'This is a test note'
    const noteData = { optionId: OPTION_ID, content: noteContent, updatedAt: new Date().toISOString() }
    localStorage.setItem(`scouter_note_${OPTION_ID}`, JSON.stringify(noteData))

    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    expect(result.current.note).toBe(noteContent)
    expect(result.current.hasNote).toBe(true)
  })

  it('save() persists note to localStorage', () => {
    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    const noteContent = 'Saving a note'

    act(() => {
      result.current.save(noteContent)
    })

    const stored = localStorage.getItem(`scouter_note_${OPTION_ID}`)
    expect(stored).toBeTruthy()
    const parsed = JSON.parse(stored!)
    expect(parsed.content).toBe(noteContent)
    expect(parsed.optionId).toBe(OPTION_ID)
  })

  it('save() updates the note in state immediately', () => {
    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    const noteContent = 'Updated note'

    act(() => {
      result.current.save(noteContent)
    })

    expect(result.current.note).toBe(noteContent)
  })

  it('save() sets updatedAt timestamp', () => {
    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    const before = new Date().getTime()

    act(() => {
      result.current.save('Note with timestamp')
    })

    const after = new Date().getTime()
    const stored = localStorage.getItem(`scouter_note_${OPTION_ID}`)
    const parsed = JSON.parse(stored!)
    const timestamp = new Date(parsed.updatedAt).getTime()

    expect(timestamp).toBeGreaterThanOrEqual(before)
    expect(timestamp).toBeLessThanOrEqual(after)
  })

  it('clear() removes note from localStorage', () => {
    const noteData = { optionId: OPTION_ID, content: 'Test', updatedAt: new Date().toISOString() }
    localStorage.setItem(`scouter_note_${OPTION_ID}`, JSON.stringify(noteData))

    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    expect(result.current.hasNote).toBe(true)

    act(() => {
      result.current.clear()
    })

    expect(localStorage.getItem(`scouter_note_${OPTION_ID}`)).toBeNull()
    expect(result.current.note).toBe('')
    expect(result.current.hasNote).toBe(false)
  })

  it('different optionIds have isolated notes', () => {
    const OPTION_A = 'option-a'
    const OPTION_B = 'option-b'

    const { result: resultA } = renderHook(() => useOptionNotes(OPTION_A))
    const { result: resultB } = renderHook(() => useOptionNotes(OPTION_B))

    act(() => {
      resultA.current.save('Note A')
    })

    act(() => {
      resultB.current.save('Note B')
    })

    expect(resultA.current.note).toBe('Note A')
    expect(resultB.current.note).toBe('Note B')
  })

  it('handles localStorage.setItem errors gracefully', () => {
    const spy = vi.spyOn(Storage.prototype, 'setItem').mockImplementationOnce(() => {
      throw new Error('QuotaExceededError')
    })

    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    expect(() => {
      act(() => {
        result.current.save('This will fail')
      })
    }).not.toThrow()

    spy.mockRestore()
  })

  it('handles localStorage.getItem errors gracefully', () => {
    const spy = vi.spyOn(Storage.prototype, 'getItem').mockImplementationOnce(() => {
      throw new Error('QuotaExceededError')
    })

    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    expect(result.current.note).toBe('')
    expect(result.current.hasNote).toBe(false)

    spy.mockRestore()
  })

  it('save() updates hasNote flag', () => {
    const { result } = renderHook(() => useOptionNotes(OPTION_ID))
    expect(result.current.hasNote).toBe(false)

    act(() => {
      result.current.save('Adding a note')
    })

    expect(result.current.hasNote).toBe(true)
  })
})
