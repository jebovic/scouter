import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { VoiceInputButton } from './VoiceInputButton'

// Mock the useSpeechRecognition hook
vi.mock('../../hooks/useSpeechRecognition', () => ({
  useSpeechRecognition: vi.fn(),
}))

import { useSpeechRecognition } from '../../hooks/useSpeechRecognition'
const mockUseSpeechRecognition = vi.mocked(useSpeechRecognition)

describe('VoiceInputButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('when supported', () => {
    beforeEach(() => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
    })

    it('renders a microphone button when supported', () => {
      render(<VoiceInputButton onTranscript={() => {}} />)
      const button = screen.getByRole('button')
      expect(button).toBeTruthy()
      expect(button.textContent).toContain('🎤')
    })

    it('shows microphone icon when not listening', () => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      render(<VoiceInputButton onTranscript={() => {}} />)
      const button = screen.getByRole('button')
      expect(button.textContent).toContain('🎤')
    })

    it('shows stop icon when listening', () => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: true,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      render(<VoiceInputButton onTranscript={() => {}} />)
      const button = screen.getByRole('button')
      expect(button.textContent).toContain('⏹')
    })

    it('calls startListening when button clicked and not listening', () => {
      const startListening = vi.fn()
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: null,
        isSupported: true,
        startListening,
        stopListening: vi.fn(),
      })
      render(<VoiceInputButton onTranscript={() => {}} />)
      const button = screen.getByRole('button')
      fireEvent.click(button)
      expect(startListening).toHaveBeenCalled()
    })

    it('calls stopListening when button clicked and listening', () => {
      const stopListening = vi.fn()
      mockUseSpeechRecognition.mockReturnValue({
        isListening: true,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening,
      })
      render(<VoiceInputButton onTranscript={() => {}} />)
      const button = screen.getByRole('button')
      fireEvent.click(button)
      expect(stopListening).toHaveBeenCalled()
    })

    it('calls onTranscript when transcript updates', () => {
      const onTranscript = vi.fn()
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: 'Buy a laptop',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      render(<VoiceInputButton onTranscript={onTranscript} />)
      expect(onTranscript).toHaveBeenCalledWith('Buy a laptop')
    })

    it('does not show error message when no error', () => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      render(<VoiceInputButton onTranscript={() => {}} />)
      expect(screen.queryByText(/error/i)).toBeFalsy()
    })

    it('shows error message when error occurs', () => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: 'network',
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      render(<VoiceInputButton onTranscript={() => {}} />)
      const errorMsg = screen.queryByText(/error|erreur/i)
      expect(errorMsg).toBeTruthy()
    })

    it('applies listening class when listening', () => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: true,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      const { container } = render(<VoiceInputButton onTranscript={() => {}} />)
      const button = container.querySelector('button')
      expect(button?.className).toContain('listening')
    })

    it('does not apply listening class when not listening', () => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: null,
        isSupported: true,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
      const { container } = render(<VoiceInputButton onTranscript={() => {}} />)
      const button = container.querySelector('button')
      expect(button?.className).not.toContain('listening')
    })
  })

  describe('when not supported', () => {
    beforeEach(() => {
      mockUseSpeechRecognition.mockReturnValue({
        isListening: false,
        transcript: '',
        error: null,
        isSupported: false,
        startListening: vi.fn(),
        stopListening: vi.fn(),
      })
    })

    it('does not render button when not supported', () => {
      render(<VoiceInputButton onTranscript={() => {}} />)
      const button = screen.queryByRole('button')
      expect(button).toBeFalsy()
    })

    it('shows not supported message when not supported', () => {
      render(<VoiceInputButton onTranscript={() => {}} />)
      const message = screen.getByText(/navigateur|support/i)
      expect(message).toBeTruthy()
    })

    it('applies dim styling to unsupported message', () => {
      const { container } = render(<VoiceInputButton onTranscript={() => {}} />)
      const message = container.querySelector('.unsupported')
      expect(message).toBeTruthy()
    })
  })
})
