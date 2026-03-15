import { useState, useEffect, useRef } from 'react'

interface SpeechRecognitionEvent extends Event {
  results?: SpeechRecognitionResultList
}

interface SpeechRecognitionErrorEvent extends Event {
  error?: string
}

interface UseSpeechRecognitionResult {
  isListening: boolean
  transcript: string
  error: string | null
  isSupported: boolean
  startListening: () => void
  stopListening: () => void
}

export function useSpeechRecognition(): UseSpeechRecognitionResult {
  const [isListening, setIsListening] = useState(false)
  const [transcript, setTranscript] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSupported, setIsSupported] = useState(false)

  const recognitionRef = useRef<any>(null)

  useEffect(() => {
    // Check browser support
    const SpeechRecognitionAPI =
      (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition

    if (!SpeechRecognitionAPI) {
      setIsSupported(false)
      return
    }

    setIsSupported(true)

    // Initialize recognition
    const recognition = new SpeechRecognitionAPI()
    recognition.continuous = false
    recognition.interimResults = true
    recognition.lang = 'fr-FR'

    recognitionRef.current = recognition

    // Handle start event
    recognition.addEventListener('start', () => {
      setIsListening(true)
      setError(null)
    })

    // Handle end event
    recognition.addEventListener('end', () => {
      setIsListening(false)
    })

    // Handle result event
    recognition.addEventListener('result', (event: SpeechRecognitionEvent) => {
      if (!event.results) return

      let interimTranscript = ''
      for (let i = event.results.length - 1; i >= 0; i--) {
        const transcriptSegment = event.results[i][0].transcript

        if (event.results[i].isFinal) {
          // Keep final results
        } else {
          // Accumulate interim results
          interimTranscript += transcriptSegment + ' '
        }
      }

      // Collect all transcripts (both interim and final)
      let fullTranscript = ''
      for (let i = 0; i < event.results.length; i++) {
        fullTranscript += event.results[i][0].transcript
        if (!event.results[i].isFinal) {
          fullTranscript += ' '
        }
      }

      setTranscript(fullTranscript.trim())
    })

    // Handle error event
    recognition.addEventListener('error', (event: SpeechRecognitionErrorEvent) => {
      setError(event.error || 'unknown_error')
    })

    // Cleanup
    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.removeEventListener('start', () => {})
        recognitionRef.current.removeEventListener('end', () => {})
        recognitionRef.current.removeEventListener('result', () => {})
        recognitionRef.current.removeEventListener('error', () => {})
        recognitionRef.current.abort()
      }
    }
  }, [])

  const startListening = () => {
    if (recognitionRef.current && isSupported) {
      recognitionRef.current.start()
    }
  }

  const stopListening = () => {
    if (recognitionRef.current && isSupported) {
      recognitionRef.current.stop()
    }
  }

  return {
    isListening,
    transcript,
    error,
    isSupported,
    startListening,
    stopListening,
  }
}
