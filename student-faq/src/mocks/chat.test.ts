import { describe, it, expect, vi, beforeEach } from 'vitest'
import { sendMessage, getChatHistory } from '../mocks/chat'

describe('chat mocks', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  it('sendMessage returns a message with assistant role', async () => {
    const promise = sendMessage('test question')
    await vi.advanceTimersByTimeAsync(2000)
    const result = await promise

    expect(result.role).toBe('assistant')
    expect(result.content).toBe('test question')
    expect(result.id).toBeDefined()
    expect(result.timestamp).toBeInstanceOf(Date)
  })

  it('getChatHistory returns array of messages', async () => {
    const promise = getChatHistory()
    await vi.advanceTimersByTimeAsync(1000)
    const result = await promise

    expect(Array.isArray(result)).toBe(true)
    expect(result.length).toBeGreaterThan(0)
    expect(result[0]).toHaveProperty('id')
    expect(result[0]).toHaveProperty('role')
    expect(result[0]).toHaveProperty('content')
  })
})
