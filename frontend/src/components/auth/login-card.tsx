import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card'

interface LoginCardProps {
  onSignIn: () => void
}

export function LoginCard({ onSignIn }: LoginCardProps) {
  return (
    <Card className="w-full max-w-sm">
      <CardHeader className="text-center">
        <h1 className="text-2xl font-semibold">Starter App</h1>
        <CardDescription>Sign in with Google to access the starter dashboard</CardDescription>
      </CardHeader>
      <CardContent>
        <Button className="w-full" onClick={onSignIn}>
          Sign in with Google
        </Button>
      </CardContent>
    </Card>
  )
}
