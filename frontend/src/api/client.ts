export const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

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
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export async function apiFetchBlob(path: string): Promise<{ blob: Blob; filename: string }> {
  const res = await fetch(`${BASE_URL}${path}`)
  if (!res.ok) throw new ApiError(res.status, {})
  const blob = await res.blob()
  const disposition = res.headers.get('Content-Disposition') ?? ''
  const match = disposition.match(/filename="?([^"]+)"?/)
  const filename = match?.[1] ?? 'document.docx'
  return { blob, filename }
}
