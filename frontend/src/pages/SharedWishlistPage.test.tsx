import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import SharedWishlistPage from './SharedWishlistPage'
import * as wishlistShare from '../utils/wishlistShare'

function renderWithRoute(search: string) {
  return render(
    <MemoryRouter initialEntries={[`/wishlist/shared${search}`]}>
      <Routes>
        <Route path="/wishlist/shared" element={<SharedWishlistPage />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('SharedWishlistPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('renders decoded wishlist items', () => {
    vi.spyOn(wishlistShare, 'decodeWishlist').mockReturnValue([
      { n: 'Laptop', p: 999, u: 'https://shop.com/laptop', c: 'EUR' },
      { n: 'Mouse', p: null, u: null, c: 'USD' },
    ])
    renderWithRoute('?data=abc123')
    expect(screen.getByText('Laptop')).toBeInTheDocument()
    expect(screen.getByText('Mouse')).toBeInTheDocument()
  })

  it('formats price with Intl.NumberFormat', () => {
    vi.spyOn(wishlistShare, 'decodeWishlist').mockReturnValue([
      { n: 'Camera', p: 499.99, u: null, c: 'EUR' },
    ])
    renderWithRoute('?data=abc')
    // Price should be formatted (locale-dependent, but must contain 499)
    expect(screen.getByText(/499/)).toBeInTheDocument()
  })

  it('renders a link when item has a URL', () => {
    vi.spyOn(wishlistShare, 'decodeWishlist').mockReturnValue([
      { n: 'Book', p: 15, u: 'https://bookshop.com', c: 'EUR' },
    ])
    renderWithRoute('?data=abc')
    const link = screen.getByRole('link', { name: /Book/i })
    expect(link).toHaveAttribute('href', 'https://bookshop.com')
  })

  it('shows error message for invalid data', () => {
    vi.spyOn(wishlistShare, 'decodeWishlist').mockReturnValue([])
    renderWithRoute('?data=bad-data')
    expect(screen.getByText(/invalide|expiré/i)).toBeInTheDocument()
  })

  it('shows error message when data param is missing', () => {
    vi.spyOn(wishlistShare, 'decodeWishlist').mockReturnValue([])
    renderWithRoute('')
    expect(screen.getByText(/invalide|expiré/i)).toBeInTheDocument()
  })

  it('shows CTA link to /wishlist', () => {
    vi.spyOn(wishlistShare, 'decodeWishlist').mockReturnValue([
      { n: 'Item', p: 10, u: null, c: 'EUR' },
    ])
    renderWithRoute('?data=abc')
    const cta = screen.getByRole('link', { name: /Créer ma liste Scouter/i })
    expect(cta).toHaveAttribute('href', '/wishlist')
  })
})
