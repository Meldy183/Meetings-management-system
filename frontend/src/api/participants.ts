import { apiFetch } from './client'
import type { Participant, ParticipantCreate } from './types'

export function getParticipants(q?: string): Promise<Participant[]> {
  const url = q ? `/participants?q=${encodeURIComponent(q)}` : '/participants'
  return apiFetch<Participant[]>(url)
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
