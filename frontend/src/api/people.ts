import { apiFetch } from './client'
import type { Person, PersonCreate } from './types'

export function getPeople(q?: string): Promise<Person[]> {
  const url = q ? `/people?q=${encodeURIComponent(q)}` : '/people'
  return apiFetch<Person[]>(url)
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
