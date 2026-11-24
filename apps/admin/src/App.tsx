import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { Toaster } from 'sonner'

import { queryClient } from '@/lib/query-client'
import { router } from '@/routes/router'
import { useThemeStore } from '@/stores/theme'

function App() {
  const theme = useThemeStore((state) => state.theme)

  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster
        theme={theme}
        richColors
        closeButton
        position="top-right"
        toastOptions={{ className: 'font-sans' }}
      />
    </QueryClientProvider>
  )
}

export default App
