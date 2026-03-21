import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getPeople, createPerson, updatePerson, deletePerson } from '../api/people'
import { ApiError } from '../api/client'
import { ParticipantForm } from '../components/ParticipantForm'
import type { Person, PersonCreate } from '../api/types'

export function ParticipantsPage() {
  const queryClient = useQueryClient()
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(t)
  }, [query])

  const { data: all = [], isLoading, isError } = useQuery({
    queryKey: ['people', debouncedQuery],
    queryFn: () => getPeople(debouncedQuery || undefined),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: PersonCreate }) => updatePerson(id, data),
    onSuccess: () => {
      setEditingId(null)
      queryClient.invalidateQueries({ queryKey: ['people'] })
      queryClient.invalidateQueries({ queryKey: ['meetings'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deletePerson(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['people'] })
      queryClient.invalidateQueries({ queryKey: ['meetings'] })
    },
    onError: (e) => {
      if (e instanceof ApiError && e.status === 409) {
        alert('Нельзя удалить: участник привязан к существующим совещаниям')
      }
    },
  })

  const createMutation = useMutation({
    mutationFn: (data: PersonCreate) => createPerson(data),
    onSuccess: () => {
      setShowAddForm(false)
      queryClient.invalidateQueries({ queryKey: ['people'] })
    },
    onError: (e) => {
      if (e instanceof ApiError && e.status === 409) {
        alert('Участник с таким именем уже существует')
      }
    },
  })

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-4">
      <div className="flex items-center gap-3">
        <Link to="/" className="text-gray-400 hover:text-gray-600">←</Link>
        <h1 className="text-xl font-semibold text-gray-900">Участники</h1>
      </div>

      <input
        value={query}
        onChange={e => setQuery(e.target.value)}
        placeholder="Поиск по имени..."
        className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />

      {isLoading && <p className="text-sm text-gray-400">Загрузка...</p>}
      {isError && <p className="text-sm text-red-500">Ошибка загрузки</p>}

      {!isLoading && !isError && (
        <div className="space-y-2">
          {all.length === 0 && (
            <p className="text-sm text-gray-400 py-4 text-center">
              {query ? 'Никого не найдено' : 'Список пуст'}
            </p>
          )}
          {all.map((p: Person) => (
            <div key={p.id}>
              {editingId === p.id ? (
                <div className="p-4 border rounded-lg bg-gray-50">
                  <ParticipantForm
                    defaultValues={{ last_name: p.last_name, first_name: p.first_name, middle_name: p.middle_name, info: p.info }}
                    onSubmit={(data) => updateMutation.mutate({ id: p.id, data })}
                    onCancel={() => setEditingId(null)}
                    isLoading={updateMutation.isPending}
                  />
                </div>
              ) : (
                <div className="flex items-center justify-between p-3 bg-white border rounded-lg">
                  <div className="min-w-0">
                    <p className="text-sm font-medium truncate">
                      {p.last_name} {p.first_name} {p.middle_name ?? ''}
                    </p>
                    {p.info && <p className="text-xs text-gray-500 mt-0.5 truncate">{p.info}</p>}
                  </div>
                  <div className="flex gap-2 shrink-0 ml-3">
                    <button
                      onClick={() => setEditingId(p.id)}
                      className="text-xs text-gray-500 hover:text-blue-600 px-2 py-1 rounded"
                    >
                      Изменить
                    </button>
                    <button
                      onClick={() => { if (confirm('Удалить участника?')) deleteMutation.mutate(p.id) }}
                      className="text-xs text-gray-500 hover:text-red-600 px-2 py-1 rounded"
                    >
                      Удалить
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {!showAddForm ? (
        <button
          onClick={() => setShowAddForm(true)}
          className="w-full border-2 border-dashed border-gray-300 rounded-lg py-3 text-sm text-gray-500 hover:border-blue-400 hover:text-blue-500"
        >
          + Добавить нового участника
        </button>
      ) : (
        <div className="bg-white border rounded-lg p-4">
          <p className="text-sm font-medium text-gray-700 mb-3">Новый участник</p>
          <ParticipantForm
            onSubmit={(data) => createMutation.mutate(data)}
            onCancel={() => setShowAddForm(false)}
            submitLabel="Добавить"
            isLoading={createMutation.isPending}
          />
        </div>
      )}
    </div>
  )
}
