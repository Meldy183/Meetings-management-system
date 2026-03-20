import { useRef, useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMeeting, downloadAgenda, downloadParticipants, reorderParticipants } from '../api/meetings'
import type { Participant } from '../api/types'

export function MeetingDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [downloading, setDownloading] = useState<'agenda' | 'participants' | null>(null)

  const { data: meeting, isLoading, isError } = useQuery({
    queryKey: ['meeting', id],
    queryFn: () => getMeeting(id!),
    enabled: !!id,
  })

  // Local ordered copy of participants for optimistic DnD updates
  const [participants, setParticipants] = useState<Participant[]>([])
  useEffect(() => {
    if (meeting) setParticipants(meeting.participants)
  }, [meeting])

  const dragIndex = useRef<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  const reorderMutation = useMutation({
    mutationFn: (ids: number[]) => reorderParticipants(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => {
      // Revert to server state on failure
      if (meeting) setParticipants(meeting.participants)
    },
  })

  function handleDragStart(index: number) {
    dragIndex.current = index
  }

  function handleDragOver(e: React.DragEvent, index: number) {
    e.preventDefault()
    setDragOverIndex(index)
  }

  function handleDrop(index: number) {
    const from = dragIndex.current
    if (from === null || from === index) {
      dragIndex.current = null
      setDragOverIndex(null)
      return
    }

    const reordered = [...participants]
    const [moved] = reordered.splice(from, 1)
    reordered.splice(index, 0, moved)

    setParticipants(reordered)
    dragIndex.current = null
    setDragOverIndex(null)

    reorderMutation.mutate(reordered.map(p => p.id))
  }

  function handleDragEnd() {
    dragIndex.current = null
    setDragOverIndex(null)
  }

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

      <div>
        <h2 className="text-sm font-semibold text-gray-700 mb-2">Повестка</h2>
        <div className="space-y-2">
          {meeting.agenda_items.map((item, i) => (
            <div key={i} className="bg-white border rounded-lg p-3">
              <p className="text-sm font-medium">{i + 1}. {item.text}</p>
              <p className="text-xs text-gray-500 mt-1">
                Докладчик: {item.speaker.last_name} {item.speaker.first_name} {item.speaker.middle_name ?? ''}
              </p>
            </div>
          ))}
        </div>
      </div>

      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">
            Участники ({participants.length})
          </h2>
          {reorderMutation.isPending && (
            <span className="text-xs text-gray-400">Сохранение...</span>
          )}
          {reorderMutation.isError && (
            <span className="text-xs text-red-500">Ошибка сохранения</span>
          )}
        </div>
        <div className="space-y-1">
          {participants.map((p, i) => (
            <div
              key={p.id}
              draggable
              onDragStart={() => handleDragStart(i)}
              onDragOver={(e) => handleDragOver(e, i)}
              onDrop={() => handleDrop(i)}
              onDragEnd={handleDragEnd}
              className={[
                'bg-white border rounded-lg p-3 flex items-center gap-3 transition-opacity',
                dragOverIndex === i && dragIndex.current !== i
                  ? 'border-blue-400 bg-blue-50'
                  : '',
                dragIndex.current === i ? 'opacity-40' : 'opacity-100',
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
