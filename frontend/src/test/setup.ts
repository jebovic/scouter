import '@testing-library/jest-dom'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'
import React from 'react'
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from '../i18n/en.json'

// Make React available globally for JSX in test files
Object.assign(globalThis, { React })

// Initialize i18next for tests so t() returns real translations
if (!i18n.isInitialized) {
  i18n
    .use(initReactI18next)
    .init({
      lng: 'en',
      fallbackLng: 'en',
      resources: { en: { translation: en } },
      interpolation: { escapeValue: false },
    })
}

// Clean up after each test
afterEach(() => {
  cleanup()
})

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

// Mock useSettings and useFormatCurrency so components using them don't need QueryClientProvider in tests
vi.mock('../hooks/useSettings', () => ({
  useSettings: () => ({ data: { currency: 'EUR', locale: 'fr-FR', llm_provider: 'ollama' }, isLoading: false }),
  useUpdateSettings: () => ({ mutate: vi.fn(), isSuccess: false, isPending: false }),
  useDeleteAllData: () => ({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('../hooks/useFormatCurrency', () => ({
  useFormatCurrency: () => ({
    fmt: (amount: number) => `€${amount.toFixed(2)}`,
    currency: 'EUR',
    locale: 'fr-FR',
  }),
}))

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})
