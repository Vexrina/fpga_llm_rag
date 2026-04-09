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
    throw new Error(result.errors[0].message)
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

export interface PreviewResult {
  previewDocument: {
    extractedText: string
    pagesExtracted: number
  }
}

export interface CommitDocumentInput {
  title: string
  content: string
  metadata?: { key: string; value: string }[]
}

export interface CommitResult {
  commitDocument: {
    success: boolean
    message: string
    id: string
  }
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
  return graphqlRequest<CommitResult>(query, { input: { title: input.title, content: input.content } })
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