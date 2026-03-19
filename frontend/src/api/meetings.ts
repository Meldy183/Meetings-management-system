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
