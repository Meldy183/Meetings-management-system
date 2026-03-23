import { apiFetch, apiFetchBlob } from './client'
import type { Meeting, MeetingCreate, MeetingList } from './types'

export function getMeetings(limit = 20, offset = 0): Promise<MeetingList> {
  return apiFetch<MeetingList>(`/meetings?limit=${limit}&offset=${offset}`)
}

export function createMeeting(data: MeetingCreate): Promise<Meeting> {
  return apiFetch<Meeting>('/meetings', { method: 'POST', body: JSON.stringify(data) })
}

export function getMeeting(id: string): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${id}`)
}

export function updateMeeting(id: string, data: { title: string; date: string; place?: string }): Promise<Meeting> {
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

export async function downloadAgenda(id: string): Promise<void> {
  const { blob, filename } = await apiFetchBlob(`/meetings/${id}/export/agenda`)
  triggerDownload(blob, filename)
}

export async function downloadParticipants(id: string): Promise<void> {
  const { blob, filename } = await apiFetchBlob(`/meetings/${id}/export/participants`)
  triggerDownload(blob, filename)
}

function triggerDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
