import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search } from 'lucide-react'
import { useMissions } from '../../hooks'
import styles from './MissionQuickSwitcher.module.css'

interface MissionQuickSwitcherProps {
  open: boolean
  onClose: () => void
}

export function MissionQuickSwitcher({ open, onClose }: MissionQuickSwitcherProps) {
  const { missions } = useMissions()
  const [query, setQuery] = useState('')
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)

  const filtered = query.trim()
    ? missions.filter(m => m.name.toLowerCase().includes(query.toLowerCase()))
    : missions.slice(0, 8)

  useEffect(() => {
    if (open) {
      setQuery('')
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  function handleSelect(slug: string) {
    navigate(`/missions/${slug}`)
    onClose()
  }

  if (!open) return null

  return (
    <div className={styles.overlay} onClick={onClose} role="presentation">
      <div
        className={styles.modal}
        onClick={e => e.stopPropagation()}
        role="dialog"
        aria-label="Navigation rapide"
      >
        <div className={styles.searchRow}>
          <Search size={16} className={styles.searchIcon} aria-hidden="true" />
          <input
            ref={inputRef}
            className={styles.input}
            placeholder="Nom de mission..."
            value={query}
            onChange={e => setQuery(e.target.value)}
            onKeyDown={e => {
              if (e.key === 'Escape') onClose()
              if (e.key === 'Enter' && filtered.length > 0) handleSelect(filtered[0].slug)
            }}
          />
          <kbd className={styles.esc}>esc</kbd>
        </div>
        <ul className={styles.list} role="listbox">
          {filtered.length === 0 && (
            <li className={styles.empty}>Aucune mission trouvée</li>
          )}
          {filtered.map(m => (
            <li key={m.id} role="option" aria-selected={false}>
              <button className={styles.item} onClick={() => handleSelect(m.slug)}>
                <span className={styles.icon}>{m.icon}</span>
                <span className={styles.name}>{m.name}</span>
                <span className={styles.phase}>{m.phase}</span>
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
