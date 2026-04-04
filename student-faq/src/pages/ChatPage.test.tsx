import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import ChatPage from '../pages/ChatPage'
import * as chatMocks from '../mocks/chat'

vi.mock('../mocks/chat', () => ({
  getChatHistory: vi.fn(),
  sendMessage: vi.fn(),
}))

const renderWithRouter = (component: React.ReactNode) => {
  return render(<MemoryRouter>{component}</MemoryRouter>)
}

describe('ChatPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders chat input and send button', async () => {
    vi.mocked(chatMocks.getChatHistory).mockResolvedValue([])

    renderWithRouter(<ChatPage />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Введите вопрос...')).toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /отправить/i })).toBeInTheDocument()
  })

  it('loads and displays chat history', async () => {
    const history = [
      {
        id: 'msg-1',
        role: 'user' as const,
        content: 'Привет',
        timestamp: new Date(),
      },
      {
        id: 'msg-2',
        role: 'assistant' as const,
        content: 'Здравствуйте!',
        timestamp: new Date(),
      },
    ]

    vi.mocked(chatMocks.getChatHistory).mockResolvedValue(history)

    renderWithRouter(<ChatPage />)

    await waitFor(() => {
      expect(screen.getByText('Привет')).toBeInTheDocument()
      expect(screen.getByText('Здравствуйте!')).toBeInTheDocument()
    })
  })

  it('sends a message and receives a response', async () => {
    const user = userEvent.setup()

    vi.mocked(chatMocks.getChatHistory).mockResolvedValue([])
    vi.mocked(chatMocks.sendMessage).mockResolvedValue({
      id: 'msg-resp',
      role: 'assistant',
      content: 'Mock response',
      timestamp: new Date(),
    })

    renderWithRouter(<ChatPage />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Введите вопрос...')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText('Введите вопрос...')
    await user.type(input, 'Test question')

    const sendButton = screen.getByRole('button', { name: /отправить/i })
    await user.click(sendButton)

    await waitFor(() => {
      expect(screen.getByText('Test question')).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText('Mock response')).toBeInTheDocument()
    })

    expect(chatMocks.sendMessage).toHaveBeenCalledWith('Test question')
  })

  it('shows error message when sendMessage fails', async () => {
    const user = userEvent.setup()

    vi.mocked(chatMocks.getChatHistory).mockResolvedValue([])
    vi.mocked(chatMocks.sendMessage).mockRejectedValue(new Error('Network error'))

    renderWithRouter(<ChatPage />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Введите вопрос...')).toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText('Введите вопрос...')
    await user.type(input, 'Test question')

    const sendButton = screen.getByRole('button', { name: /отправить/i })
    await user.click(sendButton)

    await waitFor(() => {
      expect(screen.getByText('Произошла ошибка. Попробуйте снова.')).toBeInTheDocument()
    })
  })

  it('disables send button when input is empty', async () => {
    vi.mocked(chatMocks.getChatHistory).mockResolvedValue([])

    renderWithRouter(<ChatPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /отправить/i })).toBeInTheDocument()
    })

    const sendButton = screen.getByRole('button', { name: /отправить/i })
    expect(sendButton).toBeDisabled()
  })
})
