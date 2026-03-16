import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Logo } from './Logo'

describe('Logo', () => {
  it('renders wordmark text by default', () => {
    render(<Logo />)
    expect(screen.getByText('SCOUTER')).toBeInTheDocument()
  })

  it('hides icon from accessibility tree when wordmark is visible', () => {
    const { container } = render(<Logo />)
    const svg = container.querySelector('svg')
    expect(svg).toHaveAttribute('aria-hidden', 'true')
  })

  it('renders tagline at size="lg" when showTagline=true', () => {
    render(<Logo size="lg" showTagline />)
    expect(screen.getByText('spending intelligence')).toBeInTheDocument()
  })

  it('does not render tagline at size="md" even when showTagline=true', () => {
    render(<Logo size="md" showTagline />)
    expect(screen.queryByText('spending intelligence')).not.toBeInTheDocument()
  })

  it('does not render tagline at size="sm" even when showTagline=true', () => {
    render(<Logo size="sm" showTagline />)
    expect(screen.queryByText('spending intelligence')).not.toBeInTheDocument()
  })

  it('iconOnly: no wordmark, SVG has role="img" and default aria-label', () => {
    const { container } = render(<Logo iconOnly />)
    expect(screen.queryByText('SCOUTER')).not.toBeInTheDocument()
    const svg = container.querySelector('svg')
    expect(svg).toHaveAttribute('role', 'img')
    expect(svg).toHaveAttribute('aria-label', 'SCOUTER')
  })

  it('iconOnly: custom aria-label overrides default', () => {
    const { container } = render(<Logo iconOnly aria-label="Scout Logo" />)
    const svg = container.querySelector('svg')
    expect(svg).toHaveAttribute('aria-label', 'Scout Logo')
  })

  it('multiple instances produce unique gradient ids', () => {
    const { container } = render(
      <>
        <Logo />
        <Logo />
      </>
    )
    const gradients = container.querySelectorAll('linearGradient')
    const ids = Array.from(gradients).map(g => g.id)
    const unique = new Set(ids)
    expect(unique.size).toBe(ids.length)
  })
})
