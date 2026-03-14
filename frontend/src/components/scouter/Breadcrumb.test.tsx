import { describe, it, expect } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '../../test/utils'
import { Breadcrumb } from './Breadcrumb'

describe('Breadcrumb', () => {
  it('renders HQ crumb on root route', () => {
    renderWithProviders(<Breadcrumb />, { initialEntries: ['/'] })
    expect(screen.getByText('HQ')).toBeInTheDocument()
  })

  it('renders mission name crumb when provided', () => {
    renderWithProviders(<Breadcrumb missionName="Gaming PC" />, {
      initialEntries: ['/missions/gaming-pc'],
    })
    expect(screen.getByText('HQ')).toBeInTheDocument()
    expect(screen.getByText('Gaming PC')).toBeInTheDocument()
  })

  it('renders options sub-crumb', () => {
    renderWithProviders(<Breadcrumb missionName="Gaming PC" subPage="Options" />, {
      initialEntries: ['/missions/gaming-pc/options'],
    })
    expect(screen.getByText('Options')).toBeInTheDocument()
  })

  it('renders shopping sub-crumb', () => {
    renderWithProviders(<Breadcrumb missionName="Gaming PC" subPage="Shopping" />, {
      initialEntries: ['/missions/gaming-pc/shopping'],
    })
    expect(screen.getByText('Shopping')).toBeInTheDocument()
  })

  it('HQ crumb is a link to /', () => {
    renderWithProviders(<Breadcrumb missionName="Gaming PC" />, {
      initialEntries: ['/missions/gaming-pc'],
    })
    expect(screen.getByRole('link', { name: 'HQ' })).toHaveAttribute('href', '/')
  })

  it('does not render sub-crumb when not provided', () => {
    renderWithProviders(<Breadcrumb missionName="Gaming PC" />, {
      initialEntries: ['/missions/gaming-pc'],
    })
    expect(screen.queryByText('Options')).not.toBeInTheDocument()
    expect(screen.queryByText('Shopping')).not.toBeInTheDocument()
  })
})
