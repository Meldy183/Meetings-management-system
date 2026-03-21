import { useReducer, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createMeeting, addMeetingPerson, setChairperson, addAgendaItem } from '../api/meetings'
import { updatePerson } from '../api/people'
import { ParticipantSearch } from '../components/ParticipantSearch'
import { ParticipantCard } from '../components/ParticipantCard'
import { ParticipantForm } from '../components/ParticipantForm'
import { StepIndicator } from '../components/StepIndicator'
import type { Person, PersonCreate } from '../api/types'

interface AgendaItem {
  text: string
  speaker_ids: number[]
}

interface WizardState {
  title: string
  date: string
  people: Person[]
  chairperson_id: number | null
  agenda_items: AgendaItem[]
}

type WizardAction =
  | { type: 'SET_TITLE_DATE'; title: string; date: string }
  | { type: 'ADD_PERSON'; person: Person }
  | { type: 'REMOVE_PERSON'; id: number }
  | { type: 'UPDATE_PERSON'; person: Person }
  | { type: 'SET_CHAIRPERSON'; id: number }
  | { type: 'ADD_AGENDA_ITEM' }
  | { type: 'UPDATE_AGENDA_ITEM'; index: number; item: AgendaItem }
  | { type: 'REMOVE_AGENDA_ITEM'; index: number }

function reducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'SET_TITLE_DATE':
      return { ...state, title: action.title, date: action.date }
    case 'ADD_PERSON':
      if (state.people.find(p => p.id === action.person.id)) return state
      return { ...state, people: [...state.people, action.person] }
    case 'REMOVE_PERSON': {
      const newPeople = state.people.filter(p => p.id !== action.id)
      return {
        ...state,
        people: newPeople,
        chairperson_id: state.chairperson_id === action.id ? null : state.chairperson_id,
        agenda_items: state.agenda_items.map(item => ({
          ...item,
          speaker_ids: item.speaker_ids.filter(sid => sid !== action.id),
        })),
      }
    }
    case 'UPDATE_PERSON':
      return { ...state, people: state.people.map(p => p.id === action.person.id ? action.person : p) }
    case 'SET_CHAIRPERSON':
      return { ...state, chairperson_id: action.id }
    case 'ADD_AGENDA_ITEM':
      return { ...state, agenda_items: [...state.agenda_items, { text: '', speaker_ids: [] }] }
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
  people: [],
  chairperson_id: null,
  agenda_items: [],
}

export function CreateMeetingPage() {
  const [step, setStep] = useState(1)
  const [state, dispatch] = useReducer(reducer, initialState)
  const [editingPersonId, setEditingPersonId] = useState<number | null>(null)
  const [titleInput, setTitleInput] = useState('')
  const [dateInput, setDateInput] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const updatePersonMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: PersonCreate }) => updatePerson(id, data),
    onSuccess: (updated) => {
      dispatch({ type: 'UPDATE_PERSON', person: updated })
      setEditingPersonId(null)
    },
  })

  async function handleSubmit() {
    setIsSubmitting(true)
    setSubmitError(null)
    try {
      const meeting = await createMeeting({
        title: state.title,
        date: new Date(state.date).toISOString(),
      })
      for (const p of state.people) {
        await addMeetingPerson(meeting.id, p.id)
      }
      if (state.chairperson_id !== null) {
        await setChairperson(meeting.id, state.chairperson_id)
      }
      for (const item of state.agenda_items) {
        await addAgendaItem(meeting.id, { text: item.text, speaker_ids: item.speaker_ids })
      }
      queryClient.invalidateQueries({ queryKey: ['meetings'] })
      navigate(`/meetings/${meeting.id}`)
    } catch {
      setSubmitError('Ошибка создания совещания. Попробуйте снова.')
    } finally {
      setIsSubmitting(false)
    }
  }

  function goNext() { setStep(s => s + 1) }
  function goBack() { setStep(s => s - 1) }

  function fullName(p: Person) {
    return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ')
  }

  const canProceedStep1 = titleInput.trim() && dateInput
  const canProceedStep2 = state.people.length > 0
  const canProceedStep3 = state.chairperson_id !== null
  const canProceedStep4 = state.agenda_items.length > 0 &&
    state.agenda_items.every(item => item.text.trim() && item.speaker_ids.length > 0)

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

      {/* Step 2: People */}
      {step === 2 && (
        <div className="space-y-4">
          <ParticipantSearch
            onAdd={p => dispatch({ type: 'ADD_PERSON', person: p })}
            existingIds={state.people.map(p => p.id)}
          />

          {state.people.length > 0 && (
            <div>
              <p className="text-sm font-medium text-gray-700 mb-2">
                Список участников ({state.people.length})
              </p>
              <div className="space-y-2">
                {state.people.map(p => (
                  <div key={p.id}>
                    {editingPersonId === p.id ? (
                      <div className="p-4 border rounded-lg bg-gray-50">
                        <ParticipantForm
                          defaultValues={{ last_name: p.last_name, first_name: p.first_name, middle_name: p.middle_name, info: p.info }}
                          onSubmit={data => updatePersonMutation.mutate({ id: p.id, data })}
                          onCancel={() => setEditingPersonId(null)}
                          isLoading={updatePersonMutation.isPending}
                        />
                      </div>
                    ) : (
                      <ParticipantCard
                        participant={p}
                        onEdit={() => setEditingPersonId(p.id)}
                        onRemove={() => dispatch({ type: 'REMOVE_PERSON', id: p.id })}
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
              Готово ({state.people.length}) →
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Chairperson */}
      {step === 3 && (
        <div className="space-y-4">
          <p className="text-sm text-gray-600">Выберите председательствующего:</p>
          <div className="space-y-2">
            {state.people.map(p => (
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
                  <div>
                    <p className="text-xs font-medium text-gray-600 mb-1">Докладчики *</p>
                    <div className="space-y-1">
                      {state.people.map(p => {
                        const checked = item.speaker_ids.includes(p.id)
                        return (
                          <label key={p.id} className="flex items-center gap-2 cursor-pointer">
                            <input
                              type="checkbox"
                              checked={checked}
                              onChange={() => {
                                const newIds = checked
                                  ? item.speaker_ids.filter(id => id !== p.id)
                                  : [...item.speaker_ids, p.id]
                                dispatch({ type: 'UPDATE_AGENDA_ITEM', index: i, item: { ...item, speaker_ids: newIds } })
                              }}
                              className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                            />
                            <span className="text-sm text-gray-700">{fullName(p)}</span>
                          </label>
                        )
                      })}
                    </div>
                  </div>
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
                {fullName(state.people.find(p => p.id === state.chairperson_id)!)}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500">Участники ({state.people.length})</p>
              {state.people.map(p => (
                <p key={p.id} className="text-sm mt-0.5">{fullName(p)}</p>
              ))}
            </div>
            <div>
              <p className="text-xs text-gray-500">Повестка ({state.agenda_items.length} пунктов)</p>
              {state.agenda_items.map((item, i) => {
                const speakers = item.speaker_ids
                  .map(id => state.people.find(p => p.id === id))
                  .filter(Boolean)
                  .map(p => fullName(p!))
                  .join(', ')
                return (
                  <p key={i} className="text-sm mt-0.5">
                    {i + 1}. {item.text}{speakers ? ` — ${speakers}` : ''}
                  </p>
                )
              })}
            </div>
          </div>

          {submitError && (
            <p className="text-sm text-red-500">{submitError}</p>
          )}

          <div className="flex gap-3">
            <button onClick={goBack} className="flex-1 border text-gray-700 py-3 rounded-lg text-sm hover:bg-gray-50">
              ← Назад
            </button>
            <button
              onClick={handleSubmit}
              disabled={isSubmitting}
              className="flex-1 bg-green-600 text-white py-3 rounded-lg text-sm font-semibold hover:bg-green-700 disabled:opacity-50"
            >
              {isSubmitting ? 'Сохранение...' : 'Зафиксировать'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
