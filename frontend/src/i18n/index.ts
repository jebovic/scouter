import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './en.json'
import fr from './fr.json'
import de from './de.json'
import { localeToLang } from '../utils/currency'

i18n
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      fr: { translation: fr },
      de: { translation: de },
    },
    lng: navigator.language.startsWith('fr') ? 'fr' : navigator.language.startsWith('de') ? 'de' : 'en',
    fallbackLng: 'en',
    interpolation: { escapeValue: false },
  })

export function syncLanguageFromSettings(locale: string): void {
  const lang = localeToLang(locale)
  if (i18n.language !== lang) {
    i18n.changeLanguage(lang)
  }
}

export default i18n
