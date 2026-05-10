import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'

const INACTIVITY_TIMEOUT = 2 * 60 * 60 * 1000 // 2 hours in ms

function isTokenExpired(): boolean {
  const token = localStorage.getItem('admin_token')
  if (!token) return true
  
  const lastActive = parseInt(localStorage.getItem('last_active_time') || '0', 10)
  if (!lastActive) return true
  
  return Date.now() - lastActive > INACTIVITY_TIMEOUT
}

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      if (isTokenExpired()) {
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_username')
        localStorage.removeItem('last_active_time')
        setIsAdmin(false)
        if (location.pathname === '/admin') {
          navigate('/login')
        }
      } else {
        setIsAdmin(true)
      }
    } else {
      setIsAdmin(false)
    }
  }, [location, navigate])

  const navItems = [
    { path: '/chat', label: 'Чат' },
  ]

  if (isAdmin) {
    navItems.push({ path: '/admin', label: 'Настройки RAG' })
  } else {
    navItems.push({ path: '/login', label: 'Вход' })
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <nav className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-indigo-600">
            Student FAQ
          </Link>
          <div className="flex gap-4">
            {navItems.map((item) => {
              const isActive =
                item.path === '/'
                  ? location.pathname === '/'
                  : location.pathname.startsWith(item.path)

              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-indigo-100 text-indigo-700'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                  }`}
                >
                  {item.label}
                </Link>
              )
            })}
          </div>
        </nav>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>

      <footer className="bg-white border-t border-gray-200 py-4">
        <p className="text-center text-sm text-gray-500">
          &copy; 2026 Student FAQ
        </p>
      </footer>
    </div>
  )
}