import { render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { vi } from 'vitest'
import App from './app'
import { useLogout } from '@/api/auth'
import { useGetHello } from '@/api/hello'
import { useAuth } from '@/providers/auth-provider'
import { ThemeProvider } from '@/providers/theme-provider'

vi.mock('@/api/auth')
vi.mock('@/api/hello')
vi.mock('@/providers/auth-provider')

function renderApp(initialPath = '/') {
  return render(
    <ThemeProvider>
      <MemoryRouter initialEntries={[initialPath]}>
        <App />
      </MemoryRouter>
    </ThemeProvider>,
  )
}

beforeEach(() => {
  vi.mocked(useAuth).mockReturnValue({
    user: { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User' },
    isLoading: false,
  })
  vi.mocked(useLogout).mockReturnValue({ logout: vi.fn() } as any)
  vi.mocked(useGetHello).mockReturnValue({
    data: { message: 'Hello from oops-my-turn.' },
    isLoading: false,
  } as any)
})

test('renders the oops-my-turn dashboard on the root route', async () => {
  renderApp('/')

  expect(await screen.findByRole('heading', { name: 'oops-my-turn dashboard' })).toBeInTheDocument()
  expect(screen.getByText('Hello from oops-my-turn.')).toBeInTheDocument()
})

test('shows the home breadcrumb on the dashboard', async () => {
  renderApp('/')

  const breadcrumb = await screen.findByLabelText('breadcrumb')
  expect(within(breadcrumb).getByText('Home')).toBeInTheDocument()
})

test('redirects unauthenticated users to login', async () => {
  vi.mocked(useAuth).mockReturnValue({
    user: null,
    isLoading: false,
  })

  renderApp('/')

  expect(await screen.findByText('Sign in with Google')).toBeInTheDocument()
})
