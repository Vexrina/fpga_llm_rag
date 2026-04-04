import type { LoginCredentials } from '../types'

export async function login(credentials: LoginCredentials): Promise<{ token: string; username: string }> {
  await new Promise((resolve) => setTimeout(resolve, 500))

  if (!credentials.username || !credentials.password) {
    throw new Error('Заполните все поля')
  }

  return {
    token: `mock-token-${Date.now()}`,
    username: credentials.username,
  }
}

export async function logout(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 200))
}

export async function validateSession(token: string): Promise<boolean> {
  await new Promise((resolve) => setTimeout(resolve, 200))
  return !!token
}
