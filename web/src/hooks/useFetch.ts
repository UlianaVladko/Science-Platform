import { useCallback, useEffect, useRef, useState } from 'react'
import { ApiError } from '../api/client'

interface State<T> {
  data: T | null
  loading: boolean
  error: string | null
}

export function useFetch<T>(
  loader: () => Promise<T>,
  deps: unknown[],
  options: { enabled?: boolean } = {},
): State<T> & { reload: () => void } {
  const enabled = options.enabled ?? true
  const [state, setState] = useState<State<T>>({ data: null, loading: enabled, error: null })
  const [nonce, setNonce] = useState(0)
  const alive = useRef(true)
  const loaderRef = useRef(loader)
  loaderRef.current = loader

  useEffect(() => {
    alive.current = true
    return () => {
      alive.current = false
    }
  }, [])

  useEffect(() => {
    if (!enabled) {
      setState({ data: null, loading: false, error: null })
      return
    }
    let cancelled = false
    setState((prev) => ({ ...prev, loading: true, error: null }))
    loaderRef
      .current()
      .then((data) => {
        if (!cancelled && alive.current) setState({ data, loading: false, error: null })
      })
      .catch((err: unknown) => {
        if (cancelled || !alive.current) return
        const message = err instanceof ApiError ? err.message : err instanceof Error ? err.message : 'Неизвестная ошибка'
        setState({ data: null, loading: false, error: message })
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, nonce, enabled])

  const reload = useCallback(() => setNonce((n) => n + 1), [])
  return { ...state, reload }
}

export function useDebounced<T>(value: T, delay = 250): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}

export function useDocumentTitle(title: string) {
  useEffect(() => {
    document.title = title ? `${title} — Научный атлас` : 'Научный атлас'
  }, [title])
}

export function useIntersection<T extends Element>(rootMargin = '200px') {
  const ref = useRef<T | null>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const node = ref.current
    if (!node) return
    if (typeof IntersectionObserver === 'undefined') {
      setVisible(true)
      return
    }
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setVisible(true)
            observer.disconnect()
          }
        }
      },
      { rootMargin },
    )
    observer.observe(node)
    return () => observer.disconnect()
  }, [rootMargin])

  return { ref, visible }
}
