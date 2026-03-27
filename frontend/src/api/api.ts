import { QueryClient } from '@tanstack/react-query'
import { client } from '@/api/generated/client.gen'

client.setConfig({
  baseUrl: import.meta.env.VITE_API_URL ?? '',
  credentials: 'include',
})

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 minute
    },
  },
})
