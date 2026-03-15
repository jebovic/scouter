import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { VATCalculator } from './VATCalculator'

describe('VATCalculator', () => {
  const mockOnClose = vi.fn()

  beforeEach(() => {
    mockOnClose.mockClear()
  })

  it('renders the VAT calculator panel', () => {
    render(<VATCalculator onClose={mockOnClose} />)
    expect(screen.getByText(/Taux TVA/i)).toBeInTheDocument()
    expect(screen.getByText(/Prix HT/i)).toBeInTheDocument()
    expect(screen.getByText(/Prix TTC/i)).toBeInTheDocument()
  })

  it('initializes with 20% rate selected', () => {
    render(<VATCalculator onClose={mockOnClose} />)
    const rateSelect = screen.getByLabelText(/Sélectionner le taux TVA/i) as HTMLSelectElement
    expect(rateSelect.value).toBe('0.2')
  })

  it('calculates TTC when HT is entered', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const htInput = screen.getByLabelText(/Prix HT/i) as HTMLInputElement
    await user.type(htInput, '100')

    // TTC should be 120 at 20% rate
    const ttcInput = screen.getByLabelText(/Prix TTC/i) as HTMLInputElement
    expect(ttcInput.value).toBe('120.00')
  })

  it('calculates HT when TTC is entered', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const ttcInput = screen.getByLabelText(/Prix TTC/i) as HTMLInputElement
    await user.type(ttcInput, '120')

    // HT should be 100 at 20% rate
    const htInput = screen.getByLabelText(/Prix HT/i) as HTMLInputElement
    expect(htInput.value).toBe('100.00')
  })

  it('shows VAT amount when HT is entered', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const htInput = screen.getByLabelText(/Prix HT/i)
    await user.type(htInput, '100')

    // Should show VAT amount of 20
    expect(screen.getByText(/Montant TVA/i)).toBeInTheDocument()
    expect(screen.getByText('20.00')).toBeInTheDocument()
  })

  it('handles different VAT rates', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const rateSelect = screen.getByLabelText(/Sélectionner le taux TVA/i)
    const htInput = screen.getByLabelText(/Prix HT/i) as HTMLInputElement

    // Enter 100 HT
    await user.type(htInput, '100')

    // Switch to 10% rate
    await user.selectOptions(rateSelect, '0.1')

    // TTC should now be 110
    const ttcInput = screen.getByLabelText(/Prix TTC/i) as HTMLInputElement
    expect(ttcInput.value).toBe('110.00')
  })

  it('initializes with priceHT prop', () => {
    render(<VATCalculator priceHT={50} onClose={mockOnClose} />)
    const htInput = screen.getByLabelText(/Prix HT/i) as HTMLInputElement
    expect(htInput.value).toBe('50')
  })

  it('calls onClose when Fermer button is clicked', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const closeBtn = screen.getByText(/Fermer/i)
    await user.click(closeBtn)

    expect(mockOnClose).toHaveBeenCalledOnce()
  })

  it('shows rate description', () => {
    render(<VATCalculator onClose={mockOnClose} />)
    // Check for the description section (not the option element)
    const descriptions = screen.getAllByText(/électronique, électroménager/)
    expect(descriptions.length).toBeGreaterThan(0)
  })

  it('handles clearing inputs', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const htInput = screen.getByLabelText(/Prix HT/i) as HTMLInputElement
    await user.type(htInput, '100')
    await user.clear(htInput)

    const ttcInput = screen.getByLabelText(/Prix TTC/i) as HTMLInputElement
    expect(ttcInput.value).toBe('')
  })

  it('handles decimal values', async () => {
    const user = userEvent.setup()
    render(<VATCalculator onClose={mockOnClose} />)

    const htInput = screen.getByLabelText(/Prix HT/i)
    await user.type(htInput, '19.99')

    const ttcInput = screen.getByLabelText(/Prix TTC/i) as HTMLInputElement
    // 19.99 * 1.20 = 23.988, formatted to 23.99
    expect(parseFloat(ttcInput.value)).toBeCloseTo(23.99, 2)
  })
})
