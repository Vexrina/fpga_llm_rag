import type { RagSettings, HistoryEntry, LogEntry, KnowledgeDoc, DocumentVersion } from '../types'
import { getRagSettingsAPI, updateRagSettingAPI, getRagSettingsHistoryAPI, getDocumentHistoryAPI, rollbackDocumentAPI, getAllDocumentsAPI, getDocumentByIdAPI } from '../api/graphql'

const defaultSettings: RagSettings = {
  topK: 5,
  similarityThreshold: 0.75,
  model: 'mxbai-embed-large',
  chunkSize: 512,
  chunkOverlap: 64,
  basePrompt: 'Вы — помощник для ответов на вопросы студентов университета. Отвечайте кратко и по делу.',
  comparisonMethod: 'cosine',
}

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
  try {
    const settings = await getRagSettingsAPI()
    return {
      topK: parseInt(settings.topK || '5', 10),
      similarityThreshold: parseFloat(settings.similarityThreshold || '0.75'),
      model: settings.model || 'mxbai-embed-large',
      chunkSize: parseInt(settings.chunkSize || '512', 10),
      chunkOverlap: parseInt(settings.chunkOverlap || '64', 10),
      basePrompt: settings.basePrompt || defaultSettings.basePrompt,
      comparisonMethod: (settings.comparisonMethod as RagSettings['comparisonMethod']) || 'cosine',
    }
  } catch {
    return { ...defaultSettings }
  }
}

export async function saveRagSettings(settings: RagSettings): Promise<RagSettings> {
  const fields: (keyof RagSettings)[] = ['topK', 'similarityThreshold', 'chunkSize', 'chunkOverlap', 'basePrompt', 'comparisonMethod', 'model']
  for (const key of fields) {
    const value = key === 'basePrompt' ? settings[key] : String(settings[key])
    await updateRagSettingAPI(key, value, 'admin')
  }
  return { ...settings }
}

export async function testRagConnection(): Promise<{ ok: boolean; message: string }> {
  await new Promise((resolve) => setTimeout(resolve, 1000))
  return { ok: true, message: 'Соединение установлено успешно' }
}

export async function getHistory(): Promise<HistoryEntry[]> {
  try {
    const history = await getRagSettingsHistoryAPI(20)
    return history
      .filter((h) => h.oldValue !== h.newValue)
      .map((h) => ({
        id: String(h.id),
        timestamp: h.changedAt ? new Date(h.changedAt) : new Date(),
        user: h.changedBy,
        field: h.settingKey,
        oldValue: h.oldValue,
        newValue: h.newValue,
      }))
  } catch {
    return []
  }
}

export async function getLogs(): Promise<LogEntry[]> {
  await new Promise((resolve) => setTimeout(resolve, 300))
  return [...mockLogs]
}

export async function getDocuments(): Promise<KnowledgeDoc[]> {
  try {
    const docs = await getAllDocumentsAPI()
    return docs.map(d => ({
      id: d.id,
      title: d.title,
      updatedAt: new Date(d.updatedAt),
      indexed: d.indexed,
      size: d.size,
      chunks: d.chunks,
      content: '',
    }))
  } catch (e) {
    console.error('getDocuments API failed:', e)
    return [...mockDocs]
  }
}

export async function getDocument(id: string): Promise<KnowledgeDoc | null> {
  try {
    const doc = await getDocumentByIdAPI(id)
    if (!doc) return null
    return {
      id: doc.id,
      title: doc.title,
      content: doc.content,
      updatedAt: new Date(),
      indexed: true,
      size: doc.content.length * 2,
      chunks: 1,
    }
  } catch (e) {
    console.error('getDocument API failed:', e)
    return mockDocs.find((d) => d.id === id) ?? null
  }
}

const mockDocumentHistory: Record<string, DocumentVersion[]> = {
  'doc-1': [
    {
      id: 3,
      documentId: 'doc-1',
      title: 'Справочник студента 2026.pdf',
      content: 'Справочник студента содержит основную информацию о правилах обучения, оформления документов, расписании, библиотеке и других аспектах университетской жизни.\n\nРаздел 1: Общие положения\nСтудент университета имеет право на получение образования в соответствии с федеральными государственными образовательными стандартами.\n\nРаздел 2: Оформление документов\nДля получения справки об обучении необходимо обратиться в деканат с письменным заявлением.\n\nРаздел 3: Библиотека\nБиблиотека университета работает с понедельника по пятницу с 9:00 до 20:00, в субботу с 10:00 до 17:00.',
      versionNumber: 3,
      createdAt: new Date(Date.now() - 3600000 * 24).toISOString(),
      createdBy: 'admin',
      action: 'update',
    },
    {
      id: 2,
      documentId: 'doc-1',
      title: 'Справочник студента 2026.pdf',
      content: 'Справочник студента содержит основную информацию о правилах обучения, оформления документов и расписании.\n\nРаздел 1: Общие положения\nСтудент имеет право на образование.\n\nРаздел 2: Оформление документов\nДля получения справки обратитесь в деканат.',
      versionNumber: 2,
      createdAt: new Date(Date.now() - 3600000 * 72).toISOString(),
      createdBy: 'admin',
      action: 'update',
    },
    {
      id: 1,
      documentId: 'doc-1',
      title: 'Справочник студента 2026.pdf',
      content: 'Справочник студента - начальная версия',
      versionNumber: 1,
      createdAt: new Date(Date.now() - 3600000 * 168).toISOString(),
      createdBy: 'admin',
      action: 'create',
    },
  ],
}

export async function getDocumentHistory(id: string): Promise<DocumentVersion[]> {
  await new Promise((resolve) => setTimeout(resolve, 300))
  try {
    return await getDocumentHistoryAPI(id)
  } catch {
    return mockDocumentHistory[id] || []
  }
}

export async function rollbackDocument(documentId: string, versionId: number): Promise<{ success: boolean; message: string }> {
  await new Promise((resolve) => setTimeout(resolve, 500))
  try {
    const result = await rollbackDocumentAPI(documentId, versionId)
    return { success: result.success, message: result.message }
  } catch (e) {
    return { success: false, message: String(e) }
  }
}
