import { useReducer, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createMeeting } from '../api/meetings'
import { updateParticipant } from '../api/participants'
import { ParticipantSearch } from '../components/ParticipantSearch'
import { ParticipantCard } from '../components/ParticipantCard'
import { ParticipantForm } from '../components/ParticipantForm'
import { StepIndicator } from '../components/StepIndicator'
import { ApiError } from '../api/client'
import type { Participant, ParticipantCreate } from '../api/types'

interface AgendaItem {
  text: string
  speaker_id: number
}

interface WizardState {
  title: string
  date: string
  participants: Participant[]
  chairperson_id: number | null
  agenda_items: AgendaItem[]
}

type WizardAction =
  | { type: 'SET_TITLE_DATE'; title: string; date: string }
  | { type: 'ADD_PARTICIPANT'; participant: Participant }
  | { type: 'REMOVE_PARTICIPANT'; id: number }
  | { type: 'UPDATE_PARTICIPANT'; participant: Participant }
  | { type: 'SET_CHAIRPERSON'; id: number }
  | { type: 'ADD_AGENDA_ITEM' }
  | { type: 'UPDATE_AGENDA_ITEM'; index: number; item: AgendaItem }
  | { type: 'REMOVE_AGENDA_ITEM'; index: number }

function reducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'SET_TITLE_DATE':
      return { ...state, title: action.title, date: action.date }
    case 'ADD_PARTICIPANT':
      if (state.participants.find(p => p.id === action.participant.id)) return state
      return { ...state, participants: [...state.participants, action.participant] }
    case 'REMOVE_PARTICIPANT': {
      const newParticipants = state.participants.filter(p => p.id !== action.id)
      return {
        ...state,
        participants: newParticipants,
        chairperson_id: state.chairperson_id === action.id ? null : state.chairperson_id,
        agenda_items: state.agenda_items.map(item =>
          item.speaker_id === action.id ? { ...item, speaker_id: 0 } : item
        ),
      }
    }
    case 'UPDATE_PARTICIPANT':
      return { ...state, participants: state.participants.map(p => p.id === action.participant.id ? action.participant : p) }
    case 'SET_CHAIRPERSON':
      return { ...state, chairperson_id: action.id }
    case 'ADD_AGENDA_ITEM':
      return { ...state, agenda_items: [...state.agenda_items, { text: '', speaker_id: 0 }] }
    case 'UPDATE_AGENDA_ITEM':
      return { ...state, agenda_items: state.agenda_items.map((item, i) => i === action.index ? action.item : item) }
    case 'REMOVE_AGENDA_ITEM':
      return { ...state, agenda_items: state.agenda_items.filter((_, i) => i !== action.index) }
    default:
      return state
  }
}

const STEP_LABELS = ['Тема', 'Участники', 'Председатель', 'Повестка', 'Подтверждение']

const initialState: WizardState = {
  title: '',
  date: '',
  participants: [],
  chairperson_id: null,
  agenda_items: [],
}

export function CreateMeetingPage() {
  const [step, setStep] = useState(1)
  const [state, dispatch] = useReducer(reducer, initialState)
  const [editingParticipantId, setEditingParticipantId] = useState<number | null>(null)
  const [titleInput, setTitleInput] = useState('')
  const [dateInput, setDateInput] = useState('')
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const updateParticipantMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: ParticipantCreate }) => updateParticipant(id, data),
    onSuccess: (updated) => {
      dispatch({ type: 'UPDATE_PARTICIPANT', participant: updated })
      setEditingParticipantId(null)
    },
  })

  const createMeetingMutation = useMutation({
    mutationFn: () => createMeeting({
      title: state.title,
      date: new Date(state.date).toISOString(),
      chairperson_id: state.chairperson_id!,
      agenda_items: state.agenda_items,
      participant_ids: state.participants.map(p => p.id),
    }),
    onSuccess: (meeting) => {
      queryClient.invalidateQueries({ queryKey: ['meetings'] })
      navigate(`/meetings/${meeting.id}`)
    },
    onError: (e) => {
      if (e instanceof ApiError && e.status === 422) {
        alert('Ошибка: некоторые участники не найдены в базе данных')
      } else {
        alert('Ошибка создания совещания')
      }
    },
  })

  function goNext() { setStep(s => s + 1) }
  function goBack() { setStep(s => s - 1) }

  function fullName(p: Participant) {
    return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ')
  }

  const canProceedStep1 = titleInput.trim() && dateInput
  const canProceedStep2 = state.participants.length > 0
  const canProceedStep3 = state.chairperson_id !== null
  const canProceedStep4 = state.agenda_items.length > 0 &&
    state.agenda_items.every(item => item.text.trim() && item.speaker_id > 0)

  return (
    <div className="max-w-2xl mx-auto px-4 py-6">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate('/')} className="text-gray-400 hover:text-gray-600">←</button>
        <h1 className="text-lg font-semibold text-gray-900">Новое совещание</h1>
      </div>

      <StepIndicator current={step} total={5} labels={STEP_LABELS} />

      {/* Step 1: Title + Date */}
      {step === 1 && (
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Тема совещания *</label>
            <textarea
              value={titleInput}
              onChange={e => setTitleInput(e.target.value)}
              rows={3}
              placeholder="совещания по вопросам..."
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Дата и время *</label>
            <input
              type="datetime-local"
              value={dateInput}
              onChange={e => setDateInput(e.target.value)}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <button
            disabled={!canProceedStep1}
            onClick={() => {
              dispatch({ type: 'SET_TITLE_DATE', title: titleInput.trim(), date: dateInput })
              goNext()
            }}
            className="w-full bg-blue-600 text-white py-3 rounded-lg font-medium text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            Далее →
          </button>
        </div>
      )}

      {/* Step 2: Participants */}
      {step === 2 && (
        <div className="space-y-4">
          <ParticipantSearch
            onAdd={p => dispatch({ type: 'ADD_PARTICIPANT', participant: p })}
            existingIds={state.participants.map(p => p.id)}
          />

          {state.participants.length > 0 && (
            <div>
              <p className="text-sm font-medium text-gray-700 mb-2">
                Список участников ({state.participants.length})
              </p>
              <div className="space-y-2">
                {state.participants.map(p => (
                  <div key={p.id}>
                    {editingParticipantId === p.id ? (
                      <div className="p-4 border rounded-lg bg-gray-50">
                        <ParticipantForm
                          defaultValues={{ last_name: p.last_name, first_name: p.first_name, middle_name: p.middle_name, info: p.info }}
                          onSubmit={data => updateParticipantMutation.mutate({ id: p.id, data })}
                          onCancel={() => setEditingParticipantId(null)}
                          isLoading={updateParticipantMutation.isPending}
                        />
                      </div>
                    ) : (
                      <ParticipantCard
                        participant={p}
                        onEdit={() => setEditingParticipantId(p.id)}
                        onRemove={() => dispatch({ type: 'REMOVE_PARTICIPANT', id: p.id })}
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="flex gap-3">
            <button onClick={goBack} className="flex-1 border text-gray-700 py-3 rounded-lg text-sm hover:bg-gray-50">
              ← Назад
            </button>
            <button
              disabled={!canProceedStep2}
              onClick={goNext}
              className="flex-1 bg-blue-600 text-white py-3 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              Готово ({state.participants.length}) →
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Chairperson */}
      {step === 3 && (
        <div className="space-y-4">
          <p className="text-sm text-gray-600">Выберите председательствующего:</p>
          <div className="space-y-2">
            {state.participants.map(p => (
              <button
                key={p.id}
                onClick={() => dispatch({ type: 'SET_CHAIRPERSON', id: p.id })}
                className={`w-full flex items-center gap-3 p-3 border rounded-lg text-left transition-colors
                  ${state.chairperson_id === p.id ? 'border-blue-500 bg-blue-50' : 'bg-white hover:bg-gray-50'}`}
              >
                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0
                  ${state.chairperson_id === p.id ? 'border-blue-500' : 'border-gray-300'}`}>
                  {state.chairperson_id === p.id && <div className="w-2.5 h-2.5 rounded-full bg-blue-500" />}
                </div>
                <div>
                  <p className="text-sm font-medium">{fullName(p)}</p>
                  {p.info && <p className="text-xs text-gray-500">{p.info}</p>}
                </div>
              </button>
            ))}
          </div>
          <div className="flex gap-3">
            <button onClick={goBack} className="flex-1 border text-gray-700 py-3 rounded-lg text-sm hover:bg-gray-50">
              ← Назад
            </button>
            <button
              disabled={!canProceedStep3}
              onClick={goNext}
              className="flex-1 bg-blue-600 text-white py-3 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              Далее →
            </button>
          </div>
        </div>
      )}

      {/* Step 4: Agenda Items */}
      {step === 4 && (
        <div className="space-y-4">
          <div className="space-y-2">
            {state.agenda_items.map((item, i) => (
              <div key={i} className="flex gap-2 items-start p-3 bg-white border rounded-lg">
                <span className="text-sm font-medium text-gray-400 mt-2 w-5 shrink-0">{i + 1}.</span>
                <div className="flex-1 space-y-2">
                  <input
                    value={item.text}
                    onChange={e => dispatch({ type: 'UPDATE_AGENDA_ITEM', index: i, item: { ...item, text: e.target.value } })}
                    placeholder="Тема пункта повестки"
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <select
                    value={item.speaker_id || ''}
                    onChange={e => dispatch({ type: 'UPDATE_AGENDA_ITEM', index: i, item: { ...item, speaker_id: Number(e.target.value) } })}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
                  >
                    <option value="">Выберите докладчика *</option>
                    {state.participants.map(p => (
                      <option key={p.id} value={p.id}>{fullName(p)}</option>
                    ))}
                  </select>
                </div>
                <button
                  onClick={() => dispatch({ type: 'REMOVE_AGENDA_ITEM', index: i })}
                  className="text-gray-400 hover:text-red-500 mt-2 text-xl leading-none"
                >×</button>
              </div>
            ))}
          </div>

          <button
            onClick={() => dispatch({ type: 'ADD_AGENDA_ITEM' })}
            className="w-full border-2 border-dashed border-gray-300 rounded-lg py-3 text-sm text-gray-500 hover:border-blue-400 hover:text-blue-500"
          >
            + Добавить пункт повестки
          </button>

          <div className="flex gap-3">
            <button onClick={goBack} className="flex-1 border text-gray-700 py-3 rounded-lg text-sm hover:bg-gray-50">
              ← Назад
            </button>
            <button
              disabled={!canProceedStep4}
              onClick={goNext}
              className="flex-1 bg-blue-600 text-white py-3 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              Далее ({state.agenda_items.length}) →
            </button>
          </div>
        </div>
      )}

      {/* Step 5: Review + Submit */}
      {step === 5 && (
        <div className="space-y-4">
          <div className="bg-white border rounded-lg p-4 space-y-3">
            <div>
              <p className="text-xs text-gray-500">Тема</p>
              <p className="text-sm font-medium mt-0.5">{state.title}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500">Дата и время</p>
              <p className="text-sm font-medium mt-0.5">
                {new Date(state.date).toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500">Председательствующий</p>
              <p className="text-sm font-medium mt-0.5">
                {fullName(state.participants.find(p => p.id === state.chairperson_id)!)}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500">Участники ({state.participants.length})</p>
              {state.participants.map(p => (
                <p key={p.id} className="text-sm mt-0.5">{fullName(p)}</p>
              ))}
            </div>
            <div>
              <p className="text-xs text-gray-500">Повестка ({state.agenda_items.length} пунктов)</p>
              {state.agenda_items.map((item, i) => (
                <p key={i} className="text-sm mt-0.5">
                  {i + 1}. {item.text}
                  {' — '}
                  {fullName(state.participants.find(p => p.id === item.speaker_id)!)}
                </p>
              ))}
            </div>
          </div>

          {createMeetingMutation.isError && (
            <p className="text-sm text-red-500">Ошибка создания. Попробуйте снова.</p>
          )}

          <div className="flex gap-3">
            <button onClick={goBack} className="flex-1 border text-gray-700 py-3 rounded-lg text-sm hover:bg-gray-50">
              ← Назад
            </button>
            <button
              onClick={() => createMeetingMutation.mutate()}
              disabled={createMeetingMutation.isPending}
              className="flex-1 bg-green-600 text-white py-3 rounded-lg text-sm font-semibold hover:bg-green-700 disabled:opacity-50"
            >
              {createMeetingMutation.isPending ? 'Сохранение...' : 'Зафиксировать'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
