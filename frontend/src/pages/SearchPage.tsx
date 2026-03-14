import { useState, useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useSearch } from '../hooks/useSearch'
import styles from './SearchPage.module.css'

function badgeLabel(badge: string): string {
  const map: Record<string, string> = {
    recommended: '★ Recommended',
    alternative: '◆ Alternative',
    watch: '◎ Watch',
    rejected: '✕ Rejected',
  }
  return map[badge] ?? badge
}

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const initialQ = searchParams.get('q') ?? ''
  const [inputValue, setInputValue] = useState(initialQ)

  const { data: results = [], isFetching } = useSearch(inputValue, 20)

  // Sync query param on input change (debounce handled inside useSearch)
  useEffect(() => {
    if (inputValue.trim()) {
      setSearchParams({ q: inputValue }, { replace: true })
    } else {
      setSearchParams({}, { replace: true })
    }
  }, [inputValue, setSearchParams])

  return (
    <main className={styles.page}>
      <div className={styles.header}>
        <div className={styles.title}>SEMANTIC SEARCH</div>
        <div className={styles.subtitle}>
          Find options across all missions by meaning, not just keywords.
        </div>
      </div>

      <div className={styles.searchBar}>
        <input
          className={styles.input}
          type="search"
          placeholder="e.g. quiet mechanical keyboard with RGB"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          autoFocus
        />
      </div>

      {inputValue.trim().length < 2 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>Start typing to search</div>
          <div className={styles.emptyHint}>Minimum 2 characters</div>
        </div>
      )}

      {inputValue.trim().length >= 2 && isFetching && results.length === 0 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>Searching…</div>
        </div>
      )}

      {inputValue.trim().length >= 2 && !isFetching && results.length === 0 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>No results for "{inputValue}"</div>
          <div className={styles.emptyHint}>
            Try different keywords, or run research on your missions first.
          </div>
        </div>
      )}

      {results.length > 0 && (
        <div className={styles.results}>
          {results.map((r) => (
            <Link
              key={r.id}
              to={`/missions/${r.missionSlug}/options`}
              className={styles.card}
            >
              <div className={styles.cardBody}>
                <div className={styles.cardName}>{r.name}</div>
                <div className={styles.cardMeta}>
                  {badgeLabel(r.badge)} · {r.missionName}
                  {r.category ? ` · ${r.category}` : ''}
                  {r.notes ? ` — ${r.notes.slice(0, 80)}${r.notes.length > 80 ? '…' : ''}` : ''}
                </div>
              </div>
              <div className={styles.cardScore}>
                <div>{Math.round(r.similarity * 100)}%</div>
                <div className={styles.scoreLabel}>match</div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </main>
  )
}
