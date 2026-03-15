export interface LoyaltyProgram {
  id: string
  name: string
  color: string
  icon: string
  pointsPerEuro: number
  eurPerPoint: number
  milestones: { points: number; reward: string }[]
}

export const LOYALTY_PROGRAMS: LoyaltyProgram[] = [
  {
    id: 'fnac',
    name: 'Fnac+',
    color: '#f7a600',
    icon: '🎵',
    pointsPerEuro: 1,
    eurPerPoint: 0.01,
    milestones: [
      { points: 500, reward: '5€ de réduction' },
      { points: 1000, reward: '10€ de réduction' },
      { points: 2000, reward: '20€ + cadeau fidélité' },
    ],
  },
  {
    id: 'cdiscount',
    name: 'Club Cdiscount',
    color: '#e2001a',
    icon: '🛒',
    pointsPerEuro: 10,
    eurPerPoint: 0.005,
    milestones: [
      { points: 1000, reward: '5€ de bon d\'achat' },
      { points: 2000, reward: '10€ de bon d\'achat' },
      { points: 5000, reward: '25€ de bon d\'achat' },
    ],
  },
  {
    id: 'boulanger',
    name: 'Boulanger Fidélité',
    color: '#003f8a',
    icon: '🔵',
    pointsPerEuro: 1,
    eurPerPoint: 0.01,
    milestones: [
      { points: 300, reward: '3€ de réduction' },
      { points: 1000, reward: '10€ de réduction' },
      { points: 3000, reward: '30€ de réduction premium' },
    ],
  },
  {
    id: 'darty',
    name: 'Club Darty',
    color: '#e40032',
    icon: '🔴',
    pointsPerEuro: 1,
    eurPerPoint: 0.01,
    milestones: [
      { points: 200, reward: '2€ de remise' },
      { points: 500, reward: '5€ de remise' },
      { points: 1000, reward: '10€ de remise exclusive' },
    ],
  },
  {
    id: 'amazon',
    name: 'Amazon Points',
    color: '#ff9900',
    icon: '📦',
    pointsPerEuro: 3,
    eurPerPoint: 0.01,
    milestones: [
      { points: 500, reward: '5€ d\'avoir' },
      { points: 1000, reward: '10€ d\'avoir' },
    ],
  },
]

export function getNextMilestone(
  program: LoyaltyProgram,
  currentPoints: number
): { points: number; reward: string } | null {
  return program.milestones.find((m) => m.points > currentPoints) ?? null
}

export function getEquivalentValue(program: LoyaltyProgram, points: number): number {
  return points * program.eurPerPoint
}
