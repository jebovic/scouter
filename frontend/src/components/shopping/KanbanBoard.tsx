import { useState, useCallback } from 'react'
import { useShopping, useUpdateShoppingItem } from '../../hooks'
import { formatCurrency } from '../../utils/format'
import type { ShoppingItem, ItemStatus } from '../../types'
import styles from './KanbanBoard.module.css'

// ── Column definitions ────────────────────────────────────────────────────────

interface KanbanColumn {
  id: string
  label: string
  statuses: ItemStatus[]
  accentVar: string
}

const COLUMNS: KanbanColumn[] = [
  {
    id: 'watch',
    label: '🎯 À suivre',
    statuses: ['watch', 'defer'],
    accentVar: '--purple',
  },
  {
    id: 'buy',
    label: '🛒 À acheter',
    statuses: ['buy', 'flash-sale', 'preorder'],
    accentVar: '--cyan',
  },
  {
    id: 'crisis',
    label: '⚡ Urgence / Crise',
    statuses: ['crisis'],
    accentVar: '--coral',
  },
  {
    id: 'recommended',
    label: '✅ Recommandé',
    statuses: ['recommended'],
    accentVar: '--green',
  },
]

// Status dot color maps to CSS custom property names
const STATUS_DOT_CLASS: Record<ItemStatus, string> = {
  buy: styles.dotBuy,
  'flash-sale': styles.dotFlashSale,
  preorder: styles.dotPreorder,
  defer: styles.dotDefer,
  watch: styles.dotWatch,
  crisis: styles.dotCrisis,
  recommended: styles.dotRecommended,
  rejected: styles.dotRejected,
}

// First drop target for a column is determined by the column's statuses array
function defaultStatusForColumn(col: KanbanColumn): ItemStatus {
  return col.statuses[0]
}

// ── KanbanItemCard ────────────────────────────────────────────────────────────

interface KanbanItemCardProps {
  item: ShoppingItem
  currency: string
  onDragStart: (item: ShoppingItem) => void
  onDragEnd: () => void
  isDragging: boolean
}

function KanbanItemCard({
  item,
  currency,
  onDragStart,
  onDragEnd,
  isDragging,
}: KanbanItemCardProps) {
  const fmt = (n: number) => formatCurrency(n, currency)

  return (
    <div
      className={`${styles.card} ${isDragging ? styles.cardDragging : ''}`}
      draggable
      onDragStart={() => onDragStart(item)}
      onDragEnd={onDragEnd}
      role="listitem"
      aria-label={`${item.name} — ${fmt(item.price)}`}
    >
      {/* Drag handle */}
      <span className={styles.dragHandle} aria-hidden="true">⠿</span>

      <div className={styles.cardBody}>
        {/* Name */}
        <div className={styles.cardName}>{item.name}</div>

        {/* Merchant */}
        <div className={styles.cardMerchant}>{item.merchant}</div>

        {/* Bottom row: price + status dot + target chip */}
        <div className={styles.cardFooter}>
          <span className={styles.cardPrice}>{fmt(item.price)}</span>

          <span className={`${styles.statusDot} ${STATUS_DOT_CLASS[item.status]}`} title={item.status} />

          {item.targetPrice != null && (
            <span className={styles.targetChip}>
              ↯ {fmt(item.targetPrice)}
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

// ── KanbanColumn ──────────────────────────────────────────────────────────────

interface KanbanColumnProps {
  column: KanbanColumn
  items: ShoppingItem[]
  currency: string
  draggedItem: ShoppingItem | null
  onDragStart: (item: ShoppingItem) => void
  onDragEnd: () => void
  onDrop: (targetColumn: KanbanColumn) => void
}

function KanbanColumnView({
  column,
  items,
  currency,
  draggedItem,
  onDragStart,
  onDragEnd,
  onDrop,
}: KanbanColumnProps) {
  const [isOver, setIsOver] = useState(false)

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault()
    setIsOver(true)
  }

  function handleDragLeave() {
    setIsOver(false)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setIsOver(false)
    onDrop(column)
  }

  const canDrop =
    draggedItem !== null &&
    !column.statuses.includes(draggedItem.status)

  return (
    <div
      className={`${styles.column} ${isOver && canDrop ? styles.columnOver : ''}`}
      style={{ '--col-accent': `var(${column.accentVar})` } as React.CSSProperties}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      role="list"
      aria-label={column.label}
    >
      {/* Column header */}
      <div className={styles.columnHeader}>
        <span className={styles.columnLabel}>{column.label}</span>
        <span className={styles.columnCount}>{items.length}</span>
      </div>

      {/* Cards */}
      <div className={styles.columnCards}>
        {items.length === 0 ? (
          <div className={styles.emptyColumn}>
            <span className={styles.emptyIcon}>○</span>
            <span className={styles.emptyText}>Aucun article</span>
          </div>
        ) : (
          items.map((item) => (
            <KanbanItemCard
              key={item.id}
              item={item}
              currency={currency}
              onDragStart={onDragStart}
              onDragEnd={onDragEnd}
              isDragging={draggedItem?.id === item.id}
            />
          ))
        )}
      </div>
    </div>
  )
}

// ── KanbanBoard ───────────────────────────────────────────────────────────────

interface KanbanBoardProps {
  missionId: string
  currency: string
}

export function KanbanBoard({ missionId, currency }: KanbanBoardProps) {
  const { items, isLoading } = useShopping(missionId)
  const { updateItem } = useUpdateShoppingItem(missionId)
  const [draggedItem, setDraggedItem] = useState<ShoppingItem | null>(null)

  const handleDragStart = useCallback((item: ShoppingItem) => {
    setDraggedItem(item)
  }, [])

  const handleDragEnd = useCallback(() => {
    setDraggedItem(null)
  }, [])

  const handleDrop = useCallback(
    (targetColumn: KanbanColumn) => {
      if (!draggedItem) return
      const newStatus = defaultStatusForColumn(targetColumn)
      if (draggedItem.status === newStatus) return
      updateItem({ itemId: draggedItem.id, req: { status: newStatus } })
      setDraggedItem(null)
    },
    [draggedItem, updateItem],
  )

  if (isLoading) {
    return (
      <div className={styles.loadingRow}>
        {COLUMNS.map((col) => (
          <div key={col.id} className={styles.skeletonColumn} />
        ))}
      </div>
    )
  }

  return (
    <div className={styles.board}>
      {COLUMNS.map((col) => {
        const colItems = items.filter((item) => col.statuses.includes(item.status))
        return (
          <KanbanColumnView
            key={col.id}
            column={col}
            items={colItems}
            currency={currency}
            draggedItem={draggedItem}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            onDrop={handleDrop}
          />
        )
      })}
    </div>
  )
}
