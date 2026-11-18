import { Link } from 'react-router-dom'
import { useDocumentTitle } from '../hooks/useFetch'
import { Icon } from '../components/Icons'
import { SearchBox } from '../components/SearchBox'

export default function NotFoundPage() {
  useDocumentTitle('Страница не найдена')

  return (
    <div className="page page-narrow" style={{ paddingTop: 'clamp(32px, 6vw, 80px)' }}>
      <p className="mono dim" style={{ fontSize: 13, letterSpacing: '0.1em' }}>
        404
      </p>
      <h1 style={{ marginTop: 8 }}>Такой страницы нет в атласе</h1>
      <p className="muted" style={{ maxWidth: 560 }}>
        Возможно, ссылка устарела или публикация была удалена из корпуса. Попробуйте найти нужное через
        поиск или вернитесь на главную.
      </p>

      <div style={{ margin: 'var(--space-5) 0' }}>
        <SearchBox size="md" />
      </div>

      <div className="row wrap" style={{ gap: 10 }}>
        <Link to="/" className="btn">
          <Icon name="arrowLeft" size={15} /> На главную
        </Link>
        <Link to="/search" className="btn btn-ghost">
          К поиску
        </Link>
        <Link to="/dashboard" className="btn btn-ghost">
          Аналитика
        </Link>
      </div>
    </div>
  )
}
