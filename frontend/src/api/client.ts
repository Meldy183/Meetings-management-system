// In production (Docker / nginx): VITE_API_URL is unset → empty string → relative URLs go through nginx proxy.
// In dev (npm run dev): Vite's server.proxy forwards these paths to localhost:8080, keeping cookies same-origin.
export const BASE_URL = import.meta.env.VITE_API_URL ?? ''

export class ApiError extends Error {
  status: number
  body: unknown
  constructor(status: number, body: unknown) {
    super(`API error ${status}`)
    this.status = status
    this.body = body
  }
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new ApiError(401, {})
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export async function apiFetchBlob(path: string): Promise<{ blob: Blob; filename: string }> {
  const res = await fetch(`${BASE_URL}${path}`, { credentials: 'include' })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new ApiError(401, {})
  }
  if (!res.ok) throw new ApiError(res.status, {})
  const blob = await res.blob()
  const disposition = res.headers.get('Content-Disposition') ?? ''
  const match = disposition.match(/filename="?([^"]+)"?/)
  const filename = match?.[1] ?? 'document.docx'
  return { blob, filename }
}

export async function authLogin(username: string, password: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
}

export async function authLogout(): Promise<void> {
  try {
    await fetch(`${BASE_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    })
  } finally {
    window.location.href = '/login'
  }
}
