import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider, notifyManager } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { router } from './router'
import { ToastProvider } from './components/scouter'
import './i18n'
import './styles/global.css'

// Defer TanStack Query batch notifications to the microtask queue.
// Without this, React 19's concurrent renderer throws error #310 when
// multiple queries error simultaneously (e.g. 3x 404 at page mount).
notifyManager.setScheduler(queueMicrotask)

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>
  </StrictMode>,
)
