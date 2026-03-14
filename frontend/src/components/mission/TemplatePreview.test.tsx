import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TemplatePreview } from './TemplatePreview'
import type { Template } from '../../types'

const template: Template = {
  slug: 'laptop',
  name: 'Laptop Purchase',
  icon: '💻',
  category: 'computing',
  description: 'Research and compare laptops for work or gaming',
  constraints: [
    { key: 'min_ram', label: 'Min RAM (GB)', value: 16, type: 'hard' },
    { key: 'os', label: 'Operating System', value: 'macOS', type: 'soft' },
  ],
  costCategories: ['hardware', 'warranty'],
  weightProfile: { price: 0.4, quality: 0.4, feature: 0.2 },
  suggestedBudget: { min: 800, max: 2000 },
}

describe('TemplatePreview', () => {
  it('renders template name and icon', () => {
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText('Laptop Purchase')).toBeInTheDocument()
    expect(screen.getByText('💻')).toBeInTheDocument()
  })

  it('renders description', () => {
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText('Research and compare laptops for work or gaming')).toBeInTheDocument()
  })

  it('renders constraint labels', () => {
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText('Min RAM (GB)')).toBeInTheDocument()
    expect(screen.getByText('Operating System')).toBeInTheDocument()
  })

  it('renders cost categories', () => {
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText(/hardware/i)).toBeInTheDocument()
    expect(screen.getByText(/warranty/i)).toBeInTheDocument()
  })

  it('renders suggested budget when present', () => {
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText(/800/)).toBeInTheDocument()
    expect(screen.getByText(/2.?000/)).toBeInTheDocument()
  })

  it('calls onApply when Apply button is clicked', async () => {
    const onApply = vi.fn()
    render(<TemplatePreview template={template} onApply={onApply} onClose={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: /apply/i }))
    expect(onApply).toHaveBeenCalledWith(template)
  })

  it('calls onClose when Close/Cancel button is clicked', async () => {
    const onClose = vi.fn()
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={onClose} />)
    await userEvent.click(screen.getByRole('button', { name: /close|cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('renders weight profile labels', () => {
    render(<TemplatePreview template={template} onApply={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText(/price/i)).toBeInTheDocument()
    expect(screen.getByText(/quality/i)).toBeInTheDocument()
    expect(screen.getByText(/feature/i)).toBeInTheDocument()
  })
})
