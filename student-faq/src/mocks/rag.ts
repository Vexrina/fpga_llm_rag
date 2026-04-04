import type { RagSettings, HistoryEntry, LogEntry, KnowledgeDoc } from '../types'

const defaultSettings: RagSettings = {
  topK: 5,
  similarityThreshold: 0.75,
  model: 'mxbai-embed-large',
  chunkSize: 512,
  chunkOverlap: 64,
  basePrompt: 'Вы — помощник для ответов на вопросы студентов университета. Отвечайте кратко и по делу.',
  comparisonMethod: 'cosine',
}

const mockHistory: HistoryEntry[] = [
  {
    id: 'h-1',
    timestamp: new Date(Date.now() - 3600000 * 2),
    user: 'admin',
    field: 'topK',
    oldValue: '3',
    newValue: '5',
  },
  {
    id: 'h-2',
    timestamp: new Date(Date.now() - 3600000 * 5),
    user: 'admin',
    field: 'similarityThreshold',
    oldValue: '0.7',
    newValue: '0.75',
  },
  {
    id: 'h-3',
    timestamp: new Date(Date.now() - 3600000 * 24),
    user: 'admin',
    field: 'model',
    oldValue: 'all-MiniLM-L6-v2',
    newValue: 'mxbai-embed-large',
  },
  {
    id: 'h-4',
    timestamp: new Date(Date.now() - 3600000 * 48),
    user: 'admin',
    field: 'chunkSize',
    oldValue: '256',
    newValue: '512',
  },
]

const mockLogs: LogEntry[] = [
  {
    id: 'log-1',
    timestamp: new Date(Date.now() - 60000 * 5),
    userQuery: 'Как получить справку об обучении?',
    topK: 5,
    model: 'mxbai-embed-large',
    responseTimeMs: 342,
    found: true,
  },
  {
    id: 'log-2',
    timestamp: new Date(Date.now() - 60000 * 30),
    userQuery: 'Расписание экзаменов',
    topK: 5,
    model: 'mxbai-embed-large',
    responseTimeMs: 289,
    found: true,
  },
  {
    id: 'log-3',
    timestamp: new Date(Date.now() - 60000 * 120),
    userQuery: 'Где находится библиотека?',
    topK: 5,
    model: 'mxbai-embed-large',
    responseTimeMs: 198,
    found: true,
  },
  {
    id: 'log-4',
    timestamp: new Date(Date.now() - 60000 * 180),
    userQuery: 'Какой Wi-Fi пароль в общежитии?',
    topK: 5,
    model: 'mxbai-embed-large',
    responseTimeMs: 415,
    found: false,
  },
  {
    id: 'log-5',
    timestamp: new Date(Date.now() - 60000 * 300),
    userQuery: 'Оформление академического отпуска',
    topK: 5,
    model: 'mxbai-embed-large',
    responseTimeMs: 267,
    found: true,
  },
]

const mockDocs: KnowledgeDoc[] = [
  {
    id: 'doc-1',
    title: 'Справочник студента 2026.pdf',
    updatedAt: new Date(Date.now() - 3600000 * 24),
    indexed: true,
    size: 2457600,
    chunks: 48,
    content: 'Справочник студента содержит основную информацию о правилах обучения, оформления документов, расписании, библиотеке и других аспектах университетской жизни.\n\nРаздел 1: Общие положения\nСтудент университета имеет право на получение образования в соответствии с федеральными государственными образовательными стандартами.\n\nРаздел 2: Оформление документов\nДля получения справки об обучении необходимо обратиться в деканат с письменным заявлением.\n\nРаздел 3: Библиотека\nБиблиотека университета работает с понедельника по пятницу с 9:00 до 20:00, в субботу с 10:00 до 17:00.',
  },
  {
    id: 'doc-2',
    title: 'Расписание занятий весна 2026.xlsx',
    updatedAt: new Date(Date.now() - 3600000 * 12),
    indexed: true,
    size: 512000,
    chunks: 12,
    content: 'Расписание занятий на весенний семестр 2026 года.\n\nПонедельник:\n09:00 - Математический анализ (ауд. 301)\n11:00 - Физика (ауд. 205)\n14:00 - Программирование (ауд. 410)\n\nВторник:\n09:00 - Английский язык (ауд. 115)\n11:00 - Дискретная математика (ауд. 302)\n\nСреда:\n10:00 - Физика (лаб., ауд. 108)\n13:00 - Математический анализ (ауд. 301)',
  },
  {
    id: 'doc-3',
    title: 'Правила общежития.docx',
    updatedAt: new Date(Date.now() - 3600000 * 72),
    indexed: true,
    size: 102400,
    chunks: 8,
    content: 'Правила проживания в студенческом общежитии.\n\n1. Проживающие обязаны соблюдать тишину с 23:00 до 7:00.\n2. Запрещается использование электроприборов мощностью свыше 1000 Вт.\n3. Wi-Fi сеть: University-Dorm, пароль: student2026.\n4. Посещение гостей разрешено до 22:00.\n5. Уборка общих помещений производится дежурными по графику.',
  },
  {
    id: 'doc-4',
    title: 'FAQ по стипендиям.pdf',
    updatedAt: new Date(Date.now() - 3600000 * 168),
    indexed: false,
    size: 819200,
    chunks: 0,
    content: 'Часто задаваемые вопросы о стипендиях.\n\nВопрос: Как получить повышенную стипендию?\nОтвет: Для получения повышенной стипендии необходимо иметь средний балл не ниже 4.5 и активное участие в научной деятельности.\n\nВопрос: Когда выплачивается стипендия?\nОтвет: Стипендия выплачивается ежемесячно, не позднее 20 числа текущего месяца.',
  },
]

export async function getRagSettings(): Promise<RagSettings> {
  await new Promise((resolve) => setTimeout(resolve, 300))
  return { ...defaultSettings }
}

export async function saveRagSettings(settings: RagSettings): Promise<RagSettings> {
  await new Promise((resolve) => setTimeout(resolve, 500))
  return { ...settings }
}

export async function testRagConnection(): Promise<{ ok: boolean; message: string }> {
  await new Promise((resolve) => setTimeout(resolve, 1000))
  return { ok: true, message: 'Соединение установлено успешно' }
}

export async function getHistory(): Promise<HistoryEntry[]> {
  await new Promise((resolve) => setTimeout(resolve, 300))
  return [...mockHistory]
}

export async function getLogs(): Promise<LogEntry[]> {
  await new Promise((resolve) => setTimeout(resolve, 300))
  return [...mockLogs]
}

export async function getDocuments(): Promise<KnowledgeDoc[]> {
  await new Promise((resolve) => setTimeout(resolve, 300))
  return [...mockDocs]
}

export async function getDocument(id: string): Promise<KnowledgeDoc | null> {
  await new Promise((resolve) => setTimeout(resolve, 200))
  return mockDocs.find((d) => d.id === id) ?? null
}
