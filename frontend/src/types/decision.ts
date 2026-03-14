export interface ScoreResult {
  optionId: string
  optionName: string
  score: number
  priceScore: number
  qualScore: number
  featScore: number
  eliminated: boolean
  eliminReason?: string
}

export interface Decision {
  id: string
  missionId: string
  scores: ScoreResult[]
  summary: string
  createdAt: string
}
