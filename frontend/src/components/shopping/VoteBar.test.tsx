import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { VoteBar } from './VoteBar'

// Mock the hooks
vi.mock('../../hooks', () => ({
  useVotes: vi.fn(),
  useRecordVote: vi.fn(),
}))

import { useVotes, useRecordVote } from '../../hooks'

const mockUseVotes = vi.mocked(useVotes)
const mockUseRecordVote = vi.mocked(useRecordVote)

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
}

function renderComponent(optionId = 'opt-1', collaboratorId: string | null = 'collab-1') {
  const queryClient = createQueryClient()

  // Set or clear collaborator in localStorage
  if (collaboratorId) {
    localStorage.setItem('scouter_collaborator_id', collaboratorId)
  } else {
    localStorage.removeItem('scouter_collaborator_id')
  }

  return render(
    <QueryClientProvider client={queryClient}>
      <VoteBar optionId={optionId} />
    </QueryClientProvider>
  )
}

describe('VoteBar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders loading state', () => {
    mockUseVotes.mockReturnValue({
      data: undefined,
      isLoading: true,
      isPending: true,
      status: 'pending',
      isError: false,
      error: null,
      dataUpdatedAt: 0,
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: false,
      isInitialLoading: true,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: true,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    expect(screen.getByText('Votes…')).toBeInTheDocument()
  })

  it('renders vote counts', () => {
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 3,
        downvotes: 1,
        score: 2,
        byCollaborator: {},
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('+2')).toBeInTheDocument()
  })

  it('highlights user vote', () => {
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 5,
        downvotes: 2,
        score: 3,
        byCollaborator: { 'collab-1': 1 },
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    const upvoteBtn = screen.getByTitle('Remove upvote')
    expect(upvoteBtn).toHaveClass('active')
  })

  it('calls recordVote on upvote button click', () => {
    const mockMutate = vi.fn()
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 2,
        downvotes: 0,
        score: 2,
        byCollaborator: {},
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: mockMutate,
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    const upvoteBtn = screen.getByTitle('Upvote this option')
    fireEvent.click(upvoteBtn)

    expect(mockMutate).toHaveBeenCalledWith(
      { collaboratorId: 'collab-1', vote: 1 },
      expect.any(Object)
    )
  })

  it('toggles vote off when clicking the same button', () => {
    const mockMutate = vi.fn()
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 3,
        downvotes: 1,
        score: 2,
        byCollaborator: { 'collab-1': 1 },
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: mockMutate,
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    const upvoteBtn = screen.getByTitle('Remove upvote')
    fireEvent.click(upvoteBtn)

    expect(mockMutate).toHaveBeenCalledWith(
      { collaboratorId: 'collab-1', vote: 0 },
      expect.any(Object)
    )
  })

  it('switches vote from upvote to downvote', () => {
    const mockMutate = vi.fn()
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 4,
        downvotes: 1,
        score: 3,
        byCollaborator: { 'collab-1': 1 },
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: mockMutate,
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    const downvoteBtn = screen.getByTitle('Downvote this option')
    fireEvent.click(downvoteBtn)

    expect(mockMutate).toHaveBeenCalledWith(
      { collaboratorId: 'collab-1', vote: -1 },
      expect.any(Object)
    )
  })

  it('does not show score when it is zero', () => {
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 2,
        downvotes: 2,
        score: 0,
        byCollaborator: {},
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent()
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('passes correct vote value when voting', () => {
    const mockMutate = vi.fn()
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 1,
        downvotes: 0,
        score: 1,
        byCollaborator: {},
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: mockMutate,
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent('opt-1', 'collab-1')
    const upvoteBtn = screen.getByTitle('Upvote this option')
    fireEvent.click(upvoteBtn)

    expect(mockMutate).toHaveBeenCalledWith(
      { collaboratorId: 'collab-1', vote: 1 },
      expect.any(Object)
    )
  })

  it('returns null when no collaborator is logged in', () => {
    mockUseVotes.mockReturnValue({
      data: {
        optionId: 'opt-1',
        upvotes: 2,
        downvotes: 0,
        score: 2,
        byCollaborator: {},
      },
      isLoading: false,
      isPending: false,
      status: 'success',
      isError: false,
      error: null,
      dataUpdatedAt: Date.now(),
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: null,
      isFetchedAfterMount: true,
      isInitialLoading: false,
      isPaused: false,
      isFetching: false,
      isRefetching: false,
      isStale: false,
      refetch: vi.fn(),
    })
    mockUseRecordVote.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isPaused: false,
      isError: false,
      isIdle: true,
      isSuccess: false,
      data: undefined,
      error: null,
      failureCount: 0,
      failureReason: null,
      status: 'idle',
      reset: vi.fn(),
    })

    renderComponent('opt-1', null)
    expect(screen.queryByText('votes')).not.toBeInTheDocument()
  })
})
