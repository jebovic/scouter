import { useRef, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useSearchInput, useSearch } from '../../hooks/useSearch'
import styles from './SearchDropdown.module.css'

const MAX_DROPDOWN = 5

function badgeColor(badge: string): string {
  if (badge === 'recommended') return 'var(--cyan)'
  if (badge === 'alternative') return 'var(--text-mid)'
  if (badge === 'rejected') return 'var(--coral-dim, #ff6b6b)'
  return 'var(--purple, #9b59b6)'
}

export function SearchDropdown() {
  const navigate = useNavigate()
  const { query, setQuery, isOpen, open, close, clear, inputRef } = useSearchInput()
  const { data: results = [], isFetching } = useSearch(query, MAX_DROPDOWN)
  const wrapRef = useRef<HTMLDivElement>(null)

  // Close on outside click
  useEffect(() => {
    function onMouseDown(e: MouseEvent) {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
        close()
      }
    }
    document.addEventListener('mousedown', onMouseDown)
    return () => document.removeEventListener('mousedown', onMouseDown)
  }, [close])

  const showDropdown = isOpen && query.trim().length >= 2

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Escape') {
      clear()
    } else if (e.key === 'Enter' && query.trim()) {
      const encoded = encodeURIComponent(query.trim())
      navigate(`/search?q=${encoded}`)
      clear()
    }
  }

  return (
    <div className={styles.wrapper} ref={wrapRef}>
      <div className={styles.inputWrap}>
        <span className={styles.icon}>⌕</span>
        <input
          ref={inputRef}
          className={styles.input}
          type="search"
          placeholder="Search options…"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value)
            if (e.target.value.length >= 2) open()
          }}
          onFocus={() => query.trim().length >= 2 && open()}
          onKeyDown={handleKeyDown}
          aria-label="Search options"
          aria-expanded={showDropdown}
          aria-autocomplete="list"
        />
      </div>

      {showDropdown && (
        <div className={styles.dropdown} role="listbox">
          {isFetching && results.length === 0 && (
            <div className={styles.empty}>Searching…</div>
          )}
          {!isFetching && results.length === 0 && (
            <div className={styles.empty}>No results for "{query}"</div>
          )}
          {results.map((r) => (
            <Link
              key={r.id}
              to={`/missions/${r.missionSlug}/options`}
              className={styles.result}
              role="option"
              onClick={clear}
            >
              <div className={styles.resultBody}>
                <div className={styles.resultName}>{r.name}</div>
                <div className={styles.resultMeta}>
                  <span style={{ color: badgeColor(r.badge) }}>●</span>{' '}
                  {r.badge} · {r.missionName}
                  {r.category ? ` · ${r.category}` : ''}
                </div>
              </div>
              <span className={styles.resultScore}>
                {Math.round(r.similarity * 100)}%
              </span>
            </Link>
          ))}
          {results.length > 0 && (
            <div
              className={styles.footer}
              role="button"
              tabIndex={0}
              onClick={() => {
                navigate(`/search?q=${encodeURIComponent(query.trim())}`)
                clear()
              }}
              onKeyDown={(e) => e.key === 'Enter' && navigate(`/search?q=${encodeURIComponent(query.trim())}`)}
            >
              See all results →
            </div>
          )}
        </div>
      )}
    </div>
  )
}
