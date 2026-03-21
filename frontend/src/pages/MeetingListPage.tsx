import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getMeetings } from '../api/meetings'

export function MeetingListPage() {
  const [offset, setOffset] = useState(0)
  const limit = 20

  const { data, isLoading, isError } = useQuery({
    queryKey: ['meetings', offset],
    queryFn: () => getMeetings(limit, offset),
  })

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Совещания</h1>
        <Link
          to="/meetings/new"
          className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700"
        >
          + Создать
        </Link>
      </div>

      {isLoading && <p className="text-gray-500 text-sm">Загрузка...</p>}
      {isError && <p className="text-red-500 text-sm">Ошибка загрузки</p>}

      {data && (
        <>
          <div className="space-y-3">
            {data.items.map(meeting => (
              <Link
                key={meeting.id}
                to={`/meetings/${meeting.id}`}
                className="block p-4 bg-white border rounded-lg hover:border-blue-400 transition-colors"
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="font-medium text-sm text-gray-900 leading-snug">{meeting.title}</p>
                  {meeting.status === 'incomplete' && (
                    <span className="shrink-0 text-xs bg-yellow-100 text-yellow-700 px-2 py-0.5 rounded-full">
                      Не готово
                    </span>
                  )}
                </div>
                <p className="text-xs text-gray-500 mt-1">{formatDate(meeting.date)}</p>
                {meeting.chairperson ? (
                  <p className="text-xs text-gray-400 mt-0.5">
                    {meeting.chairperson.last_name} {meeting.chairperson.first_name}
                  </p>
                ) : (
                  <p className="text-xs text-gray-400 mt-0.5 italic">Председатель не назначен</p>
                )}
              </Link>
            ))}
            {data.items.length === 0 && (
              <p className="text-gray-500 text-sm text-center py-8">Совещаний пока нет</p>
            )}
          </div>

          <div className="flex justify-between mt-6">
            {offset > 0 && (
              <button
                onClick={() => setOffset(o => Math.max(0, o - limit))}
                className="text-sm text-blue-600 hover:underline"
              >
                ← Назад
              </button>
            )}
            {offset + limit < data.total && (
              <button
                onClick={() => setOffset(o => o + limit)}
                className="text-sm text-blue-600 hover:underline ml-auto"
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
