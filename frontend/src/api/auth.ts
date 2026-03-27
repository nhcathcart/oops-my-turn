// Query hooks for the auth API resource.
import { useCallback } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getMeOptions, logoutMutation } from '@/api/generated/@tanstack/react-query.gen'

export function useGetMe() {
  return useQuery({
    ...getMeOptions(),
    retry: false,
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  const mutation = useMutation(logoutMutation())

  const logout = useCallback(async () => {
    await mutation.mutateAsync({})
    queryClient.resetQueries()
  }, [mutation, queryClient])

  return { logout, ...mutation }
}
