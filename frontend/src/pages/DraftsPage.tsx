import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getMeetings, deleteMeeting } from '../api/meetings'

export function DraftsPage() {
  const [offset, setOffset] = useState(0)
  const limit = 20
  const qc = useQueryClient()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['drafts', offset],
    queryFn: () => getMeetings(limit, offset, 'incomplete'),
  })

  const del = useMutation({
    mutationFn: (id: string) => deleteMeeting(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['drafts'] }),
  })

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-6">
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Черновики</h1>

      {isLoading && <p className="text-gray-500 text-sm">Загрузка...</p>}
      {isError && <p className="text-red-500 text-sm">Ошибка загрузки</p>}

      {data && (
        <>
          <div className="space-y-3">
            {data.items.map(meeting => (
              <div
                key={meeting.id}
                className="flex items-center gap-3 p-4 bg-white border rounded-lg hover:bg-gray-50 hover:border-gray-300 transition-colors"
              >
                <Link
                  to={`/meetings/${meeting.id}`}
                  className="flex-1 min-w-0 flex items-center gap-3"
                >
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm text-gray-900 leading-snug truncate">{meeting.title}</p>
                    <p className="text-xs text-gray-500 mt-1">{formatDate(meeting.date)}</p>
                    {meeting.chairperson ? (
                      <p className="text-xs text-gray-400 mt-0.5">
                        {meeting.chairperson.last_name} {meeting.chairperson.first_name}
                      </p>
                    ) : (
                      <p className="text-xs text-gray-400 mt-0.5 italic">Председатель не назначен</p>
                    )}
                  </div>
                  <span className="shrink-0 text-gray-300 text-lg">›</span>
                </Link>
                <button
                  onClick={() => del.mutate(meeting.id)}
                  disabled={del.isPending}
                  className="shrink-0 text-xs text-red-500 hover:text-red-700 disabled:opacity-40 px-2 py-1 border border-red-200 rounded"
                >
                  Удалить
                </button>
              </div>
            ))}
            {data.items.length === 0 && (
              <p className="text-gray-500 text-sm text-center py-8">Черновиков нет</p>
            )}
          </div>

          <div className="flex justify-between mt-6">
            {offset > 0 && (
              <button
                onClick={() => setOffset(o => Math.max(0, o - limit))}
                className="text-sm text-green-600 hover:underline"
              >
                ← Назад
              </button>
            )}
            {offset + limit < data.total && (
              <button
                onClick={() => setOffset(o => o + limit)}
                className="text-sm text-green-600 hover:underline ml-auto"
              >
                Далее →
              </button>
            )}
          </div>
        </>
      )}
    </div>
  )
}
