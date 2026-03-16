import { renderHook } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { useTranslatedOption } from './useTranslatedOption'
import type { Option } from '../types'

// Mock react-i18next — language is controlled per describe via mockReturnValue
const mockLanguage = { language: 'fr' }
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ i18n: mockLanguage }),
}))

const baseOption: Option = {
  id: '123',
  missionId: 'mission-1',
  name: 'Test Option',
  category: 'Electronics',
  badge: 'recommended',
  attributes: [
    { key: 'color', label: 'Color', value: 'Red', type: 'text' },
    { key: 'price', label: 'Price', value: 100, type: 'price' },
  ],
  warnings: ['Check compatibility'],
  pinned: false,
  rejected: false,
  createdAt: '2026-01-01T00:00:00Z',
}

describe('useTranslatedOption', () => {
  it('returns option unchanged when no translations', () => {
    const { result } = renderHook(() => useTranslatedOption(baseOption))
    expect(result.current).toBe(baseOption)
  })

  it('merges translation when available', () => {
    const option: Option = {
      ...baseOption,
      translations: {
        fr: {
          name: 'Option Test',
          category: 'Électronique',
          warnings: ['Vérifier la compatibilité'],
          attributes: [
            { label: 'Couleur', value: 'Rouge' },
            { label: 'Prix', value: '100' },
          ],
        },
      },
    }
    const { result } = renderHook(() => useTranslatedOption(option))
    expect(result.current.name).toBe('Option Test')
    expect(result.current.category).toBe('Électronique')
    expect(result.current.warnings).toEqual(['Vérifier la compatibilité'])
    // text attribute: both label and value translated
    expect(result.current.attributes[0].label).toBe('Couleur')
    expect(result.current.attributes[0].value).toBe('Rouge')
    // price attribute: only label translated, value kept as number
    expect(result.current.attributes[1].label).toBe('Prix')
    expect(result.current.attributes[1].value).toBe(100)
  })

  it('returns option unchanged when language is en even if fr translation exists', () => {
    mockLanguage.language = 'en'
    const option: Option = {
      ...baseOption,
      translations: { fr: { name: 'Option Test' } },
    }
    const { result } = renderHook(() => useTranslatedOption(option))
    expect(result.current).toBe(option)
    mockLanguage.language = 'fr' // restore for other tests
  })
})
