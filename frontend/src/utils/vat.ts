/**
 * French VAT (TVA) utilities for HT/TTC price conversions
 * HT = Price Hors Taxe (excluding tax)
 * TTC = Price Toutes Taxes Comprises (including tax)
 */

export type VATRate = 0.20 | 0.10 | 0.055 | 0.021

export interface VATRateOption {
  label: string
  rate: VATRate
  description: string
}

export const VAT_RATES: VATRateOption[] = [
  {
    label: '20%',
    rate: 0.20,
    description: 'Taux normal (électronique, électroménager...)',
  },
  {
    label: '10%',
    rate: 0.10,
    description: 'Taux intermédiaire (restauration, transports...)',
  },
  {
    label: '5,5%',
    rate: 0.055,
    description: 'Taux réduit (alimentation, livres...)',
  },
  {
    label: '2,1%',
    rate: 0.021,
    description: 'Taux super réduit (médicaments...)',
  },
]

/**
 * Convert HT (excluding tax) to TTC (including tax)
 * Formula: TTC = HT × (1 + rate)
 */
export function htToTtc(ht: number, rate: VATRate): number {
  return ht * (1 + rate)
}

/**
 * Convert TTC (including tax) to HT (excluding tax)
 * Formula: HT = TTC ÷ (1 + rate)
 */
export function ttcToHt(ttc: number, rate: VATRate): number {
  return ttc / (1 + rate)
}

/**
 * Calculate the VAT amount from HT
 * Formula: VAT = HT × rate
 */
export function vatAmount(ht: number, rate: VATRate): number {
  return ht * rate
}
