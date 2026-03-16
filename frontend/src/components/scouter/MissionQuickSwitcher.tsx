import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search, X } from 'lucide-react'
import { useMissions } from '../../hooks'
import styles from './MissionQuickSwitcher.module.css'

interface MissionQuickSwitcherProps {
  open: boolean
  onClose: () => void
}

const PHASE_LABELS: Record<string, string> = {
  researching: 'Recherche',
  comparing:   'Comparaison',
  buying:      'Achat',
  done:        'Terminé',
}

export function MissionQuickSwitcher({ open, onClose }: MissionQuickSwitcherProps) {
  const { missions } = useMissions()
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(-1)
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)

  const filtered = query.trim()
    ? missions.filter(m => m.name.toLowerCase().includes(query.toLowerCase()))
    : missions.slice(0, 8)

  useEffect(() => {
    if (open) {
      setQuery('')
      setActiveIndex(-1)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  useEffect(() => { setActiveIndex(-1) }, [query])

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
        aria-modal="true"
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
              if (e.key === 'Escape') { onClose(); return }
              if (e.key === 'ArrowDown') { e.preventDefault(); setActiveIndex(i => Math.min(i + 1, filtered.length - 1)) }
              if (e.key === 'ArrowUp') { e.preventDefault(); setActiveIndex(i => Math.max(i - 1, 0)) }
              if (e.key === 'Enter') {
                const idx = activeIndex >= 0 ? activeIndex : 0
                if (filtered[idx]) handleSelect(filtered[idx].slug)
              }
            }}
          />
          <kbd className={styles.esc}>Échap</kbd>
          <button className={styles.closeBtn} onClick={onClose} aria-label="Fermer" type="button">
            <X size={14} aria-hidden="true" />
          </button>
        </div>
        <ul className={styles.list}>
          {filtered.length === 0 && (
            <li className={styles.empty}>Aucune mission trouvée</li>
          )}
          {filtered.map((m, i) => (
            <li key={m.id}>
              <button
                className={`${styles.item} ${i === activeIndex ? styles.itemActive : ''}`}
                onClick={() => handleSelect(m.slug)}
              >
                <span className={styles.icon}>{m.icon}</span>
                <span className={styles.name}>{m.name}</span>
                <span className={styles.phase}>{PHASE_LABELS[m.phase] ?? m.phase}</span>
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
