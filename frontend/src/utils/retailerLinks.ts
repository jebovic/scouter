export interface Retailer {
  id: string
  name: string
  logo: string
  buildSearchUrl: (query: string) => string
  primaryCategories: string[]
}

export const FRENCH_RETAILERS: Retailer[] = [
  {
    id: 'fnac',
    name: 'Fnac',
    logo: '🟠',
    buildSearchUrl: (q) =>
      `https://www.fnac.com/SearchResult/ResultSet.aspx?SCat=0&Search=${encodeURIComponent(q)}`,
    primaryCategories: ['electronics', 'computing', 'all'],
  },
  {
    id: 'darty',
    name: 'Darty',
    logo: '🔴',
    buildSearchUrl: (q) =>
      `https://www.darty.com/nav/recherche?text=${encodeURIComponent(q)}`,
    primaryCategories: ['electronics', 'computing', 'all'],
  },
  {
    id: 'boulanger',
    name: 'Boulanger',
    logo: '🔵',
    buildSearchUrl: (q) =>
      `https://www.boulanger.com/recherche/${encodeURIComponent(q)}`,
    primaryCategories: ['electronics', 'computing', 'all'],
  },
  {
    id: 'cdiscount',
    name: 'Cdiscount',
    logo: '🟡',
    buildSearchUrl: (q) =>
      `https://www.cdiscount.com/search/10/${encodeURIComponent(q)}.html`,
    primaryCategories: ['all'],
  },
  {
    id: 'amazon-fr',
    name: 'Amazon FR',
    logo: '📦',
    buildSearchUrl: (q) =>
      `https://www.amazon.fr/s?k=${encodeURIComponent(q)}`,
    primaryCategories: ['all'],
  },
  {
    id: 'ldlc',
    name: 'LDLC',
    logo: '💻',
    buildSearchUrl: (q) =>
      `https://www.ldlc.com/recherche/${encodeURIComponent(q)}/`,
    primaryCategories: ['computing', 'electronics'],
  },
  {
    id: 'leclerc',
    name: 'E.Leclerc',
    logo: '🔷',
    buildSearchUrl: (q) =>
      `https://www.e.leclerc/recherche?text=${encodeURIComponent(q)}`,
    primaryCategories: ['all'],
  },
  {
    id: 'rueducommerce',
    name: 'RueDuCommerce',
    logo: '🛒',
    buildSearchUrl: (q) =>
      `https://www.rueducommerce.fr/recherche.html?q=${encodeURIComponent(q)}`,
    primaryCategories: ['electronics', 'computing'],
  },
]

/**
 * Get retailers relevant to a mission category.
 * Always includes retailers tagged with "all".
 * Deduplicates so retailers tagged with both specific + "all" appear once.
 */
export function getRetailersForCategory(category: string): Retailer[] {
  // "all" means return every retailer
  if (category === 'all') {
    return [...FRENCH_RETAILERS]
  }

  const seen = new Set<string>()
  const result: Retailer[] = []

  for (const retailer of FRENCH_RETAILERS) {
    const matches =
      retailer.primaryCategories.includes(category) ||
      retailer.primaryCategories.includes('all')

    if (matches && !seen.has(retailer.id)) {
      seen.add(retailer.id)
      result.push(retailer)
    }
  }

  return result
}

export interface RetailerLink {
  retailer: Retailer
  url: string
}

/**
 * Build retailer search links for a given query and optional category.
 * Falls back to "all" retailers when category is not provided.
 */
export function buildRetailerLinks(
  query: string,
  category: string = 'all',
): RetailerLink[] {
  const retailers = getRetailersForCategory(category)
  return retailers.map((retailer) => ({
    retailer,
    url: retailer.buildSearchUrl(query),
  }))
}
