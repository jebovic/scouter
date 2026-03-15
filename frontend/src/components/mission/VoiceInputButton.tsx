import { useEffect } from 'react'
import { useSpeechRecognition } from '../../hooks/useSpeechRecognition'
import styles from './VoiceInputButton.module.css'

interface VoiceInputButtonProps {
  onTranscript: (text: string) => void
}

export function VoiceInputButton({ onTranscript }: VoiceInputButtonProps) {
  const { isListening, transcript, error, isSupported, startListening, stopListening } =
    useSpeechRecognition()

  // Call onTranscript when transcript changes
  useEffect(() => {
    if (transcript) {
      onTranscript(transcript)
    }
  }, [transcript, onTranscript])

  if (!isSupported) {
    return <span className={styles.unsupported}>Votre navigateur ne supporte pas la saisie vocale</span>
  }

  const handleClick = () => {
    if (isListening) {
      stopListening()
    } else {
      startListening()
    }
  }

  return (
    <div className={styles.container}>
      <button
        onClick={handleClick}
        className={`${styles.button} ${isListening ? styles.listening : ''}`}
        title={isListening ? 'Arrêter' : 'Commencer'}
        aria-label={isListening ? 'Stop recording' : 'Start recording'}
      >
        {isListening ? '⏹' : '🎤'}
      </button>
      {error && <span className={styles.error}>Erreur: {error}</span>}
    </div>
  )
}
