import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useOnboarding } from './useOnboarding'

const STORAGE_KEY = 'scouter_onboarding_dismissed'

describe('useOnboarding', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('shows onboarding when localStorage key is absent', () => {
    const { result } = renderHook(() => useOnboarding())
    expect(result.current.show).toBe(true)
  })

  it('hides onboarding when localStorage key is set', () => {
    localStorage.setItem(STORAGE_KEY, 'true')
    const { result } = renderHook(() => useOnboarding())
    expect(result.current.show).toBe(false)
  })

  it('starts at step 0', () => {
    const { result } = renderHook(() => useOnboarding())
    expect(result.current.step).toBe(0)
  })

  it('nextStep increments step', () => {
    const { result } = renderHook(() => useOnboarding())
    act(() => result.current.nextStep())
    expect(result.current.step).toBe(1)
  })

  it('nextStep clamps at totalSteps - 1', () => {
    const { result } = renderHook(() => useOnboarding())
    act(() => result.current.nextStep())
    act(() => result.current.nextStep())
    act(() => result.current.nextStep()) // should not go to 3
    expect(result.current.step).toBe(2)
  })

  it('prevStep decrements step', () => {
    const { result } = renderHook(() => useOnboarding())
    act(() => result.current.nextStep())
    act(() => result.current.prevStep())
    expect(result.current.step).toBe(0)
  })

  it('prevStep clamps at 0', () => {
    const { result } = renderHook(() => useOnboarding())
    act(() => result.current.prevStep())
    expect(result.current.step).toBe(0)
  })

  it('dismiss sets localStorage and hides onboarding', () => {
    const { result } = renderHook(() => useOnboarding())
    act(() => result.current.dismiss())
    expect(result.current.show).toBe(false)
    expect(localStorage.getItem(STORAGE_KEY)).toBe('true')
  })

  it('exposes totalSteps = 3', () => {
    const { result } = renderHook(() => useOnboarding())
    expect(result.current.totalSteps).toBe(3)
  })
})
