import { useState, useEffect, type ChangeEvent } from 'react'
import type { RagSettings, HistoryEntry, KnowledgeDoc, DocumentVersion } from '../types'
import {
  getRagSettings,
  saveRagSettings,
  testRagConnection,
  getHistory,
  getDocuments,
  getDocument,
  getDocumentHistory,
  rollbackDocument,
} from '../mocks/rag'
import { commitDocument, getQueryLogsAPI, type QueryLogEntry, discoverLinks, scrapeUrls } from '../api/graphql'
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
            <option value="l1">L1 distance (Manhattan)</option>
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
  const [logs, setLogs] = useState<QueryLogEntry[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [pageSize] = useState(20)
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingMore, setIsLoadingMore] = useState(false)

  const loadLogs = async (pageNum: number, append: boolean = false) => {
    if (append) {
      setIsLoadingMore(true)
    } else {
      setIsLoading(true)
    }
    try {
      const result = await getQueryLogsAPI(pageNum, pageSize)
      if (append) {
        setLogs((prev) => [...prev, ...result.logs])
      } else {
        setLogs(result.logs)
      }
      setTotal(result.total)
      setPage(result.page)
    } catch (err) {
      console.error('Failed to load logs:', err)
    } finally {
      setIsLoading(false)
      setIsLoadingMore(false)
    }
  }

  useEffect(() => {
    loadLogs(1)
  }, [])

  const hasMore = page * pageSize < total

  if (isLoading && logs.length === 0) {
    return (
      <div className="text-center py-16">
        <div className="animate-pulse text-gray-400">Загрузка логов...</div>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
      <div className="text-sm text-gray-500 px-4 py-2 bg-gray-50 border-b border-gray-200">
        Всего записей: {total}
      </div>
      <table className="w-full text-sm">
        <thead className="bg-gray-50 border-b border-gray-200">
          <tr>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Время</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Запрос</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Модель</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Время ответа</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Статус</th>
            <th className="text-left px-4 py-3 font-medium text-gray-600">Найдено</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {logs.map((log) => (
            <tr key={log.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                {new Date(log.createdAt).toLocaleString('ru-RU')}
              </td>
              <td className="px-4 py-3 max-w-[200px] truncate" title={log.queryText}>
                {log.queryText}
              </td>
              <td className="px-4 py-3 text-gray-600">{log.embeddingModel}</td>
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
              <td className="px-4 py-3 text-gray-600">{log.resultsCount}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {hasMore && (
        <div className="p-4 border-t border-gray-200 flex justify-center">
          <button
            onClick={() => loadLogs(page + 1, true)}
            disabled={isLoadingMore}
            className="px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoadingMore ? 'Загрузка...' : 'Загрузить ещё'}
          </button>
        </div>
      )}
    </div>
  )
}

function KnowledgeTab() {
  const [docs, setDocs] = useState<KnowledgeDoc[]>([])
  const [selectedDoc, setSelectedDoc] = useState<KnowledgeDoc | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isDocLoading, setIsDocLoading] = useState(false)

  const [docHistory, setDocHistory] = useState<DocumentVersion[]>([])
  const [isRollingBack, setIsRollingBack] = useState<number | null>(null)

  const [isAddModalOpen, setIsAddModalOpen] = useState(false)
  const [newDocTitle, setNewDocTitle] = useState('')
  const [newDocSourceType, setNewDocSourceType] = useState<'URL' | 'TEXT' | 'PDF'>('URL')
  const [newDocSourceUrl, setNewDocSourceUrl] = useState('')
  const [newDocUrlMaxDepth, setNewDocUrlMaxDepth] = useState(0)
  const [newDocFileBase64, setNewDocFileBase64] = useState('')
  const [newDocContent, setNewDocContent] = useState('')
  const [isCommitting, setIsCommitting] = useState(false)

  // URL Discovery states
  const [urlStep, setUrlStep] = useState<1 | 2 | 3>(1)
  const [discoveredLinks, setDiscoveredLinks] = useState<string[]>([])
  const [selectedLinks, setSelectedLinks] = useState<Set<string>>(new Set())
  const [isDiscovering, setIsDiscovering] = useState(false)
  const [discoveryStatus, setDiscoveryStatus] = useState('')
  const [isScrapeLoading, setIsScrapeLoading] = useState(false)
  const [scrapeStatus, setScrapeStatus] = useState('')
  const [scrapedPages, setScrapedPages] = useState<{url: string; text: string; title: string}[]>([])

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      const base64 = result.split(',')[1]
      setNewDocFileBase64(base64)
      if (!newDocTitle.trim()) {
        setNewDocTitle(file.name)
      }
    }
    reader.readAsDataURL(file)
  }

  useEffect(() => {
    getDocuments()
      .then((data) => setDocs(data))
      .finally(() => setIsLoading(false))
  }, [])

  const handleSelectDoc = async (doc: KnowledgeDoc) => {
    setSelectedDoc(doc)
    setIsDocLoading(true)
    setDocHistory([])
    try {
      const full = await getDocument(doc.id)
      if (full) setSelectedDoc(full)
      
      const history = await getDocumentHistory(doc.id)
      setDocHistory(history)
    } finally {
      setIsDocLoading(false)
    }
  }

  const handleRollback = async (versionId: number) => {
    if (!selectedDoc) return
    if (!confirm(`Вы уверены, что хотите откатить документ к версии ${versionId}?`)) return
    
    setIsRollingBack(versionId)
    try {
      const result = await rollbackDocument(selectedDoc.id, versionId)
      if (result.success) {
        alert('Документ успешно откачен')
        const full = await getDocument(selectedDoc.id)
        if (full) setSelectedDoc(full)
        const history = await getDocumentHistory(selectedDoc.id)
        setDocHistory(history)
      } else {
        alert('Ошибка: ' + result.message)
      }
    } finally {
      setIsRollingBack(null)
    }
  }

  const handleOpenAddModal = () => {
    setIsAddModalOpen(true)
    setNewDocTitle('')
    setNewDocSourceType('URL')
    setNewDocSourceUrl('')
    setNewDocUrlMaxDepth(0)
    setNewDocFileBase64('')
    setNewDocContent('')
    setUrlStep(1)
    setDiscoveredLinks([])
    setSelectedLinks(new Set())
    setScrapedPages([])
    setDiscoveryStatus('')
    setScrapeStatus('')
  }

  const handleCancelAdd = () => {
    setIsAddModalOpen(false)
    setNewDocTitle('')
    setNewDocSourceType('URL')
    setNewDocSourceUrl('')
    setNewDocUrlMaxDepth(0)
    setNewDocFileBase64('')
    setNewDocContent('')
    setUrlStep(1)
    setDiscoveredLinks([])
    setSelectedLinks(new Set())
    setScrapedPages([])
    setDiscoveryStatus('')
    setScrapeStatus('')
  }

  // URL Discovery handlers
  const handleDiscoverLinks = async () => {
    if (!newDocSourceUrl.trim()) return
    setIsDiscovering(true)
    setDiscoveryStatus('Запуск браузера...')
    try {
      setDiscoveryStatus('Поиск ссылок...')
      const links = await discoverLinks(newDocSourceUrl, newDocUrlMaxDepth)
      const unique = [...new Set(links)]
      setDiscoveredLinks(unique)
      setUrlStep(2)
      setDiscoveryStatus(`Найдено ${unique.length} ссылок`)
    } catch (err) {
      console.error(err)
      setDiscoveryStatus('Ошибка')
      alert('Ошибка при поиске ссылок')
    } finally {
      setIsDiscovering(false)
    }
  }

  const handleToggleLink = (url: string) => {
    const next = new Set(selectedLinks)
    if (next.has(url)) {
      next.delete(url)
    } else {
      next.add(url)
    }
    setSelectedLinks(next)
  }

  const handleSelectAllLinks = (select: boolean) => {
    if (select) {
      setSelectedLinks(new Set(discoveredLinks))
    } else {
      setSelectedLinks(new Set())
    }
  }

  const handleScrapeSelected = async () => {
    if (selectedLinks.size === 0) return
    setIsScrapeLoading(true)
    setScrapeStatus('Скрапинг... 0%')
    setScrapedPages([])
    try {
      const urls = Array.from(selectedLinks)
      const pages: {url: string; text: string; title: string}[] = []
      
      for (let i = 0; i < urls.length; i++) {
        setScrapeStatus(`Скрапинг... ${Math.round((i / urls.length) * 100)}% (${i}/${urls.length})`)
        const result = await scrapeUrls([urls[i]])
        const text = result[urls[i]] || ''
        if (text.trim()) {
          const urlTitle = urls[i].split('/').filter(Boolean).pop() || `страница ${i + 1}`
          pages.push({ 
            url: urls[i], 
            text,
            title: urls.length > 1 ? urlTitle : newDocTitle || urlTitle
          })
        }
      }
      
      setScrapedPages(pages)
      setUrlStep(3)
      setScrapeStatus(`Готово! Собрано ${pages.length} страниц`)
    } catch (err) {
      console.error(err)
      setScrapeStatus('Ошибка')
      alert('Ошибка при скрапинге')
    } finally {
      setIsScrapeLoading(false)
    }
  }

  const handleConfirmAdd = async () => {
    if (!newDocTitle.trim() && scrapedPages.length === 0) return
    
    setIsCommitting(true)
    
    try {
      if (newDocSourceType === 'URL' && scrapedPages.length > 0) {
        // Multiple pages - save each as separate doc
        if (scrapedPages.length > 1) {
          let saved = 0
          for (const page of scrapedPages) {
            if (!page.text.trim() || !page.title.trim()) continue
            
            const result = await commitDocument({
              title: page.title,
              content: page.text,
            })
            
            if (result.commitDocument.success) {
              saved++
            }
          }
          
          if (saved === 0) {
            alert('Не удалось сохранить ни одного документа')
            return
          }
          
          alert(`Сохранено ${saved} документов`)
        } else {
          // Single page - save as one doc
          const result = await commitDocument({
            title: newDocTitle,
            content: scrapedPages[0].text,
          })
          
          if (!result.commitDocument.success) {
            alert('Ошибка: ' + result.commitDocument.message)
            return
          }
        }
      } else if (newDocSourceType === 'TEXT' && newDocContent.trim()) {
        const result = await commitDocument({
          title: newDocTitle,
          content: newDocContent,
        })
        
        if (!result.commitDocument.success) {
          alert('Ошибка: ' + result.commitDocument.message)
          return
        }
      } else if (newDocSourceType === 'PDF' && newDocFileBase64.trim()) {
        const result = await commitDocument({
          title: newDocTitle,
          content: newDocContent,
        })
        
        if (!result.commitDocument.success) {
          alert('Ошибка: ' + result.commitDocument.message)
          return
        }
      } else {
        alert('Нет содержимого для сохранения')
        return
      }
      
      setIsAddModalOpen(false)
      const data = await getDocuments()
      setDocs(data)
    } catch (err) {
      console.error(err)
      alert('Ошибка при сохранении: ' + err)
    } finally {
      setIsCommitting(false)
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
      <div className="flex justify-end">
        <button
          onClick={handleOpenAddModal}
          className="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
        >
          Добавить документ
        </button>
      </div>

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

      {selectedDoc && docHistory.length > 0 && (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-200 bg-gray-50">
            <h3 className="text-lg font-semibold text-gray-900">История версий</h3>
          </div>
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Версия</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Дата</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Автор</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Действие</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Содержимое</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {docHistory.map((version) => (
                <tr key={version.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <span className="inline-block bg-indigo-50 text-indigo-700 px-2 py-0.5 rounded text-xs font-medium">
                      v{version.versionNumber}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                    {new Date(version.createdAt).toLocaleString('ru-RU')}
                  </td>
                  <td className="px-4 py-3">{version.createdBy}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                        version.action === 'create'
                          ? 'bg-green-50 text-green-700'
                          : version.action === 'rollback'
                          ? 'bg-blue-50 text-blue-700'
                          : 'bg-gray-100 text-gray-700'
                      }`}
                    >
                      {version.action === 'create' ? 'Создан' : version.action === 'rollback' ? 'Откат' : 'Изменён'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-gray-500 max-w-xs truncate" title={version.content}>
                    {version.content.slice(0, 80)}...
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleRollback(version.id)}
                      disabled={isRollingBack === version.id}
                      className="text-indigo-600 hover:text-indigo-800 text-sm font-medium disabled:opacity-50"
                    >
                      {isRollingBack === version.id ? 'Откат...' : 'Откатить'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {isAddModalOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-hidden">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">Добавить документ</h2>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Название документа
                </label>
                <input
                  type="text"
                  value={newDocTitle}
                  onChange={(e) => setNewDocTitle(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Введите название документа"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Тип источника
                </label>
                <select
                  value={newDocSourceType}
                  onChange={(e) => setNewDocSourceType(e.target.value as 'URL' | 'TEXT' | 'PDF')}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                >
                  <option value="URL">URL</option>
                  <option value="TEXT">Текст</option>
                  <option value="PDF">Файл (PDF)</option>
                </select>
              </div>

              {newDocSourceType === 'URL' ? (
                <div className="space-y-4">
                  {/* Step 1: Discover */}
                  {urlStep === 1 && (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          URL документа
                        </label>
                        <input
                          type="text"
                          value={newDocSourceUrl}
                          onChange={(e) => setNewDocSourceUrl(e.target.value)}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                          placeholder="https://example.com"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Глубина поиска
                        </label>
                        <select
                          value={newDocUrlMaxDepth}
                          onChange={(e) => setNewDocUrlMaxDepth(parseInt(e.target.value))}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                        >
                          <option value={0}>Только эта страница</option>
                          <option value={1}>+1 уровень</option>
                          <option value={2}>+2 уровня</option>
                        </select>
                      </div>
                      {discoveryStatus && (
                        <div className={`text-sm p-2 rounded ${discoveryStatus.includes('Ошибка') ? 'bg-red-50 text-red-700' : 'bg-blue-50 text-blue-700'}`}>
                          {discoveryStatus}
                        </div>
                      )}
                      <button
                        onClick={handleDiscoverLinks}
                        disabled={!newDocSourceUrl.trim() || isDiscovering}
                        className="w-full px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isDiscovering ? 'Поиск...' : 'Найти ссылки'}
                      </button>
                    </>
                  )}

                  {/* Step 2: Select */}
                  {urlStep === 2 && (
                    <>
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-gray-700">
                          Найдено: {discoveredLinks.length}
                        </span>
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleSelectAllLinks(true)}
                            className="text-xs text-indigo-600 hover:text-indigo-800"
                          >
                            Все
                          </button>
                          <button
                            onClick={() => handleSelectAllLinks(false)}
                            className="text-xs text-gray-500 hover:text-gray-700"
                          >
                            Ничего
                          </button>
                        </div>
                      </div>
                      <div className="max-h-48 overflow-y-auto border border-gray-200 rounded-lg">
                        {discoveredLinks.map((link) => (
                          <label
                            key={link}
                            className="flex items-start gap-2 p-2 hover:bg-gray-50 border-b border-gray-100 last:border-b-0 cursor-pointer"
                          >
                            <input
                              type="checkbox"
                              checked={selectedLinks.has(link)}
                              onChange={() => handleToggleLink(link)}
                              className="mt-1 rounded border-gray-300 text-indigo-600"
                            />
                            <span className="text-xs text-gray-600 truncate">{link}</span>
                          </label>
                        ))}
                      </div>
                      <div className="flex gap-2">
                        <button
                          onClick={() => setUrlStep(1)}
                          className="flex-1 px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200"
                        >
                          Назад
                        </button>
                        <button
                          onClick={handleScrapeSelected}
                          disabled={selectedLinks.size === 0 || isScrapeLoading}
                          className="flex-1 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 disabled:opacity-50"
                        >
                          {isScrapeLoading ? 'Скрапинг...' : `Скрапить (${selectedLinks.size})`}
                        </button>
                      </div>
                      {scrapeStatus && (
                        <div className={`text-sm p-2 rounded ${scrapeStatus.includes('Ошибка') ? 'bg-red-50 text-red-700' : 'bg-blue-50 text-blue-700'}`}>
                          {scrapeStatus}
                        </div>
                      )}
                    </>
                  )}

                  {/* Step 3: Edit & Save */}
                  {urlStep === 3 && (
                    <div className="space-y-4">
                      <div className="flex gap-2">
                        <button
                          onClick={() => setUrlStep(2)}
                          className="flex-1 px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200"
                        >
                          Назад к выбору
                        </button>
                      </div>
                      <div className="text-sm font-medium text-gray-700">
                        Собрано страниц: {scrapedPages.length}
                      </div>
                      <div className="max-h-64 overflow-y-auto space-y-3">
                        {scrapedPages.map((page, idx) => (
                          <div key={idx} className="border border-gray-200 rounded-lg p-3 space-y-2">
                            <div className="text-xs text-gray-500 truncate" title={page.url}>
                              {idx + 1}. {page.url}
                            </div>
                            <input
                              type="text"
                              value={page.title}
                              onChange={(e) => {
                                const updated = [...scrapedPages]
                                updated[idx].title = e.target.value
                                setScrapedPages(updated)
                              }}
                              className="w-full px-2 py-1 border border-gray-200 rounded text-sm font-medium"
                              placeholder="Назван��е документа"
                            />
                            <textarea
                              value={page.text}
                              onChange={(e) => {
                                const updated = [...scrapedPages]
                                updated[idx].text = e.target.value
                                setScrapedPages(updated)
                              }}
                              className="w-full h-24 px-2 py-1 border border-gray-200 rounded text-xs font-mono resize-none"
                              placeholder="Содержимое"
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ) : newDocSourceType === 'TEXT' ? (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Содержимое
                  </label>
                  <textarea
                    value={newDocContent}
                    onChange={(e) => setNewDocContent(e.target.value)}
                    className="w-full h-32 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 font-mono"
                    placeholder="Введите содержимое"
                  />
                </div>
              ) : (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Загрузить файл (PDF)
                  </label>
                  <input
                    type="file"
                    accept=".pdf"
                    onChange={handleFileChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 file:mr-4 file:px-4 file:py-2 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
                  />
                  {newDocFileBase64 && (
                    <p className="text-sm text-green-600 mt-1">Файл загружен</p>
                  )}
                </div>
              )}
            </div>
            <div className="p-6 border-t border-gray-200 flex justify-end gap-3">
              <button
                onClick={handleCancelAdd}
                className="px-4 py-2 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-100"
              >
                Отмена
              </button>
              <button
                onClick={handleConfirmAdd}
                disabled={
                  (newDocSourceType === 'URL' && scrapedPages.length === 0) ||
                  (newDocSourceType === 'TEXT' && !newDocContent.trim()) ||
                  (newDocSourceType === 'PDF' && !newDocFileBase64.trim()) ||
                  isCommitting
                }
                className="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isCommitting ? 'Сохранение...' : 'Сохранить'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
