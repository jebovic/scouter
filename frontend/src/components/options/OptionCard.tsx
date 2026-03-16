import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, Badge } from '../scouter'
import { AttributeRenderer } from './AttributeRenderer'
import { VoteButtons } from './VoteButtons'
import { ReviewSummaryCard } from './ReviewSummaryCard'
import { PriceComparisonPanel } from './PriceComparisonPanel'
import { RetailerLinks } from './RetailerLinks'
import { SubstituteSuggestions } from './SubstituteSuggestions'
import { OptionNote } from './OptionNote'
import { ImageStrip } from './ImageStrip'
import { useOptionImages } from '../../hooks/useOptionImages'
import { useTranslatedOption } from '../../hooks/useTranslatedOption'
import type { Option } from '../../types'
import styles from './OptionCard.module.css'

interface OptionCardProps {
  option: Option
  missionId: string
  currency?: string
  score?: number
  onBadgeChange?: (badge: Option['badge']) => void
  onPin?: (optionId: string) => void
  onReject?: (optionId: string) => void
  onUnreject?: (optionId: string) => void
  compareMode?: boolean
  selected?: boolean
  onToggleSelect?: (id: string) => void
  selectionFull?: boolean
  onEdit?: (optionId: string) => void
  onDelete?: (optionId: string) => void
}

const BADGES: Option['badge'][] = ['recommended', 'alternative', 'watch', 'rejected']

function scoreColor(s: number): { text: string; bg: string } {
  if (s >= 75) return { text: 'var(--green)', bg: 'var(--green-dim)' }
  if (s >= 50) return { text: 'var(--gold)', bg: 'var(--gold-dim)' }
  if (s >= 25) return { text: 'var(--orange)', bg: 'var(--orange-dim)' }
  return { text: 'var(--coral)', bg: 'var(--coral-dim)' }
}

export function OptionCard({
  option,
  missionId,
  currency = 'USD',
  score,
  onBadgeChange,
  onPin,
  onReject,
  onUnreject,
  compareMode = false,
  selected = false,
  onToggleSelect,
  selectionFull = false,
  onEdit,
  onDelete,
}: OptionCardProps) {
  const { i18n, t } = useTranslation()
  const [menuOpen, setMenuOpen] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuOpen) return
    function handleMouseDown(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleMouseDown)
    return () => document.removeEventListener('mousedown', handleMouseDown)
  }, [menuOpen])

  useEffect(() => {
    if (!menuOpen && !confirmDelete) return
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        setMenuOpen(false)
        setConfirmDelete(false)
      }
    }
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [menuOpen, confirmDelete])
  const displayOption = useTranslatedOption(option)
  const isRecommended = option.badge === 'recommended'
  const isRejected = option.badge === 'rejected'
  const isDimmed = (isRejected || (compareMode && selectionFull && !selected))
  const { data: images = [], isLoading: imagesLoading } = useOptionImages(missionId, option.id)
  const isPendingTranslation = i18n.language !== 'en' && !option.translations?.[i18n.language]

  return (
    <Card
      glow={isRecommended}
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 12,
        opacity: isDimmed ? (compareMode && selectionFull && !selected ? 0.5 : 0.6) : 1,
        transition: 'opacity 0.2s, border-color 0.2s, box-shadow 0.2s',
        position: 'relative',
        borderColor: selected ? 'var(--cyan)' : undefined,
        boxShadow: selected ? '0 0 0 2px var(--cyan-dim)' : undefined,
      }}
    >
      <ImageStrip images={images} loading={imagesLoading} />

      {/* Compare mode checkbox overlay */}
      {compareMode && (
        <button
          onClick={() => onToggleSelect?.(option.id)}
          aria-label={selected ? `Deselect ${option.name}` : `Select ${option.name}`}
          aria-pressed={selected}
          style={{
            position: 'absolute',
            top: 10,
            left: 10,
            width: 22,
            height: 22,
            borderRadius: 4,
            border: `2px solid ${selected ? 'var(--cyan)' : 'var(--border)'}`,
            background: selected ? 'var(--cyan)' : 'var(--surface)',
            cursor: selectionFull && !selected ? 'not-allowed' : 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 2,
            transition: 'border-color 0.15s, background 0.15s',
            padding: 0,
          }}
          disabled={selectionFull && !selected}
        >
          {selected && (
            <span style={{ color: 'var(--bg)', fontSize: '0.75rem', fontWeight: 700, lineHeight: 1 }}>
              ✓
            </span>
          )}
        </button>
      )}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <div
            style={{
              fontFamily: 'var(--font-display)',
              fontWeight: 700,
              fontSize: '1rem',
              color: 'var(--text)',
            }}
          >
            {displayOption.name}
            {isPendingTranslation && (
              <span className={styles.translatingBadge}>🌐 {t('options.translating')}</span>
            )}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)', marginTop: 2 }}>
            {displayOption.category}
          </div>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {score !== undefined && (() => {
            const c = scoreColor(score)
            return (
              <span
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.7rem',
                  fontWeight: 700,
                  color: c.text,
                  background: c.bg,
                  border: `1px solid ${c.text}`,
                  borderRadius: 4,
                  padding: '2px 7px',
                  letterSpacing: '0.05em',
                  animation: 'card-enter 0.3s ease both',
                }}
              >
                {Math.round(score)}
              </span>
            )
          })()}
          <Badge variant={option.badge} />
          {(onEdit || onDelete) && (
            <div className={styles.menuWrapper} ref={menuRef}>
              <button
                className={styles.menuBtn}
                aria-label="more options"
                onClick={() => { setMenuOpen(o => !o); setConfirmDelete(false) }}
              >
                ⋯
              </button>
              {menuOpen && (
                <div className={styles.menu} role="menu">
                  {onEdit && (
                    <button
                      role="menuitem"
                      className={styles.menuItem}
                      onClick={() => { onEdit(option.id); setMenuOpen(false) }}
                    >
                      {t('option.actions.edit')}
                    </button>
                  )}
                  {onDelete && (
                    <button
                      role="menuitem"
                      className={`${styles.menuItem} ${styles.menuItemDanger}`}
                      onClick={() => setConfirmDelete(true)}
                    >
                      {t('option.actions.delete')}
                    </button>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <OptionNote optionId={option.id} />

      {option.priceRange && (
        <div
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '1.1rem',
            color: 'var(--cyan)',
            fontWeight: 600,
          }}
        >
          {new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(
            option.priceRange.best,
          )}
          <span style={{ fontSize: '0.7rem', color: 'var(--text-dim)', marginLeft: 6 }}>
            best price
          </span>
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {displayOption.attributes.slice(0, 5).map((attr) => (
          <AttributeRenderer key={attr.key} attribute={attr} currency={currency} />
        ))}
      </div>

      <RetailerLinks query={displayOption.name} category={displayOption.category} showStock={true} />

      {displayOption.warnings.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {displayOption.warnings.map((w, i) => (
            <div
              key={i}
              style={{
                fontSize: '0.75rem',
                color: 'var(--coral)',
                padding: '4px 8px',
                background: 'var(--coral-dim)',
                borderRadius: 4,
                border: '1px solid var(--coral-dim)',
              }}
            >
              ⚠ {w}
            </div>
          ))}
        </div>
      )}

      {displayOption.notes && (
        <p style={{ fontSize: '0.8rem', color: 'var(--text-mid)', margin: 0 }}>
          {displayOption.notes}
        </p>
      )}

      {onBadgeChange && (
        <div style={{ display: 'flex', gap: 4, borderTop: '1px solid var(--border)', paddingTop: 10 }}>
          {BADGES.map((b) => (
            <button
              key={b}
              aria-label={b}
              aria-pressed={option.badge === b}
              onClick={() => onBadgeChange(b)}
              style={{
                flex: 1,
                padding: '4px 0',
                borderRadius: 4,
                border: `1px solid ${option.badge === b ? 'var(--border-glow)' : 'var(--border)'}`,
                background: option.badge === b ? 'var(--raised)' : 'transparent',
                color: option.badge === b ? 'var(--text)' : 'var(--text-dim)',
                cursor: 'pointer',
                fontSize: '0.65rem',
                fontFamily: 'var(--font-mono)',
                letterSpacing: '0.05em',
                transition: 'all 0.15s',
              }}
            >
              {b.toUpperCase().slice(0, 3)}
            </button>
          ))}
        </div>
      )}

      {option.url && (
        <a
          href={option.url}
          target="_blank"
          rel="noopener noreferrer"
          style={{ fontSize: '0.75rem', color: 'var(--cyan)', marginTop: -4 }}
        >
          View →
        </a>
      )}

      <VoteButtons optionId={option.id} />

      {(onPin || onReject) && (
        <div style={{ display: 'flex', gap: 6, borderTop: '1px solid var(--border)', paddingTop: 10 }}>
          {onPin && (
            <button
              onClick={() => onPin(option.id)}
              style={{
                flex: 1,
                padding: '4px 0',
                borderRadius: 4,
                border: `1px solid ${option.pinned ? 'var(--cyan)' : 'var(--border)'}`,
                background: option.pinned ? 'var(--cyan-dim)' : 'transparent',
                color: option.pinned ? 'var(--cyan)' : 'var(--text-dim)',
                cursor: 'pointer',
                fontSize: '0.65rem',
                fontFamily: 'var(--font-mono)',
                letterSpacing: '0.05em',
                transition: 'all 0.15s',
              }}
            >
              {option.pinned ? '📌 PINNED' : '📌 PIN'}
            </button>
          )}
          {!option.pinned && (option.rejected ? onUnreject : onReject) && (
            <button
              onClick={() => option.rejected ? onUnreject!(option.id) : onReject!(option.id)}
              style={{
                flex: 1,
                padding: '4px 0',
                borderRadius: 4,
                border: `1px solid ${option.rejected ? 'var(--coral)' : 'var(--border)'}`,
                background: option.rejected ? 'var(--coral-dim)' : 'transparent',
                color: option.rejected ? 'var(--coral)' : 'var(--text-dim)',
                cursor: 'pointer',
                fontSize: '0.65rem',
                fontFamily: 'var(--font-mono)',
                letterSpacing: '0.05em',
                transition: 'all 0.15s',
              }}
            >
              {option.rejected ? '✕ REJECTED' : '✕ REJECT'}
            </button>
          )}
          {option.rejectReason && (
            <div
              style={{
                fontSize: '0.65rem',
                color: 'var(--text-dim)',
                fontStyle: 'italic',
                alignSelf: 'center',
              }}
            >
              {option.rejectReason}
            </div>
          )}
        </div>
      )}

      <ReviewSummaryCard optionId={option.id} />
      <PriceComparisonPanel optionId={option.id} />
      <SubstituteSuggestions optionId={option.id} />

      {confirmDelete && (
        <div className={styles.confirmOverlay} onClick={() => setConfirmDelete(false)}>
          <div className={styles.confirmDialog} role="dialog" aria-modal="true" onClick={e => e.stopPropagation()}>
            <p>{t('option.actions.deleteConfirm')}</p>
            <div className={styles.confirmActions}>
              <button onClick={() => setConfirmDelete(false)}>{t('common.cancel')}</button>
              <button
                className={styles.deleteConfirmBtn}
                aria-label={t('common.delete')}
                onClick={() => { onDelete!(option.id) }}
              >
                {t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </Card>
  )
}
