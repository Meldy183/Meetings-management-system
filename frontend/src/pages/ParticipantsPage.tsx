import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { searchParticipant, createParticipant, updateParticipant, deleteParticipant } from '../api/participants'
import { ApiError } from '../api/client'
import { ParticipantCard } from '../components/ParticipantCard'
import { ParticipantForm } from '../components/ParticipantForm'
import type { Participant, ParticipantCreate } from '../api/types'
import { Link } from 'react-router-dom'

export function ParticipantsPage() {
  const [results, setResults] = useState<Participant[]>([])
  const [searching, setSearching] = useState(false)
  const [searchError, setSearchError] = useState<string | null>(null)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [lastName, setLastName] = useState('')
  const [firstName, setFirstName] = useState('')
  const [middleName, setMiddleName] = useState('')

  const queryClient = useQueryClient()

  async function handleSearch() {
    if (!lastName || !firstName) return
    setSearching(true)
    setSearchError(null)
    try {
      const p = await searchParticipant(lastName, firstName, middleName || undefined)
      setResults([p])
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        setResults([])
        setSearchError('Участник не найден')
      } else {
        setSearchError('Ошибка поиска')
      }
    } finally {
      setSearching(false)
    }
  }

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: ParticipantCreate }) => updateParticipant(id, data),
    onSuccess: (updated) => {
      setResults(r => r.map(p => p.id === updated.id ? updated : p))
      setEditingId(null)
      queryClient.invalidateQueries({ queryKey: ['meetings'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteParticipant(id),
    onSuccess: (_, id) => {
      setResults(r => r.filter(p => p.id !== id))
      queryClient.invalidateQueries({ queryKey: ['meetings'] })
    },
    onError: (e) => {
      if (e instanceof ApiError && e.status === 409) {
        alert('Нельзя удалить: участник привязан к существующим совещаниям')
      }
    },
  })

  const createMutation = useMutation({
    mutationFn: (data: ParticipantCreate) => createParticipant(data),
    onSuccess: (p) => {
      setResults(r => [...r, p])
      setShowAddForm(false)
    },
    onError: (e) => {
      if (e instanceof ApiError && e.status === 409) {
        alert('Участник с таким именем уже существует')
      }
    },
  })

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/" className="text-gray-400 hover:text-gray-600">←</Link>
        <h1 className="text-xl font-semibold text-gray-900">Участники</h1>
      </div>

      {/* Search */}
      <div className="bg-white border rounded-lg p-4 space-y-3">
        <p className="text-sm font-medium text-gray-700">Поиск по имени</p>
        <div className="grid grid-cols-3 gap-2">
          <input value={lastName} onChange={e => setLastName(e.target.value)} placeholder="Фамилия *"
            className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input value={firstName} onChange={e => setFirstName(e.target.value)} placeholder="Имя *"
            className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input value={middleName} onChange={e => setMiddleName(e.target.value)} placeholder="Отчество"
            className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <button onClick={handleSearch} disabled={searching || !lastName || !firstName}
          className="bg-gray-100 border text-gray-700 px-4 py-2 rounded-lg text-sm hover:bg-gray-200 disabled:opacity-50">
          {searching ? 'Поиск...' : 'Найти'}
        </button>
        {searchError && <p className="text-sm text-red-500">{searchError}</p>}
      </div>

      {/* Results */}
      {results.length > 0 && (
        <div className="space-y-2">
          {results.map(p => (
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
                <ParticipantCard
                  participant={p}
                  onEdit={() => setEditingId(p.id)}
                  onRemove={() => { if (confirm('Удалить участника?')) deleteMutation.mutate(p.id) }}
                />
              )}
            </div>
          ))}
        </div>
      )}

      {/* Add new */}
      <div>
        {!showAddForm ? (
          <button onClick={() => setShowAddForm(true)}
            className="w-full border-2 border-dashed border-gray-300 rounded-lg py-3 text-sm text-gray-500 hover:border-blue-400 hover:text-blue-500">
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
    </div>
  )
}
