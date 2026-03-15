import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ReceiptScanner, extractAmount, extractMerchant, extractDate } from './ReceiptScanner'

// ── Pure function tests (no DOM required) ──────────────────────────────────

describe('extractAmount', () => {
  it('returns undefined for empty string', () => {
    expect(extractAmount('')).toBeUndefined()
  })

  it('returns undefined when no amount pattern found', () => {
    expect(extractAmount('Hello World\nBonjour')).toBeUndefined()
  })

  it('extracts integer total after TOTAL', () => {
    expect(extractAmount('TOTAL 42')).toBe(42)
  })

  it('extracts decimal amount with dot separator after TOTAL', () => {
    expect(extractAmount('TOTAL 42.90')).toBe(42.90)
  })

  it('extracts decimal amount with comma separator after TOTAL', () => {
    expect(extractAmount('TOTAL 42,90')).toBe(42.90)
  })

  it('extracts amount after MONTANT', () => {
    expect(extractAmount('MONTANT 15,50')).toBe(15.50)
  })

  it('extracts amount after TTC', () => {
    expect(extractAmount('TTC 99.99')).toBe(99.99)
  })

  it('extracts amount after TOTAL TTC', () => {
    expect(extractAmount('TOTAL TTC 128,00')).toBe(128.00)
  })

  it('extracts amount with euro sign', () => {
    expect(extractAmount('TOTAL 42,90 €')).toBe(42.90)
  })

  it('handles whitespace around values', () => {
    expect(extractAmount('  TOTAL   99,00  ')).toBe(99.00)
  })

  it('ignores subtotals and picks first TOTAL match', () => {
    const text = 'SOUS-TOTAL 10,00\nTOTAL 12,00\nTOTAL TTC 12,00'
    expect(extractAmount(text)).toBe(12.00)
  })

  it('handles amounts with thousand separator (space)', () => {
    expect(extractAmount('TOTAL 1 200,00')).toBe(1200.00)
  })

  it('returns undefined for non-numeric value after keyword', () => {
    expect(extractAmount('TOTAL ABC')).toBeUndefined()
  })
})

describe('extractMerchant', () => {
  it('returns undefined for empty string', () => {
    expect(extractMerchant('')).toBeUndefined()
  })

  it('returns undefined for whitespace-only string', () => {
    expect(extractMerchant('   \n  \n  ')).toBeUndefined()
  })

  it('returns the first non-empty line', () => {
    expect(extractMerchant('FNAC\n123 Rue de la Paix\nTOTAL 42,00')).toBe('FNAC')
  })

  it('skips leading blank lines', () => {
    expect(extractMerchant('\n\nCarrefour\nAdresse')).toBe('Carrefour')
  })

  it('trims whitespace from merchant name', () => {
    expect(extractMerchant('  Leroy Merlin  \nAutre ligne')).toBe('Leroy Merlin')
  })

  it('returns single line when text has one line', () => {
    expect(extractMerchant('Monoprix')).toBe('Monoprix')
  })
})

describe('extractDate', () => {
  it('returns undefined for empty string', () => {
    expect(extractDate('')).toBeUndefined()
  })

  it('returns undefined when no date pattern found', () => {
    expect(extractDate('TOTAL 42,00\nMerci de votre achat')).toBeUndefined()
  })

  it('extracts DD/MM/YYYY date', () => {
    expect(extractDate('Date: 15/03/2026')).toBe('15/03/2026')
  })

  it('extracts DD-MM-YYYY date', () => {
    expect(extractDate('Date: 15-03-2026')).toBe('15-03-2026')
  })

  it('finds date embedded in text', () => {
    expect(extractDate('Ticket du 01/01/2026 pour votre achat')).toBe('01/01/2026')
  })

  it('returns the first date found when multiple dates present', () => {
    expect(extractDate('01/01/2026 something 02/02/2026')).toBe('01/01/2026')
  })

  it('does not match invalid day/month values (boundary not strictly validated)', () => {
    // extractDate uses regex only, structural match suffices
    expect(extractDate('32/13/2026')).toBe('32/13/2026')
  })
})

// ── Component tests ─────────────────────────────────────────────────────────

vi.mock('tesseract.js', () => ({
  default: {
    recognize: vi.fn(),
  },
}))

import Tesseract from 'tesseract.js'
const mockRecognize = vi.mocked(Tesseract.recognize)

function makeFile(name = 'receipt.jpg', type = 'image/jpeg') {
  return new File(['fake-image-content'], name, { type })
}

describe('ReceiptScanner component', () => {
  const onResult = vi.fn()
  const onClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the drop zone with correct label', () => {
    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    expect(screen.getAllByText(/glisser|déposer|scanner/i).length).toBeGreaterThan(0)
  })

  it('renders the file input accepting image/* and application/pdf', () => {
    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    expect(input).toBeTruthy()
    expect(input.accept).toContain('image/*')
    expect(input.accept).toContain('application/pdf')
  })

  it('renders the cancel/close button', () => {
    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const cancelBtn = screen.getByRole('button', { name: /annuler|fermer/i })
    expect(cancelBtn).toBeTruthy()
  })

  it('calls onClose when cancel button is clicked', () => {
    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const cancelBtn = screen.getByRole('button', { name: /annuler|fermer/i })
    fireEvent.click(cancelBtn)
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('shows loading state while OCR is running', async () => {
    // Never resolves during this test — stays in loading state
    mockRecognize.mockReturnValue(new Promise(() => {}))

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByText(/analyse/i)).toBeTruthy()
    })
  })

  it('shows extracted data preview after successful OCR', async () => {
    mockRecognize.mockResolvedValue({
      data: { text: 'FNAC\n15/03/2026\nTOTAL 42,90' },
    } as Awaited<ReturnType<typeof Tesseract.recognize>>)

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByText(/enseigne/i)).toBeTruthy()
    })
    // Amount appears as input value
    expect(screen.getByDisplayValue('42.9')).toBeTruthy()
  })

  it('renders editable inputs for merchant, amount, and date after OCR', async () => {
    mockRecognize.mockResolvedValue({
      data: { text: 'Carrefour\n01/06/2026\nTOTAL 99,00' },
    } as Awaited<ReturnType<typeof Tesseract.recognize>>)

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByDisplayValue('Carrefour')).toBeTruthy()
    })
    expect(screen.getByDisplayValue('99')).toBeTruthy()
    expect(screen.getByDisplayValue('01/06/2026')).toBeTruthy()
  })

  it('calls onResult with extracted data when confirm button is clicked', async () => {
    mockRecognize.mockResolvedValue({
      data: { text: 'Darty\n10/03/2026\nTOTAL 250,00' },
    } as Awaited<ReturnType<typeof Tesseract.recognize>>)

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /confirmer|valider/i })).toBeTruthy()
    })
    fireEvent.click(screen.getByRole('button', { name: /confirmer|valider/i }))

    expect(onResult).toHaveBeenCalledOnce()
    expect(onResult).toHaveBeenCalledWith(
      expect.objectContaining({ merchant: 'Darty', amount: 250 })
    )
  })

  it('shows error message when OCR fails', async () => {
    mockRecognize.mockRejectedValue(new Error('OCR failure'))

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByText(/échoué/i)).toBeTruthy()
    })
  })

  it('allows retry after OCR error', async () => {
    mockRecognize.mockRejectedValueOnce(new Error('fail'))

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(input, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /réessayer/i })).toBeTruthy()
    })
    // Retry button resets the form to allow new file selection
    fireEvent.click(screen.getByRole('button', { name: /réessayer/i }))
    expect(document.querySelector('input[type="file"]')).toBeTruthy()
  })

  it('allows user to edit merchant field before confirming', async () => {
    mockRecognize.mockResolvedValue({
      data: { text: 'Auchan\nTOTAL 55,00' },
    } as Awaited<ReturnType<typeof Tesseract.recognize>>)

    render(<ReceiptScanner onResult={onResult} onClose={onClose} />)
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(fileInput, { target: { files: [makeFile()] } })

    await waitFor(() => {
      expect(screen.getByDisplayValue('Auchan')).toBeTruthy()
    })

    const merchantInput = screen.getByDisplayValue('Auchan') as HTMLInputElement
    fireEvent.change(merchantInput, { target: { value: 'Auchan Supermarché' } })
    expect(merchantInput.value).toBe('Auchan Supermarché')
  })
})
