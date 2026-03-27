import { useGetHello } from '@/api/hello'
import { PageContainer } from '@/components/page-container'
import { PageHeader } from '@/components/page-header'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/providers/auth-provider'

export default function HomePage() {
  const { user } = useAuth()
  const { data: hello } = useGetHello()

  return (
    <PageContainer>
      <PageHeader
        title="oops-my-turn dashboard"
        subtitle="This app keeps Google OAuth, the generated API client, PostgreSQL, and Terraform in place."
      />

      <Card>
        <CardHeader>
          <CardTitle>Signed In</CardTitle>
          <CardDescription>Your authenticated oops-my-turn route is working.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <p>
            <span className="font-medium">User:</span> {user?.first_name} {user?.last_name}
          </p>
          <p>
            <span className="font-medium">Email:</span> {user?.email}
          </p>
          <p>
            <span className="font-medium">API:</span> {hello?.message ?? 'Loading...'}
          </p>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
