import { useEffect, useRef, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getPeople, createPerson, updatePerson, deletePerson, importPeople } from '../api/people'
import { ApiError } from '../api/client'
import { ParticipantForm } from '../components/ParticipantForm'
import type { Person, PersonCreate } from '../api/types'

export function ParticipantsPage() {
  const queryClient = useQueryClient()
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [order, setOrder] = useState<'alpha' | 'id'>('alpha')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [duplicateWarning, setDuplicateWarning] = useState(false)
  const [importResult, setImportResult] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(t)
  }, [query])

  const { data: all = [], isLoading, isError } = useQuery({
    queryKey: ['people', debouncedQuery, order],
    queryFn: () => getPeople(debouncedQuery || undefined, debouncedQuery ? undefined : order),
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
      } else {
        alert('Ошибка при удалении участника')
      }
    },
  })

  const createMutation = useMutation({
    mutationFn: (data: PersonCreate) => createPerson(data),
    onSuccess: () => {
      setShowAddForm(false)
      setDuplicateWarning(false)
      queryClient.invalidateQueries({ queryKey: ['people'] })
    },
  })

  const importMutation = useMutation({
    mutationFn: (file: File) => importPeople(file),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['people'] })
      setImportResult(`Импортировано: ${result.imported}`)
      setTimeout(() => setImportResult(null), 4000)
    },
    onError: (e) => {
      const msg = e instanceof ApiError && (e.body as any)?.message
        ? (e.body as any).message
        : 'Ошибка импорта'
      setImportResult(`Ошибка: ${msg}`)
      setTimeout(() => setImportResult(null), 4000)
    },
  })

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    importMutation.mutate(file)
    e.target.value = ''
  }

  function handleCreateSubmit(data: PersonCreate) {
    const exists = all.some(p =>
      p.last_name.toLowerCase() === data.last_name.trim().toLowerCase() &&
      p.first_name.toLowerCase() === data.first_name.trim().toLowerCase() &&
      (p.middle_name ?? '').toLowerCase() === (data.middle_name ?? '').trim().toLowerCase()
    )
    setDuplicateWarning(exists)
    createMutation.mutate(data)
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-gray-900">Реестр</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={importMutation.isPending}
            className="text-xs px-3 py-1.5 rounded-lg border border-gray-300 text-gray-600 bg-white hover:bg-gray-50 disabled:opacity-50"
          >
            {importMutation.isPending ? 'Импорт...' : '↑ Импорт Excel'}
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".xlsx,.xls"
            className="hidden"
            onChange={handleFileChange}
          />
          <button
            onClick={() => setOrder(o => o === 'alpha' ? 'id' : 'alpha')}
            className="text-xs px-3 py-1.5 rounded-lg border border-gray-300 text-gray-600 bg-white hover:bg-gray-50"
          >
            {order === 'alpha' ? 'По алфавиту' : 'По порядку'}
          </button>
        </div>
      </div>

      {importResult && (
        <div className={[
          'text-xs px-3 py-2 rounded-lg border',
          importResult.startsWith('Ошибка')
            ? 'bg-red-50 border-red-200 text-red-700'
            : 'bg-green-50 border-green-200 text-green-700',
        ].join(' ')}>
          {importResult}
        </div>
      )}

      <input
        value={query}
        onChange={e => setQuery(e.target.value)}
        placeholder="Поиск по имени..."
        className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
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
                      className="text-xs text-gray-500 hover:text-green-600 px-2 py-1 rounded"
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
          className="w-full border-2 border-dashed border-gray-300 rounded-lg py-3 text-sm text-gray-500 hover:border-green-400 hover:text-green-500"
        >
          + Добавить нового участника
        </button>
      ) : (
        <div className="bg-white border rounded-lg p-4">
          <p className="text-sm font-medium text-gray-700 mb-3">Новый участник</p>
          {duplicateWarning && (
            <p className="text-xs text-yellow-700 bg-yellow-50 border border-yellow-200 rounded px-3 py-2 mb-3">
              Участник с таким ФИО уже есть в базе
            </p>
          )}
          <ParticipantForm
            onSubmit={handleCreateSubmit}
            onCancel={() => { setShowAddForm(false); setDuplicateWarning(false) }}
            submitLabel="Добавить"
            isLoading={createMutation.isPending}
          />
        </div>
      )}
    </div>
  )
}
