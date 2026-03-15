export const queryKeys = {
  missions: {
    all: () => ['missions'] as const,
    detail: (slug: string) => ['missions', slug] as const,
    export: (slug: string) => ['missions', slug, 'export'] as const,
  },
  options: {
    all: (missionSlug: string) => ['options', missionSlug] as const,
    detail: (id: string) => ['options', id] as const,
    similar: (id: string) => ['options', id, 'similar'] as const,
  },
  shopping: {
    all: (missionSlug: string) => ['shopping', missionSlug] as const,
    dealScore: (itemId: string) => ['deal-score', itemId] as const,
    priceHistory: (itemId: string) => ['price-history', itemId] as const,
  },
  notifications: {
    all: () => ['notifications'] as const,
    unreadCount: () => ['notifications', 'unread-count'] as const,
  },
  search: (query: string) => ['search', query] as const,
  templates: {
    all: () => ['templates'] as const,
    detail: (id: string) => ['templates', id] as const,
  },
  settings: () => ['settings'] as const,
  stats: () => ['stats'] as const,
  purchase: (missionSlug: string) => ['purchase', missionSlug] as const,
  health: () => ['health', 'llm'] as const,
}
