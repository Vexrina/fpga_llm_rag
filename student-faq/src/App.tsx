import { Routes, Route, Navigate } from 'react-router-dom'
import { useState, useEffect, useCallback } from 'react'
import Layout from './components/Layout'
import ChatPage from './pages/ChatPage'
import LoginPage from './pages/LoginPage'
import RagSettingsPage from './pages/RagSettingsPage'

const INACTIVITY_TIMEOUT = 2 * 60 * 60 * 1000 // 2 hours in ms

function getLastActiveTime(): number {
  const stored = localStorage.getItem('last_active_time')
  return stored ? parseInt(stored, 10) : 0
}

function setLastActiveTime() {
  localStorage.setItem('last_active_time', Date.now().toString())
}

function isTokenExpired(): boolean {
  const token = localStorage.getItem('admin_token')
  if (!token) return true
  
  const lastActive = getLastActiveTime()
  if (!lastActive) return true
  
  return Date.now() - lastActive > INACTIVITY_TIMEOUT
}

function clearAuth() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_username')
  localStorage.removeItem('last_active_time')
}

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    if (isTokenExpired()) {
      clearAuth()
    } else {
      setLastActiveTime()
      setIsAuthenticated(true)
    }
  }, [])

  const handleActivity = useCallback(() => {
    if (localStorage.getItem('admin_token')) {
      setLastActiveTime()
    }
  }, [])

  useEffect(() => {
    const events = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart']
    events.forEach(event => {
      document.addEventListener(event, handleActivity)
    })
    return () => {
      events.forEach(event => {
        document.removeEventListener(event, handleActivity)
      })
    }
  }, [handleActivity])

  const handleLogin = useCallback(() => {
    setLastActiveTime()
    setIsAuthenticated(true)
  }, [])

  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<ChatPage />} />
        <Route path="chat" element={<ChatPage />} />
        <Route
          path="admin"
          element={isAuthenticated ? <RagSettingsPage /> : <Navigate to="/login" replace />}
        />
        <Route
          path="login"
          element={
            isAuthenticated ? (
              <Navigate to="/admin" replace />
            ) : (
              <LoginPage onLogin={handleLogin} />
            )
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}