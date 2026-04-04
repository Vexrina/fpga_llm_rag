import { describe, it, expect, vi } from 'vitest'
import { getRagSettings, saveRagSettings, testRagConnection, getHistory, getLogs, getDocuments, getDocument } from '../mocks/rag'

describe('rag mocks', () => {
  it('getRagSettings returns default settings with all fields', async () => {
    const result = await getRagSettings()

    expect(result).toHaveProperty('topK')
    expect(result).toHaveProperty('similarityThreshold')
    expect(result).toHaveProperty('model')
    expect(result).toHaveProperty('chunkSize')
    expect(result).toHaveProperty('chunkOverlap')
    expect(result).toHaveProperty('basePrompt')
    expect(result).toHaveProperty('comparisonMethod')
  })

  it('saveRagSettings returns saved settings', async () => {
    const settings = {
      topK: 10,
      similarityThreshold: 0.8,
      model: 'test-model',
      chunkSize: 256,
      chunkOverlap: 32,
      basePrompt: 'test',
      comparisonMethod: 'dot',
    }

    const result = await saveRagSettings(settings)

    expect(result.topK).toBe(10)
    expect(result.similarityThreshold).toBe(0.8)
    expect(result.model).toBe('test-model')
  })

  it('testRagConnection returns success', async () => {
    const result = await testRagConnection()

    expect(result.ok).toBe(true)
    expect(result.message).toBeDefined()
  })

  it('getHistory returns array of entries', async () => {
    const result = await getHistory()

    expect(Array.isArray(result)).toBe(true)
    expect(result.length).toBeGreaterThan(0)
    expect(result[0]).toHaveProperty('field')
    expect(result[0]).toHaveProperty('oldValue')
    expect(result[0]).toHaveProperty('newValue')
  })

  it('getLogs returns array of entries', async () => {
    const result = await getLogs()

    expect(Array.isArray(result)).toBe(true)
    expect(result.length).toBeGreaterThan(0)
    expect(result[0]).toHaveProperty('userQuery')
    expect(result[0]).toHaveProperty('found')
  })

  it('getDocuments returns array of docs', async () => {
    const result = await getDocuments()

    expect(Array.isArray(result)).toBe(true)
    expect(result.length).toBeGreaterThan(0)
    expect(result[0]).toHaveProperty('title')
    expect(result[0]).toHaveProperty('indexed')
    expect(result[0]).toHaveProperty('content')
  })

  it('getDocument returns doc by id', async () => {
    const docs = await getDocuments()
    const result = await getDocument(docs[0].id)

    expect(result).not.toBeNull()
    expect(result?.id).toBe(docs[0].id)
  })

  it('getDocument returns null for unknown id', async () => {
    const result = await getDocument('nonexistent')

    expect(result).toBeNull()
  })
})
