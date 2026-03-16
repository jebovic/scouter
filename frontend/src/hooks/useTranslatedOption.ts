import { useTranslation } from 'react-i18next'
import type { Option, Attribute } from '../types'

export function useTranslatedOption(option: Option): Option {
  const { i18n } = useTranslation()
  const lang = i18n.language

  if (lang === 'en' || !option.translations?.[lang]) {
    return option
  }

  const blob = option.translations[lang]

  const mergedAttributes: Attribute[] = blob.attributes
    ? option.attributes.map((attr, i) => {
        const t = blob.attributes?.[i]
        if (!t) return attr
        if (attr.type === 'text') {
          return { ...attr, label: t.label, value: t.value }
        }
        return { ...attr, label: t.label }
      })
    : option.attributes

  return {
    ...option,
    ...(blob.name !== undefined && { name: blob.name }),
    ...(blob.notes !== undefined && { notes: blob.notes }),
    ...(blob.warnings !== undefined && { warnings: blob.warnings }),
    ...(blob.category !== undefined && { category: blob.category }),
    attributes: mergedAttributes,
  }
}
