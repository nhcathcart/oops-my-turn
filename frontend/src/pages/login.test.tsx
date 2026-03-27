import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { vi } from 'vitest'
import LoginPage from './login'
import { useAuth } from '@/providers/auth-provider'

vi.mock('@/providers/auth-provider', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/providers/auth-provider')>()
  return {
    ...actual,
    useAuth: vi.fn(),
  }
})

function renderPage(initialPath = '/login') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<div>Home</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

beforeEach(() => {
  vi.mocked(useAuth).mockReturnValue({ user: null, isLoading: false } as any)
})

test('renders the login card for signed out users', () => {
  renderPage()

  expect(screen.getByRole('heading', { name: 'Starter App' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Sign in with Google' })).toBeInTheDocument()
})

test('redirects authenticated users to the home route', () => {
  vi.mocked(useAuth).mockReturnValue({
    user: { id: 'usr_123', email: 'test@example.com', first_name: 'Test', last_name: 'User' },
    isLoading: false,
  } as any)

  renderPage()

  expect(screen.getByText('Home')).toBeInTheDocument()
})

test('renders nothing while auth state is loading', () => {
  vi.mocked(useAuth).mockReturnValue({ user: null, isLoading: true } as any)
  const { container } = renderPage()

  expect(container).toBeEmptyDOMElement()
})
