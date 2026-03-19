import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { searchParticipant, createParticipant } from '../api/participants'
import { ApiError } from '../api/client'
import { ParticipantForm } from './ParticipantForm'
import type { Participant, ParticipantCreate } from '../api/types'

interface SearchFields {
  last_name: string
  first_name: string
  middle_name: string
}

interface Props {
  onAdd: (participant: Participant) => void
  existingIds: number[]
}

export function ParticipantSearch({ onAdd, existingIds }: Props) {
  const [found, setFound] = useState<Participant | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [searching, setSearching] = useState(false)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { register, handleSubmit, getValues } = useForm<SearchFields>()

  async function onSearch(data: SearchFields) {
    setSearching(true)
    setFound(null)
    setNotFound(false)
    setShowCreateForm(false)
    setError(null)
    try {
      const participant = await searchParticipant(data.last_name, data.first_name, data.middle_name || undefined)
      setFound(participant)
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        setNotFound(true)
      } else {
        setError('Ошибка поиска')
      }
    } finally {
      setSearching(false)
    }
  }

  async function onCreate(data: ParticipantCreate) {
    setCreating(true)
    try {
      const participant = await createParticipant(data)
      onAdd(participant)
      setShowCreateForm(false)
      setNotFound(false)
      setFound(null)
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        setError('Участник с таким именем уже существует')
      } else {
        setError('Ошибка создания')
      }
    } finally {
      setCreating(false)
    }
  }

  const alreadyAdded = found ? existingIds.includes(found.id) : false
  const values = getValues()
  const prefillData: Partial<ParticipantCreate> = {
    last_name: values.last_name,
    first_name: values.first_name,
    middle_name: values.middle_name,
  }

  return (
    <div className="space-y-3">
      <form onSubmit={handleSubmit(onSearch)} className="space-y-2">
        <div className="grid grid-cols-3 gap-2">
          <div>
            <input
              {...register('last_name', { required: true })}
              placeholder="Фамилия *"
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <input
              {...register('first_name', { required: true })}
              placeholder="Имя *"
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <input
              {...register('middle_name')}
              placeholder="Отчество"
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="submit"
            disabled={searching}
            className="bg-gray-100 border text-gray-700 px-4 py-2 rounded-lg text-sm hover:bg-gray-200 disabled:opacity-50"
          >
            {searching ? 'Поиск...' : 'Найти'}
          </button>
          <button
            type="button"
            onClick={() => { setShowCreateForm(v => !v); setFound(null); setNotFound(false); setError(null) }}
            className="text-blue-600 text-sm px-2 py-2 hover:underline"
          >
            + Новый участник
          </button>
        </div>
      </form>

      {error && <p className="text-sm text-red-500">{error}</p>}

      {found && (
        <div className="p-3 bg-green-50 border border-green-200 rounded-lg flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">
              {found.last_name} {found.first_name} {found.middle_name}
            </p>
            {found.info && <p className="text-xs text-gray-500">{found.info}</p>}
          </div>
          {alreadyAdded ? (
            <span className="text-xs text-gray-500">Уже добавлен</span>
          ) : (
            <button
              onClick={() => { onAdd(found); setFound(null) }}
              className="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-xs font-medium hover:bg-blue-700"
            >
              Добавить
            </button>
          )}
        </div>
      )}

      {notFound && !showCreateForm && (
        <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
          <p className="text-sm text-yellow-800">Участник не найден.</p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="mt-2 text-sm text-blue-600 hover:underline"
          >
            Добавить в базу данных
          </button>
        </div>
      )}

      {showCreateForm && (
        <div className="p-4 border rounded-lg bg-gray-50">
          <p className="text-sm font-medium text-gray-700 mb-3">Новый участник</p>
          <ParticipantForm
            defaultValues={prefillData}
            onSubmit={onCreate}
            onCancel={() => setShowCreateForm(false)}
            submitLabel="Создать и добавить"
            isLoading={creating}
          />
        </div>
      )}
    </div>
  )
}
