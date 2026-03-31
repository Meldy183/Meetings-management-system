import { apiFetch, BASE_URL } from './client'
import type { Meeting, MeetingCreate, MeetingList } from './types'

export function getMeetings(limit = 20, offset = 0, status = ''): Promise<MeetingList> {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) })
  if (status) params.set('status', status)
  return apiFetch<MeetingList>(`/meetings?${params}`)
}

export function createMeeting(data: MeetingCreate): Promise<Meeting> {
  return apiFetch<Meeting>('/meetings', { method: 'POST', body: JSON.stringify(data) })
}

export function getMeeting(id: string): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${id}`)
}

export function updateMeeting(id: string, data: { title: string; date: string; place?: string; title_phrase?: string; chairperson_phrase?: string }): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${id}`, { method: 'PATCH', body: JSON.stringify(data) })
}

export function setChairperson(meetingId: string, personId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/chairperson`, {
    method: 'PUT',
    body: JSON.stringify({ person_id: personId }),
  })
}

export function deleteMeeting(id: string): Promise<void> {
  return apiFetch<void>(`/meetings/${id}`, { method: 'DELETE' })
}

export function addMeetingPerson(meetingId: string, personId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/people`, {
    method: 'POST',
    body: JSON.stringify({ person_id: personId }),
  })
}

export function removeMeetingPerson(meetingId: string, personId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/people/${personId}`, { method: 'DELETE' })
}

export function sortMeetingPeople(meetingId: string): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/people/sort`, { method: 'POST' })
}

export function reorderPeople(meetingId: string, personIds: number[]): Promise<void> {
  return apiFetch<void>(`/meetings/${meetingId}/people/order`, {
    method: 'PUT',
    body: JSON.stringify({ person_ids: personIds }),
  })
}

export function addAgendaItem(meetingId: string, data: { text: string; speaker_ids: number[] }): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda-items`, { method: 'POST', body: JSON.stringify(data) })
}

export function updateAgendaItem(meetingId: string, itemId: number, data: { text: string; speaker_ids: number[] }): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda-items/${itemId}`, { method: 'PUT', body: JSON.stringify(data) })
}

export function deleteAgendaItem(meetingId: string, itemId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda-items/${itemId}`, { method: 'DELETE' })
}

export function reorderAgendaItems(meetingId: string, agendaItemIds: number[]): Promise<void> {
  return apiFetch<void>(`/meetings/${meetingId}/agenda-items/order`, {
    method: 'PUT',
    body: JSON.stringify({ agenda_item_ids: agendaItemIds }),
  })
}

export function addAgendaItemSpeaker(meetingId: string, itemId: number, personId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda-items/${itemId}/speakers`, {
    method: 'POST',
    body: JSON.stringify({ person_id: personId }),
  })
}

export function removeAgendaItemSpeaker(meetingId: string, itemId: number, pid: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda-items/${itemId}/speakers/${pid}`, { method: 'DELETE' })
}

export function reorderAgendaItemSpeakers(meetingId: string, itemId: number, personIds: number[]): Promise<void> {
  return apiFetch<void>(`/meetings/${meetingId}/agenda-items/${itemId}/speakers/order`, {
    method: 'PUT',
    body: JSON.stringify({ person_ids: personIds }),
  })
}

export function downloadAgenda(id: string): void {
  window.location.href = `${BASE_URL}/meetings/${id}/export/agenda`
}

export function downloadParticipants(id: string): void {
  window.location.href = `${BASE_URL}/meetings/${id}/export/participants`
}
