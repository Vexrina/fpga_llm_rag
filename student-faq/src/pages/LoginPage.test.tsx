import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from '../pages/LoginPage'
import * as authMocks from '../mocks/auth'

vi.mock('../mocks/auth', () => ({
  login: vi.fn().mockResolvedValue({ token: 'mock-token', username: 'admin' }),
}))

const renderWithRouter = (component: React.ReactNode, onLogin = vi.fn()) => {
  return render(<MemoryRouter><LoginPage onLogin={onLogin} /></MemoryRouter>)
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.mocked(authMocks.login).mockClear()
  })

  it('renders login form', () => {
    renderWithRouter(<LoginPage onLogin={vi.fn()} />)

    expect(screen.getByLabelText(/имя пользователя/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/пароль/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /войти/i })).toBeInTheDocument()
  })

  it('shows validation errors for empty fields', async () => {
    const user = userEvent.setup()
    renderWithRouter(<LoginPage onLogin={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /войти/i }))

    await waitFor(() => {
      expect(screen.getByText('Введите имя пользователя')).toBeInTheDocument()
      expect(screen.getByText('Введите пароль')).toBeInTheDocument()
    })
  })

  it('shows validation error for short username', async () => {
    const user = userEvent.setup()
    renderWithRouter(<LoginPage onLogin={vi.fn()} />)

    await user.type(screen.getByLabelText(/имя пользователя/i), 'ab')
    await user.click(screen.getByRole('button', { name: /войти/i }))

    await waitFor(() => {
      expect(screen.getByText('Имя должно содержать минимум 3 символа')).toBeInTheDocument()
    })
  })

  it('shows validation error for short password', async () => {
    const user = userEvent.setup()
    renderWithRouter(<LoginPage onLogin={vi.fn()} />)

    await user.type(screen.getByLabelText(/имя пользователя/i), 'admin')
    await user.type(screen.getByLabelText(/пароль/i), '123')
    await user.click(screen.getByRole('button', { name: /войти/i }))

    await waitFor(() => {
      expect(screen.getByText('Пароль должен содержать минимум 6 символов')).toBeInTheDocument()
    })
  })

  it('calls login with correct credentials when form is submitted', async () => {
    const onLogin = vi.fn()

    vi.mocked(authMocks.login).mockResolvedValue({ token: 'mock-token', username: 'admin' })
    const { container } = renderWithRouter(<LoginPage onLogin={onLogin} />)

    const usernameInput = screen.getByLabelText(/имя пользователя/i)
    const passwordInput = screen.getByLabelText(/пароль/i)

    fireEvent.change(usernameInput, { target: { value: 'admin' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })

    const form = container.querySelector('form')!
    await act(async () => {
      fireEvent.submit(form)
    })

    await waitFor(() => {
      expect(authMocks.login).toHaveBeenCalledWith({
        username: 'admin',
        password: 'password123',
      })
    })

    await waitFor(() => {
      expect(onLogin).toHaveBeenCalled()
    })
  })

  it('shows error when login fails', async () => {
    vi.mocked(authMocks.login).mockRejectedValue(new Error('Неверные учётные данные'))
    const { container } = renderWithRouter(<LoginPage onLogin={vi.fn()} />)

    fireEvent.change(screen.getByLabelText(/имя пользователя/i), { target: { value: 'admin' } })
    fireEvent.change(screen.getByLabelText(/пароль/i), { target: { value: 'password123' } })

    const form = container.querySelector('form')!
    fireEvent.submit(form)

    await waitFor(() => {
      expect(screen.getByText('Неверные учётные данные')).toBeInTheDocument()
    })
  })

  it('clears validation errors on input change', async () => {
    const { container } = renderWithRouter(<LoginPage onLogin={vi.fn()} />)

    const form = container.querySelector('form')!
    fireEvent.submit(form)

    await waitFor(() => {
      expect(screen.getByText('Введите имя пользователя')).toBeInTheDocument()
    })

    fireEvent.change(screen.getByLabelText(/имя пользователя/i), { target: { value: 'a' } })

    expect(screen.queryByText('Введите имя пользователя')).not.toBeInTheDocument()
  })
})
