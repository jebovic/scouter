import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TemplateGallery } from './TemplateGallery'
import type { Template } from '../../types'

const templates: Template[] = [
  {
    slug: 'test-custom-laptop',
    name: 'Laptop Purchase',
    icon: '💻',
    category: 'computing',
    description: 'Research and compare laptops',
    constraints: [],
    costCategories: ['hardware'],
    weightProfile: { price: 0.4, quality: 0.4, feature: 0.2 },
  },
  {
    slug: 'test-custom-trip',
    name: 'Trip Planning',
    icon: '✈️',
    category: 'travel',
    description: 'Plan and budget a trip',
    constraints: [],
    costCategories: ['flights', 'hotel'],
    weightProfile: { price: 0.5, quality: 0.3, feature: 0.2 },
  },
]

describe('TemplateGallery', () => {
  it('renders a card for each template', () => {
    render(<TemplateGallery templates={templates} onSelect={vi.fn()} />)
    expect(screen.getByText('Laptop Purchase')).toBeInTheDocument()
    expect(screen.getByText('Trip Planning')).toBeInTheDocument()
  })

  it('shows empty state when templates list is empty', () => {
    render(<TemplateGallery templates={[]} onSelect={vi.fn()} />)
    expect(screen.getByText(/no templates/i)).toBeInTheDocument()
  })

  it('shows loading skeleton when isLoading is true', () => {
    const { container } = render(
      <TemplateGallery templates={[]} onSelect={vi.fn()} isLoading />
    )
    expect(container.querySelector('[data-testid="skeleton"]')).toBeInTheDocument()
  })

  it('renders a heading or label for the gallery', () => {
    render(<TemplateGallery templates={templates} onSelect={vi.fn()} />)
    expect(screen.getByText(/quick.?start|templates/i)).toBeInTheDocument()
  })
})
