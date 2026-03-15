/**
 * Map a settings locale (e.g. 'fr-FR', 'en-US') to an i18next language code.
 */
export function localeToLang(locale: string): string {
  if (locale.startsWith('fr')) return 'fr'
  if (locale.startsWith('de')) return 'de'
  return 'en'
}

/**
 * Format a monetary amount with the given currency code and locale.
 * Falls back to 'en-US' / 'USD' if the inputs are invalid.
 */
export function formatCurrency(amount: number, currency: string, locale: string): string {
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency,
      maximumFractionDigits: 2,
    }).format(amount)
  } catch {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 2,
    }).format(amount)
  }
}
