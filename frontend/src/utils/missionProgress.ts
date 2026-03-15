export interface MissionProgress {
  // Budget
  budgetTotal: number
  budgetSpent: number
  budgetPercent: number
  budgetStatus: 'on_track' | 'over_budget' | 'under_budget'

  // Readiness
  totalItems: number
  readyItems: number
  readinessPercent: number

  // Timeline
  daysElapsed: number
  daysTotal: number
  timePercent: number
  urgency: 'low' | 'medium' | 'high'

  // Verdict
  overallStatus: 'ahead' | 'on_track' | 'at_risk' | 'overdue'
}

export function computeMissionProgress(
  mission: { budget?: number | null; createdAt: string },
  items: Array<{ currentPrice: number; targetPrice?: number | null }>,
  targetDays = 90
): MissionProgress {
  // Budget calculations
  const budgetTotal = mission.budget ?? 0
  const budgetSpent = items.reduce((sum, item) => sum + item.currentPrice, 0)
  const budgetPercent = budgetTotal > 0 ? Math.min(100, (budgetSpent / budgetTotal) * 100) : 100

  // Determine budget status
  let budgetStatus: 'on_track' | 'over_budget' | 'under_budget'
  if (budgetSpent > budgetTotal) {
    budgetStatus = 'over_budget'
  } else if (budgetTotal > 0 && budgetSpent < budgetTotal * 0.5) {
    budgetStatus = 'under_budget'
  } else {
    budgetStatus = 'on_track'
  }

  // Readiness calculations
  const totalItems = items.length
  const readyItems = items.filter(
    item => !item.targetPrice || item.currentPrice <= item.targetPrice
  ).length
  const readinessPercent = totalItems > 0 ? (readyItems / totalItems) * 100 : 100

  // Time calculations
  const createdDate = new Date(mission.createdAt).getTime()
  const currentDate = new Date().getTime()
  const daysElapsed = Math.floor((currentDate - createdDate) / (24 * 60 * 60 * 1000))
  const daysTotal = targetDays
  const timePercent = Math.min(100, (daysElapsed / daysTotal) * 100)

  // Urgency
  let urgency: 'low' | 'medium' | 'high'
  if (timePercent > 75 && readinessPercent < 50) {
    urgency = 'high'
  } else if (timePercent > 50 && readinessPercent < 75) {
    urgency = 'medium'
  } else {
    urgency = 'low'
  }

  // Overall status
  let overallStatus: 'ahead' | 'on_track' | 'at_risk' | 'overdue'
  if (daysElapsed > daysTotal) {
    overallStatus = 'overdue'
  } else if (timePercent > 75 && readinessPercent < 50) {
    overallStatus = 'at_risk'
  } else if (readinessPercent > 75) {
    overallStatus = 'ahead'
  } else {
    overallStatus = 'on_track'
  }

  return {
    budgetTotal,
    budgetSpent,
    budgetPercent,
    budgetStatus,
    totalItems,
    readyItems,
    readinessPercent,
    daysElapsed,
    daysTotal,
    timePercent,
    urgency,
    overallStatus,
  }
}
