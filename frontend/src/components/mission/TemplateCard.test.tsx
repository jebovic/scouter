import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TemplateCard } from './TemplateCard'
import type { Template } from '../../types'

const template: Template = {
  slug: 'laptop',
  name: 'Laptop Purchase',
  icon: '💻',
  category: 'computing',
  description: 'Research and compare laptops for work or gaming',
  constraints: [
    { key: 'min_ram', label: 'Min RAM (GB)', value: 16, type: 'hard' },
  ],
  costCategories: ['hardware', 'warranty'],
  weightProfile: { price: 0.4, quality: 0.4, feature: 0.2 },
  suggestedBudget: { min: 800, max: 2000 },
}

describe('TemplateCard', () => {
  it('renders template name and icon', () => {
    render(<TemplateCard template={template} onSelect={vi.fn()} />)
    expect(screen.getByText('Laptop Purchase')).toBeInTheDocument()
    expect(screen.getByText('💻')).toBeInTheDocument()
  })

  it('renders category', () => {
    render(<TemplateCard template={template} onSelect={vi.fn()} />)
    expect(screen.getByText(/computing/i)).toBeInTheDocument()
  })

  it('renders description', () => {
    render(<TemplateCard template={template} onSelect={vi.fn()} />)
    expect(screen.getByText('Research and compare laptops for work or gaming')).toBeInTheDocument()
  })

  it('renders suggested budget when present', () => {
    render(<TemplateCard template={template} onSelect={vi.fn()} />)
    expect(screen.getByText(/800/)).toBeInTheDocument()
    expect(screen.getByText(/2.?000/)).toBeInTheDocument()
  })

  it('does not render budget range when absent', () => {
    const nobudget: Template = { ...template, suggestedBudget: undefined }
    render(<TemplateCard template={nobudget} onSelect={vi.fn()} />)
    expect(screen.queryByText(/Budget:/i)).not.toBeInTheDocument()
  })

  it('calls onSelect when card or button is clicked', async () => {
    const onSelect = vi.fn()
    render(<TemplateCard template={template} onSelect={onSelect} />)
    await userEvent.click(screen.getByRole('button'))
    expect(onSelect).toHaveBeenCalledWith(template)
  })
})
