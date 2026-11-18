import { NavLink, Route, Routes, useLocation } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { api } from './api/client'
import type { Health } from './api/types'
import { Icon } from './components/Icons'
import HomePage from './pages/HomePage'
import SearchPage from './pages/SearchPage'
import PublicationPage from './pages/PublicationPage'
import AuthorPage from './pages/AuthorPage'
import DashboardPage from './pages/DashboardPage'
import QuestionsPage, { QuestionDetailPage } from './pages/QuestionsPage'
import AddPublicationPage from './pages/AddPublicationPage'
import NotFoundPage from './pages/NotFoundPage'
import './styles/pages.css'

const NAV = [
  { to: '/', label: 'Главная', end: true },
  { to: '/search', label: 'Поиск' },
  { to: '/dashboard', label: 'Аналитика' },
  { to: '/questions', label: 'Вопросы' },
  { to: '/add', label: 'Добавить' },
]

function UpstreamBadge() {
  const [health, setHealth] = useState<Health | null>(null)

  useEffect(() => {
    let cancelled = false
    const load = () => {
      api
        .health()
        .then((value) => {
          if (!cancelled) setHealth(value)
        })
        .catch(() => {
          if (!cancelled) setHealth(null)
        })
    }
    load()
    const timer = setInterval(load, 30000)
    return () => {
      cancelled = true
      clearInterval(timer)
    }
  }, [])

  if (!health) return null

  const up = health.upstream === 'up'
  return (
    <span
      className={up ? 'chip badge-api' : 'chip badge-demo'}
      title={
        up
          ? 'Сервис API доступен: публикации и поиск идут из согласованного контракта'
          : 'Сервис API недоступен: показан демо-корпус BFF'
      }
    >
      <span
        style={{
          width: 7,
          height: 7,
          borderRadius: 999,
          background: up ? 'var(--teal)' : 'var(--amber)',
          boxShadow: up ? '0 0 8px var(--teal)' : '0 0 8px var(--amber)',
        }}
      />
      {up ? 'API онлайн' : 'демо-режим'}
    </span>
  )
}

function ScrollToTop() {
  const { pathname } = useLocation()
  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'instant' as ScrollBehavior })
  }, [pathname])
  return null
}

export default function App() {
  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="app-header-inner">
          <NavLink to="/" className="brand">
            <span className="brand-mark">
              <Icon name="graph" size={20} />
            </span>
            <span className="brand-text">
              Научный <span>атлас</span>
            </span>
          </NavLink>

          <nav className="app-nav">
            {NAV.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div className="app-header-side">
            <UpstreamBadge />
          </div>
        </div>
      </header>

      <main className="app-main">
        <ScrollToTop />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/publications/:id" element={<PublicationPage />} />
          <Route path="/authors/:id" element={<AuthorPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/questions" element={<QuestionsPage />} />
          <Route path="/questions/:id" element={<QuestionDetailPage />} />
          <Route path="/add" element={<AddPublicationPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>

      <footer className="app-footer">
        <div className="app-footer-inner">
          <span>Научный атлас · демонстрационный фронтенд команды UI</span>
          <span className="dim">
            Источники: согласованный API (20 операций) · демо-корпус BFF · состояние BFF
          </span>
        </div>
      </footer>
    </div>
  )
}
