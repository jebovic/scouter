import { type ReactNode } from 'react'
import { render, type RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '../components/scouter'

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
      mutations: { retry: false },
    },
  })
}

interface ProvidersProps {
  children: ReactNode
  initialEntries?: string[]
}

function AllProviders({ children, initialEntries = ['/'] }: ProvidersProps) {
  const queryClient = createTestQueryClient()
  return (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>
        <ToastProvider>
          {children}
        </ToastProvider>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

function renderWithProviders(
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'> & { initialEntries?: string[] },
) {
  const { initialEntries, ...rest } = options ?? {}
  return render(ui, {
    wrapper: ({ children }) => (
      <AllProviders initialEntries={initialEntries}>{children}</AllProviders>
    ),
    ...rest,
  })
}

export { renderWithProviders, createTestQueryClient }
export * from '@testing-library/react'
