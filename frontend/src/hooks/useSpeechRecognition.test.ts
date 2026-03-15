import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useSpeechRecognition } from './useSpeechRecognition'

describe('useSpeechRecognition', () => {
  let mockSpeechRecognition: any
  let mockRecognitionInstance: any

  beforeEach(() => {
    // Reset mocks before each test
    vi.clearAllMocks()

    // Mock SpeechRecognition instance
    mockRecognitionInstance = {
      start: vi.fn(),
      stop: vi.fn(),
      abort: vi.fn(),
      addEventListener: vi.fn((event: string, callback: Function) => {
        if (event === 'result') {
          mockRecognitionInstance._onResult = callback
        }
        if (event === 'error') {
          mockRecognitionInstance._onError = callback
        }
        if (event === 'end') {
          mockRecognitionInstance._onEnd = callback
        }
        if (event === 'start') {
          mockRecognitionInstance._onStart = callback
        }
      }),
      removeEventListener: vi.fn(),
      continuous: false,
      interimResults: false,
      lang: 'fr-FR',
    }

    // Mock constructor
    mockSpeechRecognition = vi.fn(() => mockRecognitionInstance)
    mockSpeechRecognition.prototype = {}

    // Store original window values
    const windowAny = window as any
    windowAny.SpeechRecognition = mockSpeechRecognition
  })

  afterEach(() => {
    const windowAny = window as any
    delete windowAny.SpeechRecognition
    delete windowAny.webkitSpeechRecognition
  })

  it('detects when SpeechRecognition is supported', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    expect(result.current.isSupported).toBe(true)
  })

  it('detects when SpeechRecognition is not supported', () => {
    const windowAny = window as any
    delete windowAny.SpeechRecognition
    delete windowAny.webkitSpeechRecognition

    const { result } = renderHook(() => useSpeechRecognition())
    expect(result.current.isSupported).toBe(false)
  })

  it('starts with isListening = false', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    expect(result.current.isListening).toBe(false)
  })

  it('starts with empty transcript', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    expect(result.current.transcript).toBe('')
  })

  it('starts with no error', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    expect(result.current.error).toBe(null)
  })

  it('configures recognition with correct language and settings', () => {
    renderHook(() => useSpeechRecognition())
    expect(mockSpeechRecognition).toHaveBeenCalled()
    expect(mockRecognitionInstance.continuous).toBe(false)
    expect(mockRecognitionInstance.interimResults).toBe(true)
    expect(mockRecognitionInstance.lang).toBe('fr-FR')
  })

  it('startListening calls start on recognition when supported', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      result.current.startListening()
    })
    expect(mockRecognitionInstance.start).toHaveBeenCalled()
  })

  it('startListening does nothing gracefully when not supported', () => {
    const windowAny = window as any
    delete windowAny.SpeechRecognition
    delete windowAny.webkitSpeechRecognition

    const { result } = renderHook(() => useSpeechRecognition())
    expect(() => {
      act(() => {
        result.current.startListening()
      })
    }).not.toThrow()
  })

  it('stopListening calls stop on recognition when supported', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      result.current.startListening()
    })
    act(() => {
      result.current.stopListening()
    })
    expect(mockRecognitionInstance.stop).toHaveBeenCalled()
  })

  it('stopListening does nothing gracefully when not supported', () => {
    const windowAny = window as any
    delete windowAny.SpeechRecognition
    delete windowAny.webkitSpeechRecognition

    const { result } = renderHook(() => useSpeechRecognition())
    expect(() => {
      act(() => {
        result.current.stopListening()
      })
    }).not.toThrow()
  })

  it('updates isListening to true on start event', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      result.current.startListening()
    })
    // Simulate start event
    act(() => {
      mockRecognitionInstance._onStart?.()
    })
    expect(result.current.isListening).toBe(true)
  })

  it('updates isListening to false on end event', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      result.current.startListening()
    })
    act(() => {
      mockRecognitionInstance._onStart?.()
    })
    act(() => {
      mockRecognitionInstance._onEnd?.()
    })
    expect(result.current.isListening).toBe(false)
  })

  it('clears error on start event', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    // First trigger an error
    act(() => {
      mockRecognitionInstance._onError?.({ error: 'network' })
    })
    expect(result.current.error).not.toBe(null)
    // Now start and should clear error
    act(() => {
      result.current.startListening()
    })
    act(() => {
      mockRecognitionInstance._onStart?.()
    })
    expect(result.current.error).toBe(null)
  })

  it('sets error on error event', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      mockRecognitionInstance._onError?.({ error: 'network' })
    })
    expect(result.current.error).toBe('network')
  })

  it('accumulates transcript from result events', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      result.current.startListening()
    })
    // Simulate interim result
    act(() => {
      const mockEvent = {
        results: [
          [{ transcript: 'Hello', confidence: 0.95 }],
        ] as any as SpeechRecognitionResultList,
      } as any as SpeechRecognitionEvent
      Object.defineProperty(mockEvent, 'results', {
        value: [
          { 0: { transcript: 'Hello' }, isFinal: false, length: 1 },
        ],
      })
      mockRecognitionInstance._onResult?.(mockEvent)
    })
    expect(result.current.transcript).toContain('Hello')
  })

  it('handles multiple interim results correctly', () => {
    const { result } = renderHook(() => useSpeechRecognition())
    act(() => {
      result.current.startListening()
    })
    act(() => {
      const mockEvent = {
        results: [
          { 0: { transcript: 'Buy a' }, isFinal: false, length: 1 },
          { 0: { transcript: ' laptop' }, isFinal: false, length: 1 },
        ],
      } as any as SpeechRecognitionEvent
      mockRecognitionInstance._onResult?.(mockEvent)
    })
    expect(result.current.transcript).toContain('Buy a')
    expect(result.current.transcript).toContain('laptop')
  })
})
