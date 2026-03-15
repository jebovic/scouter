import { useRef, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useSearchInput, useSearch } from '../../hooks/useSearch'
import { useSearchSuggestions } from '../../hooks/useSearchSuggestions'
import { addToHistory, removeFromHistory } from '../../utils/searchHistory'
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
  const { suggestions, history, clearHistory, refreshHistory } = useSearchSuggestions(query)
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

  const trimmed = query.trim()
  const showResultsDropdown = isOpen && trimmed.length >= 2
  // Show suggestions/history panel when focused with short or empty query
  const showSuggestionsDropdown = isOpen && trimmed.length < 2 && (history.length > 0 || trimmed.length > 0)

  function navigateToSearch(q: string) {
    addToHistory(q)
    navigate(`/search?q=${encodeURIComponent(q)}`)
    clear()
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Escape') {
      clear()
    } else if (e.key === 'Enter' && trimmed) {
      navigateToSearch(trimmed)
    }
  }

  function handleRemoveHistory(e: React.MouseEvent, entry: string) {
    e.preventDefault()
    e.stopPropagation()
    removeFromHistory(entry)
    refreshHistory()
  }

  function handleSuggestionClick(term: string) {
    navigateToSearch(term)
  }

  return (
    <div className={styles.wrapper} ref={wrapRef}>
      <div className={styles.inputWrap}>
        <span className={styles.icon}>⌕</span>
        <input
          ref={inputRef}
          className={styles.input}
          type="search"
          placeholder="Rechercher…"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value)
            open()
          }}
          onFocus={() => open()}
          onKeyDown={handleKeyDown}
          aria-label="Rechercher des options"
          aria-expanded={showResultsDropdown || showSuggestionsDropdown}
          aria-autocomplete="list"
        />
      </div>

      {/* Suggestions / History panel (query < 2 chars) */}
      {showSuggestionsDropdown && (
        <div className={styles.dropdown} role="listbox">
          {history.length > 0 && (
            <>
              <div className={styles.sectionHeader}>
                <span>Recherches récentes</span>
                <button
                  className={styles.clearBtn}
                  onClick={clearHistory}
                  aria-label="Effacer l'historique"
                >
                  Effacer l&apos;historique
                </button>
              </div>
              {history.map((entry) => (
                <div key={entry} className={styles.suggestion} role="option">
                  <span className={styles.suggestionIcon}>🕐</span>
                  <span
                    className={styles.suggestionText}
                    onClick={() => handleSuggestionClick(entry)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => e.key === 'Enter' && handleSuggestionClick(entry)}
                  >
                    {entry}
                  </span>
                  <button
                    className={styles.removeBtn}
                    onClick={(e) => handleRemoveHistory(e, entry)}
                    aria-label={`Supprimer "${entry}" de l'historique`}
                  >
                    ×
                  </button>
                </div>
              ))}
            </>
          )}

          {history.length === 0 && trimmed.length === 0 && (
            <div className={styles.empty}>Tapez pour rechercher…</div>
          )}
        </div>
      )}

      {/* Results + suggestions panel (query >= 2 chars) */}
      {showResultsDropdown && (
        <div className={styles.dropdown} role="listbox">
          {isFetching && results.length === 0 && (
            <div className={styles.empty}>Recherche en cours…</div>
          )}
          {!isFetching && results.length === 0 && suggestions.length === 0 && (
            <div className={styles.empty}>Aucun résultat pour &quot;{query}&quot;</div>
          )}

          {results.map((r) => (
            <Link
              key={r.id}
              to={`/missions/${r.missionSlug}/options`}
              className={styles.result}
              role="option"
              onClick={() => {
                addToHistory(trimmed)
                clear()
              }}
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

          {suggestions.length > 0 && (
            <>
              <div className={styles.sectionHeader}>
                <span>Suggestions</span>
              </div>
              {suggestions.slice(0, 4).map((term) => (
                <div
                  key={term}
                  className={styles.suggestion}
                  role="option"
                  onClick={() => handleSuggestionClick(term)}
                  tabIndex={0}
                  onKeyDown={(e) => e.key === 'Enter' && handleSuggestionClick(term)}
                >
                  <span className={styles.suggestionIcon}>🔍</span>
                  <span className={styles.suggestionText}>{term}</span>
                </div>
              ))}
            </>
          )}

          {(results.length > 0) && (
            <div
              className={styles.footer}
              role="button"
              tabIndex={0}
              onClick={() => navigateToSearch(trimmed)}
              onKeyDown={(e) => e.key === 'Enter' && navigateToSearch(trimmed)}
            >
              Voir tous les résultats →
            </div>
          )}
        </div>
      )}
    </div>
  )
}
