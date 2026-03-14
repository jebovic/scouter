import { useEffect, useRef } from 'react'

export type ShortcutMap = Record<string, () => void>

export function useKeyboardShortcuts(shortcuts: ShortcutMap): void {
  const shortcutsRef = useRef(shortcuts)
  useEffect(() => { shortcutsRef.current = shortcuts })

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.ctrlKey || event.metaKey || event.altKey) return

      const target = event.target as Element | null
      if (target) {
        const tag = target.tagName?.toLowerCase() ?? ''
        if (tag === 'input' || tag === 'textarea' || tag === 'select') return
        if (target.hasAttribute?.('contenteditable')) return
      }

      const handler = shortcutsRef.current[event.key]
      if (handler) {
        event.preventDefault()
        handler()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])
}
