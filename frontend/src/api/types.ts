export interface Person {
  id: number
  last_name: string
  first_name: string
  middle_name?: string
  info?: string
}

export interface PersonCreate {
  last_name: string
  first_name: string
  middle_name?: string
  info?: string
}

export interface AgendaItem {
  id: number
  text: string
  speakers: Person[]
}

export interface MeetingCreate {
  title: string
  date: string // ISO 8601
  place?: string
}

export interface MeetingSummary {
  id: string
  title: string
  date: string
  place?: string
  chairperson: Person | null
  status: string
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
  place?: string
  chairperson: Person | null
  agenda_items: AgendaItem[]
  people: Person[]
  status: string
  created_at: string
}

export interface ApiError {
  message: string
  details?: Record<string, unknown>
}
