import { useState, useEffect, type ChangeEvent } from 'react'
import type { RagSettings, HistoryEntry, LogEntry, KnowledgeDoc } from '../types'
import {
  getRagSettings,
  saveRagSettings,
  testRagConnection,
  getHistory,
  getLogs,
  getDocuments,
  getDocument,
} from '../mocks/rag'
import Tooltip from '../components/Tooltip'

type Tab = 'settings' | 'history' | 'logs' | 'knowledge'

const TABS: { key: Tab; label: string }[] = [
  { key: 'settings', label: 'Настройки RAG' },
  { key: 'history', label: 'История изменений' },
  { key: 'logs', label: 'Логи запросов' },
  { key: 'knowledge', label: 'База знаний' },
]

const TOOLTIPS = {
  topK: 'Количество наиболее релевантных документов, которые извлекаются из базы знаний для формирования ответа',
  similarityThreshold: 'Минимальный порог cosine similarity. Документы с оценкой ниже этого значения отбрасываются',
  chunkSize: 'Размер текстового фрагмента (в токенах), на который разбивается документ перед эмбеддингом',
  chunkOverlap: 'Количество токенов перекрытия между соседними фрагментами для сохранения контекста',
  basePrompt: 'Системный промпт, который передаётся LLM перед каждым запросом. Определяет роль и стиль ответов',
  comparisonMethod: 'Метод вычисления схожести векторов: cosine — косинусное расстояние, dot — скалярное произведение, euclidean — евклидово расстояние',
  model: 'Модель для генерации векторных представлений (эмбеддингов) текста',
}

export default function RagSettingsPage() {
  const [activeTab, setActiveTab] = useState<Tab>('settings')

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Панель администратора</h1>

      <div className="flex gap-1 mb-6 bg-gray-100 rounded-lg p-1">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex-1 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.key
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'settings' && <SettingsTab />}
      {activeTab === 'history' && <HistoryTab />}
      {activeTab === 'logs' && <LogsTab />}
      {activeTab === 'knowledge' && <KnowledgeTab />}
    </div>
  )
}

function SettingsTab() {
  const [settings, setSettings] = useState<RagSettings | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [isTesting, setIsTesting] = useState(false)
  const [isInitialLoading, setIsInitialLoading] = useState(true)
  const [testResult, setTestResult] = useState<{ ok: boolean; message: string } | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    getRagSettings()
      .then((data) => setSettings(data))
      .finally(() => setIsInitialLoading(false))
  }, [])

  if (isInitialLoading) {
    return (
      <div className="text-center py-16">
        <div className="animate-pulse text-gray-400">Загрузка настроек...</div>
      </div>
    )
  }

  if (!settings) return null

  const handleChange = (field: keyof RagSettings, value: number | string) => {
    setSettings((prev) => (prev ? { ...prev, [field]: value } : prev))
    setSaveSuccess(false)
  }

  const handleSave = async () => {
    if (!settings) return
    setIsSaving(true)
    try {
      await saveRagSettings(settings)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch {
      setSaveSuccess(false)
    } finally {
      setIsSaving(false)
    }
  }

  const handleTest = async () => {
    setIsTesting(true)
    setTestResult(null)
    try {
      const result = await testRagConnection()
      setTestResult(result)
    } catch {
      setTestResult({ ok: false, message: 'Не удалось подключиться' })
    } finally {
      setIsTesting(false)
    }
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 space-y-6">
      {saveSuccess && (
        <div className="p-3 bg-green-50 border border-green-200 text-green-700 rounded-lg text-sm">
          Настройки сохранены
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <Field label="Top K" tooltip={TOOLTIPS.topK} id="topK">
          <input
            id="topK"
            type="number"
            min={1}
            max={20}
            value={settings.topK}
            onChange={(e: ChangeEvent<HTMLInputElement>) => handleChange('topK', parseInt(e.target.value) || 1)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </Field>

        <Field label="Порог схожести" tooltip={TOOLTIPS.similarityThreshold} id="similarityThreshold">
          <input
            id="similarityThreshold"
            type="number"
            min={0}
            max={1}
            step={0.05}
            value={settings.similarityThreshold}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              handleChange('similarityThreshold', parseFloat(e.target.value) || 0)
            }
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </Field>

        <Field label="Размер чанка" tooltip={TOOLTIPS.chunkSize} id="chunkSize">
          <input
            id="chunkSize"
            type="number"
            min={64}
            max={4096}
            step={64}
            value={settings.chunkSize}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              handleChange('chunkSize', parseInt(e.target.value) || 256)
            }
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </Field>

        <Field label="Перекрытие чанков" tooltip={TOOLTIPS.chunkOverlap} id="chunkOverlap">
          <input
            id="chunkOverlap"
            type="number"
            min={0}
            max={512}
            step={16}
            value={settings.chunkOverlap}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              handleChange('chunkOverlap', parseInt(e.target.value) || 0)
            }
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </Field>
      </div>

      <Field label="Базовый промпт LLM" tooltip={TOOLTIPS.basePrompt} id="basePrompt">
        <textarea
          id="basePrompt"
          rows={3}
          value={settings.basePrompt}
          onChange={(e: ChangeEvent<HTMLTextAreaElement>) => handleChange('basePrompt', e.target.value)}
          className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
        />
      </Field>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <Field label="Метод сравнения" tooltip={TOOLTIPS.comparisonMethod} id="comparisonMethod">
          <select
            id="comparisonMethod"
            value={settings.comparisonMethod}
            onChange={(e: ChangeEvent<HTMLSelectElement>) => handleChange('comparisonMethod', e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          >
            <option value="cosine">Cosine similarity</option>
            <option value="dot">Dot product</option>
            <option value="euclidean">Euclidean distance</option>
          </select>
        </Field>

        <Field label="Модель эмбеддингов" tooltip={TOOLTIPS.model} id="model">
          <select
            id="model"
            value={settings.model}
            onChange={(e: ChangeEvent<HTMLSelectElement>) => handleChange('model', e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          >
            <option value="mxbai-embed-large">mxbai-embed-large</option>
            <option value="all-MiniLM-L6-v2">all-MiniLM-L6-v2</option>
            <option value="text-embedding-ada-002">text-embedding-ada-002</option>
          </select>
        </Field>
      </div>

      <div className="flex gap-3 pt-2">
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="bg-indigo-600 text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSaving ? 'Сохранение...' : 'Сохранить'}
        </button>
        <button
          onClick={handleTest}
          disabled={isTesting}
          className="bg-white border border-gray-300 text-gray-700 px-5 py-2 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isTesting ? 'Проверка...' : 'Тест соединения'}
        </button>
      </div>

      {testResult && (
        <div
          className={`p-3 rounded-lg text-sm ${
            testResult.ok
              ? 'bg-green-50 border border-green-200 text-green-700'
              : 'bg-red-50 border border-red-200 text-red-700'
          }`}
        >
          {testResult.message}
        </div>
      )}
    </div>
  )
}

function Field({ label, tooltip, children, id }: { label: string; tooltip: string; children: React.ReactNode; id: string }) {
  return (
    <div>
      <div className="flex items-center gap-1.5 mb-1">
        <label htmlFor={id} className="text-sm font-medium text-gray-700">{label}</label>
        <Tooltip content={tooltip}>
          <span className="inline-flex items-center justify-center w-4 h-4 rounded-full bg-gray-300 text-gray-600 text-xs font-bold cursor-help leading-none">
            ?
          </span>
        </Tooltip>
      </div>
      {children}
    </div>
  )
}

function HistoryTab() {
  const [history, setHistory] = useState<HistoryEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    getHistory()
      .then((data) => setHistory(data))
      .finally(() => setIsLoading(false))
  }, [])

  if (isLoading) {
    return (
      <div className="text-center py-16">
        <div className="animate-pulse text-gray-400">Загрузка истории...</div>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-gray-50 border-b border-gray-200">
          <tr>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Дата</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Пользователь</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Поле</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Было</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Стало</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {history.map((entry) => (
            <tr key={entry.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                {entry.timestamp.toLocaleString('ru-RU')}
              </td>
              <td className="px-4 py-3">{entry.user}</td>
              <td className="px-4 py-3">
                <span className="inline-block bg-indigo-50 text-indigo-700 px-2 py-0.5 rounded text-xs font-medium">
                  {entry.field}
                </span>
              </td>
              <td className="px-4 py-3 text-gray-500 line-through">{entry.oldValue}</td>
              <td className="px-4 py-3 text-green-700">{entry.newValue}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function LogsTab() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    getLogs()
      .then((data) => setLogs(data))
      .finally(() => setIsLoading(false))
  }, [])

  if (isLoading) {
    return (
      <div className="text-center py-16">
        <div className="animate-pulse text-gray-400">Загрузка логов...</div>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-gray-50 border-b border-gray-200">
          <tr>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Время</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Запрос</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Top K</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Модель</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Время ответа</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Статус</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {logs.map((log) => (
            <tr key={log.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                {log.timestamp.toLocaleString('ru-RU')}
              </td>
              <td className="px-4 py-3 max-w-[200px] truncate" title={log.userQuery}>
                {log.userQuery}
              </td>
              <td className="px-4 py-3">{log.topK}</td>
              <td className="px-4 py-3 text-gray-600">{log.model}</td>
              <td className="px-4 py-3">{log.responseTimeMs} мс</td>
              <td className="px-4 py-3">
                <span
                  className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                    log.found
                      ? 'bg-green-50 text-green-700'
                      : 'bg-red-50 text-red-700'
                  }`}
                >
                  {log.found ? 'Найдено' : 'Не найдено'}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function KnowledgeTab() {
  const [docs, setDocs] = useState<KnowledgeDoc[]>([])
  const [selectedDoc, setSelectedDoc] = useState<KnowledgeDoc | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isDocLoading, setIsDocLoading] = useState(false)

  useEffect(() => {
    getDocuments()
      .then((data) => setDocs(data))
      .finally(() => setIsLoading(false))
  }, [])

  const handleSelectDoc = async (doc: KnowledgeDoc) => {
    setSelectedDoc(doc)
    setIsDocLoading(true)
    try {
      const full = await getDocument(doc.id)
      if (full) setSelectedDoc(full)
    } finally {
      setIsDocLoading(false)
    }
  }

  if (isLoading) {
    return (
      <div className="text-center py-16">
        <div className="animate-pulse text-gray-400">Загрузка документов...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Документ</th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Обновлён</th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Статус</th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Размер</th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">Чанки</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {docs.map((doc) => (
              <tr
                key={doc.id}
                onClick={() => handleSelectDoc(doc)}
                className={`hover:bg-gray-50 cursor-pointer transition-colors ${
                  selectedDoc?.id === doc.id ? 'bg-indigo-50' : ''
                }`}
              >
                <td className="px-4 py-3 font-medium text-gray-900">{doc.title}</td>
                <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                  {doc.updatedAt.toLocaleString('ru-RU')}
                </td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                      doc.indexed
                        ? 'bg-green-50 text-green-700'
                        : 'bg-yellow-50 text-yellow-700'
                    }`}
                  >
                    {doc.indexed ? 'Проиндексирован' : 'Не проиндексирован'}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-500">{(doc.size / 1024).toFixed(0)} КБ</td>
                <td className="px-4 py-3 text-gray-500">{doc.chunks}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {selectedDoc && (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">{selectedDoc.title}</h3>
            {isDocLoading && (
              <span className="text-sm text-gray-400 animate-pulse">Загрузка...</span>
            )}
          </div>
          <div className="bg-gray-50 rounded-lg p-4 text-sm text-gray-700 whitespace-pre-wrap max-h-96 overflow-y-auto">
            {isDocLoading ? 'Загрузка содержимого...' : selectedDoc.content || 'Содержимое отсутствует'}
          </div>
        </div>
      )}
    </div>
  )
}
