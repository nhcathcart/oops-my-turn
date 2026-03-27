import { render, screen } from '@testing-library/react'
import { ReactQueryProvider } from './react-query-provider'

test('renders children', () => {
  render(
    <ReactQueryProvider>
      <span>hello from child</span>
    </ReactQueryProvider>,
  )

  expect(screen.getByText('hello from child')).toBeInTheDocument()
})
