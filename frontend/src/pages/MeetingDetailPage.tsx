import { useRef, useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMeeting, downloadAgenda, downloadParticipants, reorderParticipants, reorderAgendaItems } from '../api/meetings'
import type { Participant, AgendaItem } from '../api/types'

function useDragReorder<T>(
  items: T[],
  onDrop: (reordered: T[]) => void,
) {
  const dragIndex = useRef<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  function handleDragStart(i: number) { dragIndex.current = i }

  function handleDragOver(e: React.DragEvent, i: number) {
    e.preventDefault()
    setDragOverIndex(i)
  }

  function handleDrop(i: number) {
    const from = dragIndex.current
    dragIndex.current = null
    setDragOverIndex(null)
    if (from === null || from === i) return
    const next = [...items]
    const [moved] = next.splice(from, 1)
    next.splice(i, 0, moved)
    onDrop(next)
  }

  function handleDragEnd() {
    dragIndex.current = null
    setDragOverIndex(null)
  }

  return { dragIndex, dragOverIndex, handleDragStart, handleDragOver, handleDrop, handleDragEnd }
}

export function MeetingDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [downloading, setDownloading] = useState<'agenda' | 'participants' | null>(null)

  const { data: meeting, isLoading, isError } = useQuery({
    queryKey: ['meeting', id],
    queryFn: () => getMeeting(id!),
    enabled: !!id,
  })

  const [agendaItems, setAgendaItems] = useState<AgendaItem[]>([])
  const [participants, setParticipants] = useState<Participant[]>([])

  useEffect(() => {
    if (meeting) {
      setAgendaItems(meeting.agenda_items)
      setParticipants(meeting.participants)
    }
  }, [meeting])

  const agendaMutation = useMutation({
    mutationFn: (ids: number[]) => reorderAgendaItems(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => { if (meeting) setAgendaItems(meeting.agenda_items) },
  })

  const participantsMutation = useMutation({
    mutationFn: (ids: number[]) => reorderParticipants(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => { if (meeting) setParticipants(meeting.participants) },
  })

  const agendaDnd = useDragReorder(agendaItems, (reordered) => {
    setAgendaItems(reordered)
    agendaMutation.mutate(reordered.map(i => i.id))
  })

  const participantsDnd = useDragReorder(participants, (reordered) => {
    setParticipants(reordered)
    participantsMutation.mutate(reordered.map(p => p.id))
  })

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', year: 'numeric',
      hour: '2-digit', minute: '2-digit',
    })
  }

  async function handleDownload(type: 'agenda' | 'participants') {
    if (!id) return
    setDownloading(type)
    try {
      if (type === 'agenda') await downloadAgenda(id)
      else await downloadParticipants(id)
    } finally {
      setDownloading(null)
    }
  }

  if (isLoading) return <div className="max-w-2xl mx-auto px-4 py-6 text-gray-500 text-sm">Загрузка...</div>
  if (isError || !meeting) return <div className="max-w-2xl mx-auto px-4 py-6 text-red-500 text-sm">Совещание не найдено</div>

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">
      <div className="flex items-start gap-3">
        <Link to="/" className="text-gray-400 hover:text-gray-600 mt-1">←</Link>
        <h1 className="text-lg font-semibold text-gray-900 leading-snug">{meeting.title}</h1>
      </div>

      <div className="bg-white border rounded-lg p-4 space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Дата</span>
          <span className="font-medium">{formatDate(meeting.date)}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Председательствующий</span>
          <span className="font-medium text-right">
            {meeting.chairperson.last_name} {meeting.chairperson.first_name} {meeting.chairperson.middle_name ?? ''}
          </span>
        </div>
      </div>

      {/* Agenda items */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">
            Повестка ({agendaItems.length})
          </h2>
          {agendaMutation.isPending && <span className="text-xs text-gray-400">Сохранение...</span>}
          {agendaMutation.isError && <span className="text-xs text-red-500">Ошибка сохранения</span>}
        </div>
        <div className="space-y-1">
          {agendaItems.map((item, i) => (
            <div
              key={item.id}
              draggable
              onDragStart={() => agendaDnd.handleDragStart(i)}
              onDragOver={(e) => agendaDnd.handleDragOver(e, i)}
              onDrop={() => agendaDnd.handleDrop(i)}
              onDragEnd={agendaDnd.handleDragEnd}
              className={[
                'bg-white border rounded-lg p-3 flex items-start gap-3 transition-opacity',
                agendaDnd.dragOverIndex === i && agendaDnd.dragIndex.current !== i
                  ? 'border-blue-400 bg-blue-50'
                  : '',
                agendaDnd.dragIndex.current === i ? 'opacity-40' : 'opacity-100',
              ].join(' ')}
            >
              <span
                className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none text-lg leading-none mt-0.5"
                title="Перетащить для изменения порядка"
              >
                ⠿
              </span>
              <span className="text-xs text-gray-400 w-5 shrink-0 mt-0.5">{i + 1}.</span>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium">{item.text}</p>
                <p className="text-xs text-gray-500 mt-1">
                  Докладчик: {item.speaker.last_name} {item.speaker.first_name} {item.speaker.middle_name ?? ''}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Participants */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">
            Участники ({participants.length})
          </h2>
          {participantsMutation.isPending && <span className="text-xs text-gray-400">Сохранение...</span>}
          {participantsMutation.isError && <span className="text-xs text-red-500">Ошибка сохранения</span>}
        </div>
        <div className="space-y-1">
          {participants.map((p, i) => (
            <div
              key={p.id}
              draggable
              onDragStart={() => participantsDnd.handleDragStart(i)}
              onDragOver={(e) => participantsDnd.handleDragOver(e, i)}
              onDrop={() => participantsDnd.handleDrop(i)}
              onDragEnd={participantsDnd.handleDragEnd}
              className={[
                'bg-white border rounded-lg p-3 flex items-center gap-3 transition-opacity',
                participantsDnd.dragOverIndex === i && participantsDnd.dragIndex.current !== i
                  ? 'border-blue-400 bg-blue-50'
                  : '',
                participantsDnd.dragIndex.current === i ? 'opacity-40' : 'opacity-100',
              ].join(' ')}
            >
              <span
                className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none text-lg leading-none"
                title="Перетащить для изменения порядка"
              >
                ⠿
              </span>
              <span className="text-xs text-gray-400 w-5 shrink-0">{i + 1}.</span>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">
                  {p.last_name} {p.first_name} {p.middle_name ?? ''}
                </p>
                {p.info && <p className="text-xs text-gray-500 mt-0.5 truncate">{p.info}</p>}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex gap-3 pt-2">
        <button
          onClick={() => handleDownload('agenda')}
          disabled={downloading === 'agenda'}
          className="flex-1 bg-white border border-gray-300 text-gray-700 px-4 py-3 rounded-lg text-sm font-medium hover:bg-gray-50 disabled:opacity-50"
        >
          {downloading === 'agenda' ? 'Загрузка...' : '↓ Повестка (.docx)'}
        </button>
        <button
          onClick={() => handleDownload('participants')}
          disabled={downloading === 'participants'}
          className="flex-1 bg-white border border-gray-300 text-gray-700 px-4 py-3 rounded-lg text-sm font-medium hover:bg-gray-50 disabled:opacity-50"
        >
          {downloading === 'participants' ? 'Загрузка...' : '↓ Список участников (.docx)'}
        </button>
      </div>
    </div>
  )
}
