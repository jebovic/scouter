import { useMemo } from 'react'
import { useMission } from './useMission'
import { useOptions } from './useOptions'
import { useShopping } from './useShopping'

export type TimelineEventType =
  | 'mission_created'
  | 'option_added'
  | 'research_run'
  | 'item_added'
  | 'item_purchased'
  | 'phase_changed'

export interface TimelineEvent {
  id: string
  type: TimelineEventType
  date: Date
  title: string
  description?: string
  icon: string
  entityId?: string
}

const ICON_MAP: Record<TimelineEventType, string> = {
  mission_created: '🚀',
  option_added: '📦',
  research_run: '🔬',
  item_added: '🛒',
  item_purchased: '✅',
  phase_changed: '🔄',
}

export function useTimeline(missionId: string, missionSlug: string) {
  const { mission, isLoading: missionLoading } = useMission(missionSlug)
  const { options, isLoading: optionsLoading } = useOptions(missionId)
  const { items, isLoading: shoppingLoading } = useShopping(missionId)

  const events = useMemo(() => {
    const events: TimelineEvent[] = []

    // Mission created event
    if (mission) {
      events.push({
        id: `mission_created-${new Date(mission.createdAt).toISOString()}`,
        type: 'mission_created',
        date: new Date(mission.createdAt),
        title: 'Mission créée',
        icon: ICON_MAP.mission_created,
      })
    }

    // Option added events
    if (options && options.length > 0) {
      options.forEach((option) => {
        events.push({
          id: `option_added-${option.id}`,
          type: 'option_added',
          date: new Date(option.createdAt),
          title: `Option ajoutée: ${option.name}`,
          description: option.category,
          icon: ICON_MAP.option_added,
          entityId: option.id,
        })
      })
    }

    // Shopping items events
    if (items && items.length > 0) {
      items.forEach((item) => {
        // Item added event
        events.push({
          id: `item_added-${item.id}`,
          type: 'item_added',
          date: new Date(item.createdAt),
          title: `Article ajouté: ${item.name}`,
          description: `${item.merchant}`,
          icon: ICON_MAP.item_added,
          entityId: item.id,
        })

        // Item purchased event (if status is buy - indicating a completed purchase)
        if (item.status === 'buy') {
          events.push({
            id: `item_purchased-${item.id}`,
            type: 'item_purchased',
            date: new Date(item.createdAt),
            title: `Article acheté: ${item.name}`,
            description: item.merchant,
            icon: ICON_MAP.item_purchased,
            entityId: item.id,
          })
        }
      })
    }

    // Sort newest first
    events.sort((a, b) => b.date.getTime() - a.date.getTime())

    return events
  }, [mission, options, items])

  return {
    events,
    isLoading: missionLoading || optionsLoading || shoppingLoading,
  }
}
