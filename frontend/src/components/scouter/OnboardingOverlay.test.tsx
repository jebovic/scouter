import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OnboardingOverlay } from './OnboardingOverlay'

const defaultProps = {
  show: true,
  step: 0,
  totalSteps: 3,
  onNext: vi.fn(),
  onPrev: vi.fn(),
  onDismiss: vi.fn(),
}

describe('OnboardingOverlay', () => {
  it('does not render when show is false', () => {
    render(<OnboardingOverlay {...defaultProps} show={false} />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('renders when show is true', () => {
    render(<OnboardingOverlay {...defaultProps} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('renders step 0 content', () => {
    render(<OnboardingOverlay {...defaultProps} step={0} />)
    expect(screen.getByText(/Welcome to SCOUTER/i)).toBeInTheDocument()
  })

  it('renders step 1 content', () => {
    render(<OnboardingOverlay {...defaultProps} step={1} />)
    expect(screen.getByText(/Research & Compare/i)).toBeInTheDocument()
  })

  it('renders step 2 content', () => {
    render(<OnboardingOverlay {...defaultProps} step={2} />)
    expect(screen.getByText(/Track & Decide/i)).toBeInTheDocument()
  })

  it('renders step indicator dots', () => {
    render(<OnboardingOverlay {...defaultProps} />)
    const dots = screen.getAllByRole('button', { name: /step/i })
    expect(dots).toHaveLength(3)
  })

  it('calls onNext when Next is clicked', async () => {
    const onNext = vi.fn()
    render(<OnboardingOverlay {...defaultProps} step={0} onNext={onNext} />)
    await userEvent.click(screen.getByRole('button', { name: /next/i }))
    expect(onNext).toHaveBeenCalledOnce()
  })

  it('calls onPrev when Back is clicked on step > 0', async () => {
    const onPrev = vi.fn()
    render(<OnboardingOverlay {...defaultProps} step={1} onPrev={onPrev} />)
    await userEvent.click(screen.getByRole('button', { name: /back/i }))
    expect(onPrev).toHaveBeenCalledOnce()
  })

  it('calls onDismiss when Skip is clicked', async () => {
    const onDismiss = vi.fn()
    render(<OnboardingOverlay {...defaultProps} onDismiss={onDismiss} />)
    await userEvent.click(screen.getByRole('button', { name: /skip/i }))
    expect(onDismiss).toHaveBeenCalledOnce()
  })

  it('shows Done on last step instead of Next', () => {
    render(<OnboardingOverlay {...defaultProps} step={2} />)
    expect(screen.getByRole('button', { name: /done/i })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /next/i })).not.toBeInTheDocument()
  })

  it('calls onDismiss when Done is clicked on last step', async () => {
    const onDismiss = vi.fn()
    render(<OnboardingOverlay {...defaultProps} step={2} onDismiss={onDismiss} />)
    await userEvent.click(screen.getByRole('button', { name: /done/i }))
    expect(onDismiss).toHaveBeenCalledOnce()
  })
})
