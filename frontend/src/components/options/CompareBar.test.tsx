import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CompareBar } from './CompareBar'

describe('CompareBar', () => {
  const noop = vi.fn()

  it('renders nothing when compareMode is false', () => {
    const { container } = render(
      <CompareBar
        isCompareMode={false}
        selectedCount={0}
        onCompare={noop}
        onClear={noop}
      />,
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders nothing when compareMode is true but no items selected', () => {
    const { container } = render(
      <CompareBar
        isCompareMode={true}
        selectedCount={0}
        onCompare={noop}
        onClear={noop}
      />,
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders with selection count when compareMode=true and count > 0', () => {
    render(
      <CompareBar
        isCompareMode={true}
        selectedCount={2}
        onCompare={noop}
        onClear={noop}
      />,
    )
    // The count is rendered in the bar somewhere
    expect(screen.getByRole('region', { name: /compare bar/i })).toBeInTheDocument()
    expect(screen.getByText('options sélectionnées')).toBeInTheDocument()
  })

  it('renders Comparer and Effacer buttons', () => {
    render(
      <CompareBar
        isCompareMode={true}
        selectedCount={2}
        onCompare={noop}
        onClear={noop}
      />,
    )
    expect(screen.getByRole('button', { name: /comparer/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /effacer/i })).toBeInTheDocument()
  })

  it('calls onCompare when Comparer button is clicked', async () => {
    const onCompare = vi.fn()
    render(
      <CompareBar
        isCompareMode={true}
        selectedCount={2}
        onCompare={onCompare}
        onClear={noop}
      />,
    )
    await userEvent.click(screen.getByRole('button', { name: /comparer/i }))
    expect(onCompare).toHaveBeenCalledOnce()
  })

  it('calls onClear when Effacer button is clicked', async () => {
    const onClear = vi.fn()
    render(
      <CompareBar
        isCompareMode={true}
        selectedCount={2}
        onCompare={noop}
        onClear={onClear}
      />,
    )
    await userEvent.click(screen.getByRole('button', { name: /effacer/i }))
    expect(onClear).toHaveBeenCalledOnce()
  })

  it('disables Comparer button when only 1 option selected', () => {
    render(
      <CompareBar
        isCompareMode={true}
        selectedCount={1}
        onCompare={noop}
        onClear={noop}
      />,
    )
    expect(screen.getByRole('button', { name: /comparer/i })).toBeDisabled()
  })

  it('enables Comparer button with 2 or more options', () => {
    render(
      <CompareBar
        isCompareMode={true}
        selectedCount={2}
        onCompare={noop}
        onClear={noop}
      />,
    )
    expect(screen.getByRole('button', { name: /comparer/i })).not.toBeDisabled()
  })
})
