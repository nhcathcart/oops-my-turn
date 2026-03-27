import { createContext, useContext } from 'react'
import { useGetMe } from '@/api/auth'
import type { GetMeResponse } from '@/api/generated/types.gen'

type AuthContextValue = {
  user: GetMeResponse | null
  isLoading: boolean
}

const AuthContext = createContext<AuthContextValue>({ user: null, isLoading: true })

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data, isLoading } = useGetMe()

  return (
    <AuthContext.Provider value={{ user: data ?? null, isLoading }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
