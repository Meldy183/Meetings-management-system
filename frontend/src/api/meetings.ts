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

export function updateMeeting(id: string, data: { title: string; date: string; chairperson_id: number }): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${id}`, { method: 'PUT', body: JSON.stringify(data) })
}

export function deleteMeeting(id: string): Promise<void> {
  return apiFetch<void>(`/meetings/${id}`, { method: 'DELETE' })
}

export function addMeetingParticipant(meetingId: string, participantId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/participants`, {
    method: 'POST',
    body: JSON.stringify({ participant_id: participantId }),
  })
}

export function removeMeetingParticipant(meetingId: string, participantId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/participants/${participantId}`, { method: 'DELETE' })
}

export function addAgendaItem(meetingId: string, data: { text: string; speaker_id: number }): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda`, { method: 'POST', body: JSON.stringify(data) })
}

export function updateAgendaItem(meetingId: string, itemId: number, data: { text: string; speaker_id: number }): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda/${itemId}`, { method: 'PUT', body: JSON.stringify(data) })
}

export function deleteAgendaItem(meetingId: string, itemId: number): Promise<Meeting> {
  return apiFetch<Meeting>(`/meetings/${meetingId}/agenda/${itemId}`, { method: 'DELETE' })
}

export function reorderParticipants(meetingId: string, participantIds: number[]): Promise<void> {
  return apiFetch<void>(`/meetings/${meetingId}/participants/order`, {
    method: 'PUT',
    body: JSON.stringify({ participant_ids: participantIds }),
  })
}

export function reorderAgendaItems(meetingId: string, agendaItemIds: number[]): Promise<void> {
  return apiFetch<void>(`/meetings/${meetingId}/agenda/order`, {
    method: 'PUT',
    body: JSON.stringify({ agenda_item_ids: agendaItemIds }),
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
  a.click()
  URL.revokeObjectURL(url)
}
