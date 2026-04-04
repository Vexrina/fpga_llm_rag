import { describe, it, expect, vi, beforeEach } from 'vitest'
import { login, logout, validateSession } from '../mocks/auth'

describe('auth mocks', () => {
  it('login returns token and username', async () => {
    const result = await login({ username: 'admin', password: 'password123' })

    expect(result.token).toBeDefined()
    expect(result.username).toBe('admin')
  })

  it('login throws error for empty credentials', async () => {
    await expect(login({ username: '', password: '' })).rejects.toThrow('Заполните все поля')
  })

  it('logout resolves without error', async () => {
    await expect(logout()).resolves.toBeUndefined()
  })

  it('validateSession returns true for valid token', async () => {
    const result = await validateSession('some-token')

    expect(result).toBe(true)
  })

  it('validateSession returns false for empty token', async () => {
    const result = await validateSession('')

    expect(result).toBe(false)
  })
})
