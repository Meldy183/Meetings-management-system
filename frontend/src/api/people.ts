import { apiFetch, BASE_URL } from './client'
import { ApiError } from './client'
import type { Person, PersonCreate } from './types'

export function getPeople(q?: string, order?: string): Promise<Person[]> {
  const params = new URLSearchParams()
  if (q) params.set('q', q)
  if (order) params.set('order', order)
  const qs = params.toString()
  return apiFetch<Person[]>(qs ? `/people?${qs}` : '/people')
}

export function getPersonById(id: number): Promise<Person> {
  return apiFetch<Person>(`/people/${id}`)
}

export function createPerson(data: PersonCreate): Promise<Person> {
  return apiFetch<Person>('/people', { method: 'POST', body: JSON.stringify(data) })
}

export function updatePerson(id: number, data: PersonCreate): Promise<Person> {
  return apiFetch<Person>(`/people/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
}

export function deletePerson(id: number): Promise<void> {
  return apiFetch<void>(`/people/${id}`, { method: 'DELETE' })
}

export function sortPeople(ids: number[]): Promise<number[]> {
  return apiFetch<{ ids: number[] }>('/people/sort', {
    method: 'POST',
    body: JSON.stringify({ ids }),
  }).then(r => r.ids)
}

export async function importPeople(file: File): Promise<{ imported: number }> {
  const form = new FormData()
  form.append('file', file)
  const res = await fetch(`${BASE_URL}/people/import`, {
    method: 'POST',
    credentials: 'include',
    body: form,
  })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new ApiError(401, {})
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body)
  }
  return res.json()
}
