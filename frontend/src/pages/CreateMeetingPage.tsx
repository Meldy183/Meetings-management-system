import { useReducer, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createMeeting, addMeetingPerson, reorderPeople, setChairperson, addAgendaItem } from '../api/meetings'
import { updatePerson, sortPeople } from '../api/people'
import { ParticipantSearch } from '../components/ParticipantSearch'
import { ParticipantCard } from '../components/ParticipantCard'
import { ParticipantForm } from '../components/ParticipantForm'
import { StepIndicator } from '../components/StepIndicator'
import { SpeakerPicker } from '../components/SpeakerPicker'
import type { Person, PersonCreate } from '../api/types'

interface AgendaItem {
  text: string
  speaker_ids: number[]
}

interface WizardState {
  title: string
  date: string
  place: string
  people: Person[]
  chairperson_id: number | null
  agenda_items: AgendaItem[]
}

type WizardAction =
  | { type: 'SET_TITLE_DATE'; title: string; date: string; place: string }
  | { type: 'ADD_PERSON'; person: Person }
  | { type: 'REORDER_PEOPLE'; people: Person[] }
  | { type: 'REMOVE_PERSON'; id: number }
  | { type: 'UPDATE_PERSON'; person: Person }
  | { type: 'SET_CHAIRPERSON'; id: number }
  | { type: 'ADD_AGENDA_ITEM' }
  | { type: 'UPDATE_AGENDA_ITEM'; index: number; item: AgendaItem }
  | { type: 'REMOVE_AGENDA_ITEM'; index: number }
  | { type: 'REORDER_AGENDA_ITEMS'; items: AgendaItem[] }

function reducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'SET_TITLE_DATE':
      return { ...state, title: action.title, date: action.date, place: action.place }
    case 'ADD_PERSON':
      if (state.people.find(p => p.id === action.person.id)) return state
      return { ...state, people: [...state.people, action.person] }
    case 'REORDER_PEOPLE':
      return { ...state, people: action.people }
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
    case 'REORDER_AGENDA_ITEMS':
      return { ...state, agenda_items: action.items }
    default:
      return state
  }
}

const STEP_LABELS = ['Тема', 'Участники', 'Председатель', 'Повестка', 'Подтверждение']

const initialState: WizardState = {
  title: '',
  date: '',
  place: '',
  people: [],
  chairperson_id: null,
  agenda_items: [],
}

export function CreateMeetingPage() {
  const [step, setStep] = useState(1)
  const [state, dispatch] = useReducer(reducer, initialState)
  const [editingPersonId, setEditingPersonId] = useState<number | null>(null)
  const [titleInput, setTitleInput] = useState('Совещание по вопросам ')
  const [dateInput, setDateInput] = useState(() => {
    const d = new Date()
    d.setDate(d.getDate() + 1)
    return d.toISOString().slice(0, 10)
  })
  const [timeInput, setTimeInput] = useState('10:00')
  const [placeInput, setPlaceInput] = useState('')
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)
  const dragIndexRef = useRef<number | null>(null)
  const s5Ref = useRef<{ ctx: string; from: number } | null>(null)
  const [s5Over, setS5Over] = useState<{ ctx: string; idx: number } | null>(null)
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
        date: state.date + ':00.000Z',
        ...(state.place ? { place: state.place } : {}),
      })
      for (const p of state.people) {
        await addMeetingPerson(meeting.id, p.id)
      }
      await reorderPeople(meeting.id, state.people.map(p => p.id))
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

  function handleAddPerson(p: Person) {
    dispatch({ type: 'ADD_PERSON', person: p })
  }

  async function handleSort() {
    if (state.people.length === 0) return
    try {
      const sortedIds = await sortPeople(state.people.map(p => p.id))
      dispatch({ type: 'REORDER_PEOPLE', people: sortedIds.map(id => state.people.find(p => p.id === id)!) })
    } catch { /* ignore */ }
  }

  function handleDragStart(i: number) { dragIndexRef.current = i }
  function handleDragOver(e: React.DragEvent, i: number) { e.preventDefault(); setDragOverIndex(i) }
  function handleDragEnd() { dragIndexRef.current = null; setDragOverIndex(null) }
  function handleDrop(i: number) {
    const from = dragIndexRef.current
    dragIndexRef.current = null
    setDragOverIndex(null)
    if (from === null || from === i) return
    const next = [...state.people]
    const [moved] = next.splice(from, 1)
    next.splice(i, 0, moved)
    dispatch({ type: 'REORDER_PEOPLE', people: next })
  }

  function s5DragStart(ctx: string, from: number) { s5Ref.current = { ctx, from } }
  function s5DragOver(e: React.DragEvent, ctx: string, idx: number) {
    e.preventDefault()
    if (s5Ref.current?.ctx === ctx) setS5Over({ ctx, idx })
  }
  function s5DragEnd() { s5Ref.current = null; setS5Over(null) }
  function s5Drop(ctx: string, to: number) {
    const src = s5Ref.current
    s5Ref.current = null
    setS5Over(null)
    if (!src || src.ctx !== ctx || src.from === to) return
    if (ctx === 'people') {
      const chair = state.people.find(p => p.id === state.chairperson_id)
      const others = state.people.filter(p => p.id !== state.chairperson_id)
      const next = [...others]
      const [moved] = next.splice(src.from, 1)
      next.splice(to, 0, moved)
      dispatch({ type: 'REORDER_PEOPLE', people: [...(chair ? [chair] : []), ...next] })
    } else if (ctx === 'agenda') {
      const next = [...state.agenda_items]
      const [moved] = next.splice(src.from, 1)
      next.splice(to, 0, moved)
      dispatch({ type: 'REORDER_AGENDA_ITEMS', items: next })
    } else if (ctx.startsWith('speakers-')) {
      const ai = parseInt(ctx.slice('speakers-'.length))
      const item = state.agenda_items[ai]
      const next = [...item.speaker_ids]
      const [moved] = next.splice(src.from, 1)
      next.splice(to, 0, moved)
      dispatch({ type: 'UPDATE_AGENDA_ITEM', index: ai, item: { ...item, speaker_ids: next } })
    }
  }

  const canProceedStep1 = titleInput.trim() && dateInput && timeInput
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
            <label className="block text-sm font-medium text-gray-700 mb-1">Тема *</label>
            <textarea
              value={titleInput}
              onChange={e => setTitleInput(e.target.value)}
              rows={2}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 resize-none"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Дата *</label>
              <input
                type="date"
                value={dateInput}
                onChange={e => setDateInput(e.target.value)}
                className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Время *</label>
              <input
                type="time"
                value={timeInput}
                onChange={e => setTimeInput(e.target.value)}
                className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Место</label>
            <input
              value={placeInput}
              onChange={e => setPlaceInput(e.target.value)}
              placeholder="г. Москва, ул. Тверская, д. 13"
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
            />
          </div>
          <button
            disabled={!canProceedStep1}
            onClick={() => {
              dispatch({ type: 'SET_TITLE_DATE', title: titleInput.trim(), date: `${dateInput}T${timeInput}`, place: placeInput.trim() })
              goNext()
            }}
            className="w-full bg-green-600 text-white py-3 rounded-lg font-medium text-sm hover:bg-green-700 disabled:opacity-50"
          >
            Далее →
          </button>
        </div>
      )}

      {/* Step 2: People */}
      {step === 2 && (
        <div className="space-y-4">
          <p className="text-sm text-gray-500">Найдите участников по имени или добавьте нового в базу данных.</p>
          <ParticipantSearch
            onAdd={handleAddPerson}
            existingIds={state.people.map(p => p.id)}
          />

          {state.people.length > 0 && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <p className="text-sm font-medium text-gray-700">
                  Список участников ({state.people.length})
                </p>
                <button
                  onClick={handleSort}
                  className="text-xs px-2 py-0.5 rounded border border-gray-300 text-gray-500 bg-white hover:bg-gray-50"
                >
                  Сортировать
                </button>
              </div>
              <div className="space-y-2">
                {state.people.map((p, i) => (
                  <div
                    key={p.id}
                    draggable
                    onDragStart={() => handleDragStart(i)}
                    onDragOver={e => handleDragOver(e, i)}
                    onDrop={() => handleDrop(i)}
                    onDragEnd={handleDragEnd}
                    className={[
                      'rounded-lg transition-opacity cursor-grab',
                      dragOverIndex === i && dragIndexRef.current !== i ? 'ring-2 ring-green-400' : '',
                      dragIndexRef.current === i ? 'opacity-40' : 'opacity-100',
                    ].join(' ')}
                  >
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
                        dragHandle
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
              className="flex-1 bg-green-600 text-white py-3 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50"
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
              className="flex-1 bg-green-600 text-white py-3 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50"
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
                <div className="flex-1 min-w-0 space-y-2">
                  <input
                    value={item.text}
                    onChange={e => dispatch({ type: 'UPDATE_AGENDA_ITEM', index: i, item: { ...item, text: e.target.value } })}
                    placeholder="Тема пункта повестки"
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                  />
                  <div>
                    <p className="text-xs font-medium text-gray-600 mb-1">Докладчики</p>
                    <SpeakerPicker
                      people={state.people}
                      speakerIds={item.speaker_ids}
                      onChange={ids => dispatch({ type: 'UPDATE_AGENDA_ITEM', index: i, item: { ...item, speaker_ids: ids } })}
                    />
                  </div>
                </div>
                <button
                  onClick={() => dispatch({ type: 'REMOVE_AGENDA_ITEM', index: i })}
                  className="text-gray-400 hover:text-red-500 mt-2 text-xl leading-none"
                >×</button>
              </div>
            ))}
          </div>

          {state.agenda_items.length === 0 && (
            <p className="text-sm text-gray-400 text-center py-1">Добавьте хотя бы один пункт повестки</p>
          )}
          <button
            onClick={() => dispatch({ type: 'ADD_AGENDA_ITEM' })}
            className="w-full border-2 border-dashed border-gray-300 rounded-lg py-3 text-sm text-gray-500 hover:border-green-400 hover:text-green-500"
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
              className="flex-1 bg-green-600 text-white py-3 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50"
            >
              Далее ({state.agenda_items.length}) →
            </button>
          </div>
        </div>
      )}

      {/* Step 5: Review + Submit */}
      {step === 5 && (
        <div className="space-y-4">
          <div className="bg-white border rounded-lg p-4 divide-y divide-gray-100">
            <div className="pb-3">
              <p className="text-xs text-gray-500">Тема</p>
              <p className="text-sm font-medium mt-0.5">{state.title}</p>
            </div>
            <div className="py-3">
              <p className="text-xs text-gray-500">Дата и время</p>
              <p className="text-sm font-medium mt-0.5">
                {new Date(state.date + ':00.000Z').toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })}
              </p>
            </div>
            {state.place && (
              <div className="py-3">
                <p className="text-xs text-gray-500">Место</p>
                <p className="text-sm font-medium mt-0.5">{state.place}</p>
              </div>
            )}
            <div className="py-3">
              <p className="text-xs text-gray-500">Председательствующий</p>
              <p className="text-sm font-medium mt-0.5">
                {fullName(state.people.find(p => p.id === state.chairperson_id)!)}
              </p>
            </div>
            <div className="py-3">
              <div className="flex items-center justify-between mb-1">
                <p className="text-xs text-gray-500">Участники ({state.people.length})</p>
                <button
                  onClick={handleSort}
                  className="text-xs px-2 py-0.5 rounded border border-gray-300 text-gray-500 bg-white hover:bg-gray-50"
                >
                  Сортировать
                </button>
              </div>
              {(() => {
                const chair = state.people.find(p => p.id === state.chairperson_id)
                const others = state.people.filter(p => p.id !== state.chairperson_id)
                return (
                  <div className="space-y-1">
                    {chair && (
                      <div className="flex items-center gap-2 px-2 py-1 rounded bg-gray-50">
                        <span className="text-gray-300 text-sm select-none w-3">⠿</span>
                        <p className="text-sm">{fullName(chair)} <span className="text-gray-400 text-xs">(председатель)</span></p>
                      </div>
                    )}
                    {others.map((p, i) => (
                      <div
                        key={p.id}
                        draggable
                        onDragStart={() => s5DragStart('people', i)}
                        onDragOver={e => s5DragOver(e, 'people', i)}
                        onDrop={() => s5Drop('people', i)}
                        onDragEnd={s5DragEnd}
                        className={[
                          'flex items-center gap-2 px-2 py-1 rounded cursor-grab',
                          s5Over?.ctx === 'people' && s5Over.idx === i && s5Ref.current?.from !== i ? 'bg-green-50 border border-green-300' : 'bg-gray-50',
                          s5Ref.current?.ctx === 'people' && s5Ref.current.from === i ? 'opacity-40' : '',
                        ].join(' ')}
                      >
                        <span className="text-gray-400 text-sm select-none w-3">⠿</span>
                        <p className="text-sm">{fullName(p)}</p>
                      </div>
                    ))}
                  </div>
                )
              })()}
            </div>
            <div className="pt-3">
              <p className="text-xs text-gray-500 mb-1">Повестка ({state.agenda_items.length} пунктов)</p>
              <div className="space-y-1.5">
                {state.agenda_items.map((item, i) => {
                  const speakersCtx = `speakers-${i}`
                  return (
                    <div
                      key={i}
                      draggable
                      onDragStart={() => s5DragStart('agenda', i)}
                      onDragOver={e => s5DragOver(e, 'agenda', i)}
                      onDrop={() => s5Drop('agenda', i)}
                      onDragEnd={s5DragEnd}
                      className={[
                        'p-2 border rounded-lg cursor-grab',
                        s5Over?.ctx === 'agenda' && s5Over.idx === i && s5Ref.current?.from !== i ? 'border-green-400 bg-green-50' : 'bg-white',
                        s5Ref.current?.ctx === 'agenda' && s5Ref.current.from === i ? 'opacity-40' : '',
                      ].join(' ')}
                    >
                      <div className="flex items-start gap-2">
                        <span className="text-gray-400 text-sm select-none w-3 mt-0.5">⠿</span>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm">{i + 1}. {item.text}</p>
                          <div className="mt-1 space-y-0.5">
                            {item.speaker_ids.map((sid, j) => {
                              const sp = state.people.find(p => p.id === sid)
                              if (!sp) return null
                              return (
                                <div
                                  key={sid}
                                  draggable
                                  onDragStart={e => { e.stopPropagation(); s5DragStart(speakersCtx, j) }}
                                  onDragOver={e => { e.stopPropagation(); s5DragOver(e, speakersCtx, j) }}
                                  onDrop={e => { e.stopPropagation(); s5Drop(speakersCtx, j) }}
                                  onDragEnd={e => { e.stopPropagation(); s5DragEnd() }}
                                  className={[
                                    'flex items-center gap-1.5 pl-2 py-0.5 rounded cursor-grab',
                                    s5Over?.ctx === speakersCtx && s5Over.idx === j && s5Ref.current?.from !== j ? 'bg-green-100' : '',
                                    s5Ref.current?.ctx === speakersCtx && s5Ref.current.from === j ? 'opacity-40' : '',
                                  ].join(' ')}
                                >
                                  <span className="text-gray-300 text-xs select-none w-2.5">⠿</span>
                                  <p className="text-sm text-gray-500">{fullName(sp)}</p>
                                </div>
                              )
                            })}
                          </div>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
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
