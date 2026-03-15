import { useState, useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSearch } from '../hooks/useSearch'
import styles from './SearchPage.module.css'

export function SearchPage() {
  const { t } = useTranslation()
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

  return (
    <main className={styles.page}>
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

      {inputValue.trim().length < 2 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>{t('search.startTyping')}</div>
          <div className={styles.emptyHint}>{t('search.minChars')}</div>
        </div>
      )}

      {inputValue.trim().length >= 2 && isFetching && results.length === 0 && (
        <div className={styles.empty}>
          <div className={styles.emptyTitle}>{t('search.searching')}</div>
        </div>
      )}

      {inputValue.trim().length >= 2 && !isFetching && results.length === 0 && (
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
  )
}
