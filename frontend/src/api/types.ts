export interface Participant {
  id: number
  last_name: string
  first_name: string
  middle_name?: string
  info?: string
}

export interface ParticipantCreate {
  last_name: string
  first_name: string
  middle_name?: string
  info?: string
}

export interface AgendaItemCreate {
  text: string
  speaker_id: number
}

export interface AgendaItem {
  id: number
  text: string
  speaker: Participant
}

export interface MeetingCreate {
  title: string
  date: string // ISO 8601
  chairperson_id: number
  agenda_items: AgendaItemCreate[]
  participant_ids: number[]
}

export interface MeetingSummary {
  id: string
  title: string
  date: string
  chairperson: Participant
  created_at: string
}

export interface MeetingList {
  total: number
  limit: number
  offset: number
  items: MeetingSummary[]
}

export interface Meeting {
  id: string
  title: string
  date: string
  chairperson: Participant
  agenda_items: AgendaItem[]
  participants: Participant[]
  created_at: string
}

export interface ApiError {
  message: string
  details?: Record<string, unknown>
}
