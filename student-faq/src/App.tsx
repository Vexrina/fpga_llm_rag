import { Routes, Route, Navigate } from 'react-router-dom'
import { useState } from 'react'
import Layout from './components/Layout'
import ChatPage from './pages/ChatPage'
import LoginPage from './pages/LoginPage'
import RagSettingsPage from './pages/RagSettingsPage'

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

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
              <LoginPage onLogin={() => setIsAuthenticated(true)} />
            )
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
