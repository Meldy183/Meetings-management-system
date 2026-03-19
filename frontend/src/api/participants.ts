import { apiFetch } from './client'
import type { Participant, ParticipantCreate } from './types'

export function searchParticipant(
  lastName: string,
  firstName: string,
  middleName?: string
): Promise<Participant> {
  const params = new URLSearchParams({ last_name: lastName, first_name: firstName })
  if (middleName) params.set('middle_name', middleName)
  return apiFetch<Participant>(`/participants?${params}`)
}

export function createParticipant(data: ParticipantCreate): Promise<Participant> {
  return apiFetch<Participant>('/participants', { method: 'POST', body: JSON.stringify(data) })
}

export function updateParticipant(id: number, data: ParticipantCreate): Promise<Participant> {
  return apiFetch<Participant>(`/participants/${id}`, { method: 'PUT', body: JSON.stringify(data) })
}

export function deleteParticipant(id: number): Promise<void> {
  return apiFetch<void>(`/participants/${id}`, { method: 'DELETE' })
}
