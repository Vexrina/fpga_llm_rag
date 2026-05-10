const GRAPHQL_ENDPOINT = '/graphql'

async function graphqlRequest<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  const response = await fetch(GRAPHQL_ENDPOINT, {
    method: 'POST',
    mode: 'cors',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({ query, variables }),
  })

  const result = await response.json()

  if (result.errors) {
    console.error('GraphQL errors:', result.errors)
    const errorMessage = result.errors[0]?.message || 'Unknown GraphQL error'
    throw new Error(errorMessage)
  }

  if (!result.data) {
    console.error('No data in response:', result)
    throw new Error('No data in response')
  }

  return result.data
}

export interface PreviewDocumentInput {
  title: string
  sourceType: 'URL' | 'TEXT' | 'PDF'
  sourceUrl?: string
  contentBase64?: string
  urlMaxDepth?: number
}

export interface DiscoverLinksResult {
  discoverLinks: {
    links: string[]
  }
}

export interface ScrapeUrlsResult {
  scrapeUrls: {
    texts: { url: string; text: string }[]
  }
}

export async function discoverLinks(url: string, maxDepth: number): Promise<string[]> {
  const query = `
    mutation DiscoverLinks($url: String!, $maxDepth: Int!) {
      discoverLinks(url: $url, maxDepth: $maxDepth) {
        links
      }
    }
  `
  const result = await graphqlRequest<DiscoverLinksResult>(query, { url, maxDepth })
  return result.discoverLinks.links
}

export async function scrapeUrls(urls: string[]): Promise<Record<string, string>> {
  const query = `
    mutation ScrapeUrls($urls: [String!]!) {
      scrapeUrls(urls: $urls) {
        texts {
          url
          text
        }
      }
    }
  `
  const result = await graphqlRequest<ScrapeUrlsResult>(query, { urls })
  const texts: Record<string, string> = {}
  for (const t of result.scrapeUrls.texts) {
    texts[t.url] = t.text
  }
  return texts
}

export interface PreviewResult {
  previewDocument: {
    extractedText: string
    pagesExtracted: number
  }
}

export interface CommitDocumentInput {
  title: string
  content?: string
  metadata?: { key: string; value: string }[]
  sourceType?: 'URL' | 'TEXT' | 'PDF'
  contentBase64?: string
}

export interface CommitResult {
  commitDocument: {
    success: boolean
    message: string
    id: string
  }
}

export async function commitDocument(input: CommitDocumentInput): Promise<CommitResult> {
  const query = `
    mutation CommitDocument($input: CommitDocumentInput!) {
      commitDocument(input: $input) {
        success
        message
        id
      }
    }
  `
  return graphqlRequest<CommitResult>(query, { input })
}

export async function previewDocument(input: PreviewDocumentInput): Promise<PreviewResult> {
  const query = `
    mutation PreviewDocument($input: PreviewDocumentInput!) {
      previewDocument(input: $input) {
        extractedText
        pagesExtracted
      }
    }
  `
  return graphqlRequest<PreviewResult>(query, { input })
}

export interface AskResult {
  ask: string
}

export async function askQuestion(question: string): Promise<string> {
  const query = `
    query Ask($question: String!) {
      ask(question: $question)
    }
  `
  const result = await graphqlRequest<AskResult>(query, { question })
  return result.ask
}

export interface SettingEntry {
  key: string
  value: string
}

export interface GetSettingsResult {
  getRagSettings: SettingEntry[]
}

export interface UpdateSettingResult {
  updateRagSetting: {
    success: boolean
    message: string
  }
}

export async function getRagSettingsAPI(): Promise<Record<string, string>> {
  const query = `
    query GetRagSettings {
      getRagSettings {
        key
        value
      }
    }
  `
  const result = await graphqlRequest<GetSettingsResult>(query)
  const settings: Record<string, string> = {}
  for (const entry of result.getRagSettings) {
    settings[entry.key] = entry.value
  }
  return settings
}

export async function updateRagSettingAPI(key: string, value: string, changedBy?: string): Promise<boolean> {
  const query = `
    mutation UpdateRagSetting($key: String!, $value: String!, $changedBy: String) {
      updateRagSetting(key: $key, value: $value, changedBy: $changedBy) {
        success
        message
      }
    }
  `
  const result = await graphqlRequest<UpdateSettingResult>(query, { key, value, changedBy })
  return result.updateRagSetting.success
}

export interface SettingsHistoryEntry {
  id: number
  settingKey: string
  oldValue: string
  newValue: string
  changedBy: string
  changedAt?: string
}

export interface GetSettingsHistoryResult {
  getRagSettingsHistory: SettingsHistoryEntry[]
}

export async function getRagSettingsHistoryAPI(limit?: number): Promise<SettingsHistoryEntry[]> {
  const query = `
    query GetRagSettingsHistory($limit: Int) {
      getRagSettingsHistory(limit: $limit) {
        id
        settingKey
        oldValue
        newValue
        changedBy
        changedAt
      }
    }
  `
  const result = await graphqlRequest<GetSettingsHistoryResult>(query, { limit })
  return result.getRagSettingsHistory
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

export interface GetDocumentHistoryResult {
  getDocumentHistory: DocumentVersion[]
}

export async function getDocumentHistoryAPI(documentId: string, limit?: number): Promise<DocumentVersion[]> {
  const query = `
    query GetDocumentHistory($documentId: String!, $limit: Int) {
      getDocumentHistory(documentId: $documentId, limit: $limit) {
        id
        documentId
        title
        content
        versionNumber
        createdAt
        createdBy
        action
      }
    }
  `
  const result = await graphqlRequest<GetDocumentHistoryResult>(query, { documentId, limit })
  return result.getDocumentHistory
}

export interface RollbackResult {
  success: boolean
  message: string
  newVersionId: string
}

export interface RollbackDocumentResult {
  rollbackDocument: RollbackResult
}

export async function rollbackDocumentAPI(documentId: string, versionId: number, rollbackBy?: string): Promise<RollbackResult> {
  const query = `
    mutation RollbackDocument($documentId: String!, $versionId: Int!, $rollbackBy: String) {
      rollbackDocument(documentId: $documentId, versionId: $versionId, rollbackBy: $rollbackBy) {
        success
        message
        newVersionId
      }
    }
  `
  const result = await graphqlRequest<RollbackDocumentResult>(query, { documentId, versionId, rollbackBy })
  return result.rollbackDocument
}

export interface DocumentListItem {
  id: string
  title: string
  updatedAt: string
  indexed: boolean
  size: number
  chunks: number
}

export interface GetDocumentsResult {
  getDocuments: DocumentListItem[]
}

export async function getAllDocumentsAPI(): Promise<DocumentListItem[]> {
  const query = `
    query GetDocuments {
      getDocuments {
        id
        title
        updatedAt
        indexed
        size
        chunks
      }
    }
  `
  const result = await graphqlRequest<GetDocumentsResult>(query)
  return result.getDocuments
}

export interface DocumentDetail {
  id: string
  title: string
  content: string
}

export interface GetDocumentResult {
  getDocument: DocumentDetail
}

export async function getDocumentByIdAPI(id: string): Promise<DocumentDetail | null> {
  const query = `
    query GetDocument($id: String!) {
      getDocument(id: $id) {
        id
        title
        content
      }
    }
  `
  const result = await graphqlRequest<GetDocumentResult>(query, { id })
  return result.getDocument
}

export interface QueryLogEntry {
  id: number
  queryText: string
  embeddingModel: string
  responseTimeMs: number
  found: boolean
  resultsCount: number
  createdAt: string
}

export interface GetQueryLogsResult {
  getQueryLogs: {
    logs: QueryLogEntry[]
    total: number
    page: number
    pageSize: number
  }
}

export interface UpdateDocumentResult {
  updateDocument: {
    success: boolean
    message: string
  }
}

export async function updateDocumentAPI(id: string, title: string, content: string, updatedBy?: string): Promise<{ success: boolean; message: string }> {
  const query = `
    mutation UpdateDocument($id: String!, $title: String!, $content: String!, $updatedBy: String) {
      updateDocument(id: $id, title: $title, content: $content, updatedBy: $updatedBy) {
        success
        message
      }
    }
  `
  const result = await graphqlRequest<UpdateDocumentResult>(query, { id, title, content, updatedBy })
  return result.updateDocument
}

export interface DeleteDocumentResult {
  deleteDocument: {
    success: boolean
    message: string
  }
}

export async function deleteDocumentAPI(id: string): Promise<{ success: boolean; message: string }> {
  const query = `
    mutation DeleteDocument($id: String!) {
      deleteDocument(id: $id) {
        success
        message
      }
    }
  `
  const result = await graphqlRequest<DeleteDocumentResult>(query, { id })
  return result.deleteDocument
}

export async function getQueryLogsAPI(page: number = 1, pageSize: number = 20): Promise<{
  logs: QueryLogEntry[]
  total: number
  page: number
  pageSize: number
}> {
  const query = `
    query GetQueryLogs($page: Int, $pageSize: Int) {
      getQueryLogs(page: $page, pageSize: $pageSize) {
        logs {
          id
          queryText
          embeddingModel
          responseTimeMs
          found
          resultsCount
          createdAt
        }
        total
        page
        pageSize
      }
    }
  `
  const result = await graphqlRequest<GetQueryLogsResult>(query, { page, pageSize })
  return result.getQueryLogs
}

export interface AdminLoginResult {
  adminLogin: {
    token: string
    expiresAt: string
    admin: {
      id: number
      username: string
      role: string
    }
    success: boolean
    message: string
  }
}

export async function adminLoginAPI(username: string, password: string): Promise<{ token: string; username: string }> {
  const query = `
    mutation AdminLogin($username: String!, $password: String!) {
      adminLogin(username: $username, password: $password) {
        token
        expiresAt
        admin {
          id
          username
          role
        }
        success
        message
      }
    }
  `
  const result = await graphqlRequest<AdminLoginResult>(query, { username, password })
  if (!result.adminLogin.success) {
    throw new Error(result.adminLogin.message || 'Ошибка входа')
  }
  return {
    token: result.adminLogin.token,
    username: result.adminLogin.admin.username,
  }
}

export interface ValidateTokenResult {
  validateToken: {
    id: number
    username: string
    role: string
  } | null
}

export async function validateTokenAPI(token: string): Promise<{ id: number; username: string; role: string } | null> {
  const query = `
    query ValidateToken($token: String!) {
      validateToken(token: $token) {
        id
        username
        role
      }
    }
  `
  const result = await graphqlRequest<ValidateTokenResult>(query, { token })
  return result.validateToken
}