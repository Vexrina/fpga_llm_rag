export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

export interface Chat {
  id: string
  title: string
  messages: Message[]
  createdAt: Date
  updatedAt: Date
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface RagSettings {
  topK: number
  similarityThreshold: number
  model: string
  chunkSize: number
  chunkOverlap: number
  basePrompt: string
  comparisonMethod: string
}

export interface HistoryEntry {
  id: string
  timestamp: Date
  user: string
  field: string
  oldValue: string
  newValue: string
}

export interface LogEntry {
  id: string
  timestamp: Date
  userQuery: string
  topK: number
  model: string
  responseTimeMs: number
  found: boolean
}

export interface KnowledgeDoc {
  id: string
  title: string
  updatedAt: Date
  indexed: boolean
  size: number
  chunks: number
  content: string
}

export interface DocumentVersion {
  id: number
  documentId: string
  title: string
  content: string
  versionNumber: number
  createdAt: string
  createdBy: string
  action: string
}
