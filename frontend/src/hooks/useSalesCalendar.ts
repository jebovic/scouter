import { useMemo } from 'react'
import {
  getUpcomingEvents,
  getNextBestEvent,
  getEventsForCategory,
} from '../utils/frenchSalesCalendar'

export function useSalesCalendar(category = '') {
  return useMemo(
    () => ({
      upcomingEvents: getUpcomingEvents(),
      nextBestEvent: getNextBestEvent(category),
      categoryEvents: getEventsForCategory(category),
    }),
    [category],
  )
}
