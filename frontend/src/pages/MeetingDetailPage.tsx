import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getMeeting, downloadAgenda, downloadParticipants } from '../api/meetings'

export function MeetingDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [downloading, setDownloading] = useState<'agenda' | 'participants' | null>(null)

  const { data: meeting, isLoading, isError } = useQuery({
    queryKey: ['meeting', id],
    queryFn: () => getMeeting(id!),
    enabled: !!id,
  })

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })
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
        <h2 className="text-sm font-semibold text-gray-700 mb-2">
          Участники ({meeting.participants.length})
        </h2>
        <div className="space-y-2">
          {meeting.participants.map(p => (
            <div key={p.id} className="bg-white border rounded-lg p-3">
              <p className="text-sm font-medium">
                {p.last_name} {p.first_name} {p.middle_name ?? ''}
              </p>
              {p.info && <p className="text-xs text-gray-500 mt-0.5">{p.info}</p>}
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
