import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OptionNote } from './OptionNote'

const OPTION_ID = 'test-option-456'

describe('OptionNote', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('renders collapsed state with add note button', () => {
    render(<OptionNote optionId={OPTION_ID} />)
    expect(screen.getByRole('button', { name: /ajouter une note/i })).toBeInTheDocument()
  })

  it('shows note preview when note exists', () => {
    const noteContent = 'This is a short note'
    const noteData = { optionId: OPTION_ID, content: noteContent, updatedAt: new Date().toISOString() }
    localStorage.setItem(`scouter_note_${OPTION_ID}`, JSON.stringify(noteData))

    render(<OptionNote optionId={OPTION_ID} />)
    expect(screen.getByText(noteContent)).toBeInTheDocument()
  })

  it('truncates long note preview to 60 chars', () => {
    const longNote = 'A'.repeat(100)
    const noteData = { optionId: OPTION_ID, content: longNote, updatedAt: new Date().toISOString() }
    localStorage.setItem(`scouter_note_${OPTION_ID}`, JSON.stringify(noteData))

    render(<OptionNote optionId={OPTION_ID} />)
    const preview = screen.getByText(/^A+\.\.\.$/)
    expect(preview.textContent).toHaveLength(63) // 60 + 3 dots
  })

  it('expands to show textarea on button click', async () => {
    const user = userEvent.setup()
    render(<OptionNote optionId={OPTION_ID} />)

    const button = screen.getByRole('button', { name: /ajouter une note/i })
    await user.click(button)

    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('shows character count 0/500 when empty', async () => {
    const user = userEvent.setup()
    render(<OptionNote optionId={OPTION_ID} />)

    const button = screen.getByRole('button', { name: /ajouter une note/i })
    await user.click(button)

    expect(screen.getByText(/0\/500/)).toBeInTheDocument()
  })

  it('updates character count as user types', async () => {
    const user = userEvent.setup()
    render(<OptionNote optionId={OPTION_ID} />)

    const button = screen.getByRole('button', { name: /ajouter une note/i })
    await user.click(button)

    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    await user.type(textarea, 'Hello')

    expect(screen.getByText(/5\/500/)).toBeInTheDocument()
  })

  it('auto-saves on textarea blur', async () => {
    const user = userEvent.setup()
    render(<OptionNote optionId={OPTION_ID} />)

    const button = screen.getByRole('button', { name: /ajouter une note/i })
    await user.click(button)

    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    await user.type(textarea, 'Auto-save test')
    await user.click(document.body) // blur

    await waitFor(() => {
      const stored = localStorage.getItem(`scouter_note_${OPTION_ID}`)
      expect(stored).toBeTruthy()
      const parsed = JSON.parse(stored!)
      expect(parsed.content).toBe('Auto-save test')
    })
  })

  it('shows clear button in expanded state', async () => {
    const user = userEvent.setup()
    render(<OptionNote optionId={OPTION_ID} />)

    const button = screen.getByRole('button', { name: /ajouter une note/i })
    await user.click(button)

    expect(screen.getByRole('button', { name: /effacer la note/i })).toBeInTheDocument()
  })

  it('clears note when clear button clicked', async () => {
    const user = userEvent.setup()
    const noteData = { optionId: OPTION_ID, content: 'Test note', updatedAt: new Date().toISOString() }
    localStorage.setItem(`scouter_note_${OPTION_ID}`, JSON.stringify(noteData))

    render(<OptionNote optionId={OPTION_ID} />)

    // Expand first by clicking on the button with the note
    const editButton = screen.getByRole('button', { name: /modifier une note/i })
    fireEvent.click(editButton)

    // Verify textarea is visible
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    // Find and click the clear button
    const clearButton = screen.getByRole('button', { name: /effacer la note/i })
    fireEvent.click(clearButton)

    // Verify the note was cleared from localStorage
    await waitFor(() => {
      expect(localStorage.getItem(`scouter_note_${OPTION_ID}`)).toBeNull()
    })
  })
})
