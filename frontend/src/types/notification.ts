export type NotificationType = 'price_drop' | 'target_hit' | 'trend_alert'

export interface Notification {
  id: string
  missionId: string
  itemId?: string
  type: NotificationType
  title: string
  body: string
  read: boolean
  createdAt: string
}
