import { useState, useEffect } from 'react'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSearch } from '../hooks/useSearch'
import { addToHistory } from '../utils/searchHistory'
import { Topnav } from '../components/scouter/Topnav'
import styles from './SearchPage.module.css'

const POPULAR_TERMS: string[] = [
  'iPhone',
  'MacBook',
  'PS5',
  'TV 4K',
  'aspirateur robot',
  'vélo électrique',
  'AirPods',
  'iPad',
  'Nintendo Switch',
  'casque audio',
  'drone',
  'trottinette électrique',
]

export function SearchPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const initialQ = searchParams.get('q') ?? ''
  const [inputValue, setInputValue] = useState(initialQ)
  // Track whether we've saved the current query to history
  const [lastSaved, setLastSaved] = useState('')

  const { data: results = [], isFetching } = useSearch(inputValue, 20)

  // Sync query param on input change (debounce handled inside useSearch)
  useEffect(() => {
    if (inputValue.trim()) {
      setSearchParams({ q: inputValue }, { replace: true })
    } else {
      setSearchParams({}, { replace: true })
    }
  }, [inputValue, setSearchParams])

  // Save to history when results arrive for a query
  useEffect(() => {
    const trimmed = inputValue.trim()
    if (trimmed.length >= 2 && !isFetching && results.length > 0 && trimmed !== lastSaved) {
      addToHistory(trimmed)
      setLastSaved(trimmed)
    }
  }, [inputValue, isFetching, results.length, lastSaved])

  function handlePopularClick(term: string) {
    addToHistory(term)
    navigate(`/search?q=${encodeURIComponent(term)}`)
    setInputValue(term)
  }

  function badgeLabel(badge: string): string {
    const prefixes: Record<string, string> = {
      recommended: '★',
      alternative: '◆',
      watch: '◎',
      rejected: '✕',
    }
    const prefix = prefixes[badge] ?? ''
    const label = t(`search.${badge}`, { defaultValue: badge })
    return prefix ? `${prefix} ${label}` : label
  }

  const showEmpty = inputValue.trim().length < 2

  return (
    <>
      <Topnav />
      <main className={`page ${styles.page}`}>
      <div className={styles.header}>
        <div className={styles.title}>{t('search.title').toUpperCase()}</div>
        <div className={styles.subtitle}>{t('search.subtitle')}</div>
      </div>

      <div className={styles.searchBar}>
        <input
          className={styles.input}
          type="search"
          placeholder={t('search.placeholder')}
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          autoFocus
        />
      </div>

      {showEmpty && (
        <>
          <div className={styles.empty}>
            <div className={styles.emptyTitle}>{t('search.startTyping')}</div>
            <div className={styles.emptyHint}>{t('search.minChars')}</div>
          </div>

          <div className={styles.popularSection}>
            <div className={styles.popularLabel}>{t('search.popularLabel')}</div>
            <div className={styles.chipList}>
              {POPULAR_TERMS.map((term) => (
                <button
                  key={term}
                  className={styles.chip}
                  onClick={() => handlePopularClick(term)}
                >
                  🔍 {term}
                </button>
              ))}
            </div>
          </div>
        </>
      )}

      {!showEmpty && isFetching && results.length === 0 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>{t('search.searching')}</div>
        </div>
      )}

      {!showEmpty && !isFetching && results.length === 0 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>{t('search.noResults', { query: inputValue })}</div>
          <div className={styles.emptyHint}>{t('search.noResultsHint')}</div>
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
                <div className={styles.scoreLabel}>{t('search.match')}</div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </main>
    </>
  )
}
