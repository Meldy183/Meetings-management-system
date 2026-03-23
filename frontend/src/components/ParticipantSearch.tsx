import { useEffect, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getPeople, createPerson } from '../api/people'
import { ApiError } from '../api/client'
import { ParticipantForm } from './ParticipantForm'
import type { Person, PersonCreate } from '../api/types'

interface Props {
  onAdd: (person: Person) => void
  existingIds: number[]
}

export function ParticipantSearch({ onAdd, existingIds }: Props) {
  const queryClient = useQueryClient()
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(t)
  }, [query])

  const hasQuery = query.trim().length > 0

  const { data: results = [], isFetching } = useQuery({
    queryKey: ['people', 'search', debouncedQuery],
    queryFn: () => getPeople(debouncedQuery),
    enabled: debouncedQuery.trim().length > 0,
  })

  const noResults = debouncedQuery.trim().length > 0 && !isFetching && results.length === 0

  async function onCreate(data: PersonCreate) {
    setCreating(true)
    setCreateError(null)
    try {
      const p = await createPerson(data)
      queryClient.invalidateQueries({ queryKey: ['people'] })
      onAdd(p)
      setShowCreateForm(false)
      setQuery('')
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        setCreateError('Участник с таким именем уже существует')
      } else {
        setCreateError('Ошибка создания')
      }
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="space-y-2">
      <input
        value={query}
        onChange={e => { setQuery(e.target.value); setShowCreateForm(false); setCreateError(null) }}
        placeholder="Поиск по имени..."
        className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
      />

      {/* Results */}
      {hasQuery && !showCreateForm && results.length > 0 && (
        <div className="border rounded-lg divide-y max-h-48 overflow-y-auto">
          {results.map(p => {
            const added = existingIds.includes(p.id)
            return (
              <div
                key={p.id}
                onClick={() => { if (!added) { onAdd(p); setQuery('') } }}
                className={`flex items-center justify-between px-3 py-2 bg-white ${added ? 'opacity-50' : 'hover:bg-gray-50 cursor-pointer'}`}
              >
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">
                    {p.last_name} {p.first_name} {p.middle_name ?? ''}
                  </p>
                  {p.info && <p className="text-xs text-gray-500 truncate">{p.info}</p>}
                </div>
                {added && <span className="text-xs text-gray-400 shrink-0 ml-3">Уже добавлен</span>}
              </div>
            )
          })}
        </div>
      )}

      {/* Not found */}
      {noResults && !showCreateForm && (
        <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
          <p className="text-sm text-yellow-800">Никого не найдено.</p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="mt-1 text-sm text-green-600 hover:underline"
          >
            Добавить в базу данных
          </button>
        </div>
      )}

      {/* Create form */}
      {showCreateForm && (
        <div className="p-4 border rounded-lg bg-gray-50">
          <p className="text-sm font-medium text-gray-700 mb-3">Новый участник</p>
          {createError && <p className="text-sm text-red-500 mb-2">{createError}</p>}
          <ParticipantForm
            onSubmit={onCreate}
            onCancel={() => { setShowCreateForm(false); setCreateError(null) }}
            submitLabel="Создать и добавить"
            isLoading={creating}
          />
        </div>
      )}

      {/* Always-visible add button when no query */}
      {!hasQuery && !showCreateForm && (
        <button
          onClick={() => setShowCreateForm(true)}
          className="text-sm text-green-600 hover:underline"
        >
          + Новый участник
        </button>
      )}
    </div>
  )
}
