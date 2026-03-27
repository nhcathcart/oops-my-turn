import { Navigate } from 'react-router'
import { LoginCard } from '@/components/auth/login-card'
import { useAuth } from '@/providers/auth-provider'

export default function LoginPage() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return null
  }

  if (user) {
    return <Navigate to="/" replace />
  }

  return (
    <div className="flex min-h-svh items-center justify-center">
      <LoginCard
        onSignIn={() => {
          window.location.href = `${import.meta.env.VITE_API_URL ?? ''}/api/v1/auth/google/login`
        }}
      />
    </div>
  )
}
