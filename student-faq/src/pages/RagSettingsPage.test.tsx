import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import RagSettingsPage from '../pages/RagSettingsPage'
import * as ragMocks from '../mocks/rag'

vi.mock('../mocks/rag', () => ({
  getRagSettings: vi.fn(),
  saveRagSettings: vi.fn(),
  testRagConnection: vi.fn(),
  getHistory: vi.fn(),
  getLogs: vi.fn(),
  getDocuments: vi.fn(),
  getDocument: vi.fn(),
}))

const renderWithRouter = (component: React.ReactNode) => {
  return render(<MemoryRouter>{component}</MemoryRouter>)
}

const defaultSettings = {
  topK: 5,
  similarityThreshold: 0.75,
  model: 'mxbai-embed-large',
  chunkSize: 512,
  chunkOverlap: 64,
  basePrompt: 'test prompt',
  comparisonMethod: 'cosine',
}

describe('RagSettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(ragMocks.getRagSettings).mockResolvedValue({ ...defaultSettings })
    vi.mocked(ragMocks.getHistory).mockResolvedValue([])
    vi.mocked(ragMocks.getLogs).mockResolvedValue([])
    vi.mocked(ragMocks.getDocuments).mockResolvedValue([])
  })

  it('renders all tabs', async () => {
    renderWithRouter(<RagSettingsPage />)

    await waitFor(() => {
      expect(screen.getByText('Настройки RAG')).toBeInTheDocument()
    })
    expect(screen.getByText('История изменений')).toBeInTheDocument()
    expect(screen.getByText('Логи запросов')).toBeInTheDocument()
    expect(screen.getByText('База знаний')).toBeInTheDocument()
  })

  it('renders all settings fields with tooltips', async () => {
    renderWithRouter(<RagSettingsPage />)

    await waitFor(() => {
      expect(screen.getByLabelText(/top k/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/порог схожести/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/размер чанка/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/перекрытие чанков/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/базовый промпт llm/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/метод сравнения/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/модель эмбеддингов/i)).toBeInTheDocument()
    })

    expect(screen.getAllByText('?').length).toBeGreaterThanOrEqual(7)
  })

  it('loads and displays current settings', async () => {
    renderWithRouter(<RagSettingsPage />)

    await waitFor(() => {
      expect(screen.getByDisplayValue('5')).toBeInTheDocument()
      expect(screen.getByDisplayValue('0.75')).toBeInTheDocument()
      expect(screen.getByDisplayValue('512')).toBeInTheDocument()
      expect(screen.getByDisplayValue('64')).toBeInTheDocument()
    })
  })

  it('saves settings when save button is clicked', async () => {
    vi.mocked(ragMocks.saveRagSettings).mockResolvedValue({ ...defaultSettings })

    renderWithRouter(<RagSettingsPage />)

    await waitFor(() => {
      expect(screen.getByLabelText(/top k/i)).toBeInTheDocument()
    })

    const saveButton = screen.getByRole('button', { name: /сохранить/i })
    fireEvent.click(saveButton)

    await waitFor(() => {
      expect(ragMocks.saveRagSettings).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(screen.getByText('Настройки сохранены')).toBeInTheDocument()
    })
  })

  it('tests connection when test button is clicked', async () => {
    vi.mocked(ragMocks.testRagConnection).mockResolvedValue({
      ok: true,
      message: 'Соединение установлено успешно',
    })

    renderWithRouter(<RagSettingsPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /тест соединения/i })).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /тест соединения/i }))

    await waitFor(() => {
      expect(ragMocks.testRagConnection).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(screen.getByText('Соединение установлено успешно')).toBeInTheDocument()
    })
  })

  it('shows history tab content', async () => {
    const mockHistory = [
      {
        id: 'h-1',
        timestamp: new Date(),
        user: 'admin',
        field: 'topK',
        oldValue: '3',
        newValue: '5',
      },
    ]
    vi.mocked(ragMocks.getHistory).mockResolvedValue(mockHistory)

    renderWithRouter(<RagSettingsPage />)

    fireEvent.click(screen.getByText('История изменений'))

    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument()
      expect(screen.getByText('topK')).toBeInTheDocument()
    })
  })

  it('shows logs tab content', async () => {
    const mockLogs = [
      {
        id: 'log-1',
        timestamp: new Date(),
        userQuery: 'Test query',
        topK: 5,
        model: 'mxbai-embed-large',
        responseTimeMs: 300,
        found: true,
      },
    ]
    vi.mocked(ragMocks.getLogs).mockResolvedValue(mockLogs)

    renderWithRouter(<RagSettingsPage />)

    fireEvent.click(screen.getByText('Логи запросов'))

    await waitFor(() => {
      expect(screen.getByText('Test query')).toBeInTheDocument()
      expect(screen.getByText('Найдено')).toBeInTheDocument()
    })
  })

  it('shows knowledge base tab with documents', async () => {
    const mockDocs = [
      {
        id: 'doc-1',
        title: 'test.pdf',
        updatedAt: new Date(),
        indexed: true,
        size: 1024,
        chunks: 10,
        content: 'Document content here',
      },
    ]
    vi.mocked(ragMocks.getDocuments).mockResolvedValue(mockDocs)

    renderWithRouter(<RagSettingsPage />)

    fireEvent.click(screen.getByText('База знаний'))

    await waitFor(() => {
      expect(screen.getByText('test.pdf')).toBeInTheDocument()
      expect(screen.getByText('Проиндексирован')).toBeInTheDocument()
    })
  })
})
