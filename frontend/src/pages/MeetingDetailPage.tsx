import { useRef, useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getMeeting, downloadAgenda, downloadParticipants,
  reorderPeople, reorderAgendaItems, reorderAgendaItemSpeakers,
  updateMeeting, setChairperson, deleteMeeting,
  addMeetingPerson, removeMeetingPerson,
  addAgendaItem, updateAgendaItem, deleteAgendaItem,
} from '../api/meetings'
import { getPeople } from '../api/people'
import { ApiError } from '../api/client'
import { SpeakerPicker } from '../components/SpeakerPicker'
import type { Person, AgendaItem, Meeting } from '../api/types'

function useDragReorder<T>(
  items: T[],
  onDrop: (reordered: T[]) => void,
) {
  const dragIndex = useRef<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  function handleDragStart(i: number) { dragIndex.current = i }
  function handleDragOver(e: React.DragEvent, i: number) { e.preventDefault(); setDragOverIndex(i) }
  function handleDrop(i: number) {
    const from = dragIndex.current
    dragIndex.current = null
    setDragOverIndex(null)
    if (from === null || from === i) return
    const next = [...items]
    const [moved] = next.splice(from, 1)
    next.splice(i, 0, moved)
    onDrop(next)
  }
  function handleDragEnd() { dragIndex.current = null; setDragOverIndex(null) }

  return { dragIndex, dragOverIndex, handleDragStart, handleDragOver, handleDrop, handleDragEnd }
}

function fullName(p: Person) {
  return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ')
}


export function MeetingDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [downloading, setDownloading] = useState<'agenda' | 'participants' | null>(null)

  // DnD local state
  const [agendaItems, setAgendaItems] = useState<AgendaItem[]>([])
  const [people, setPeople] = useState<Person[]>([])

  // Speaker drag-and-drop state (one drag active at a time across all agenda items)
  const speakerDragSrc = useRef<{ itemId: number; index: number } | null>(null)
  const [speakerDragOver, setSpeakerDragOver] = useState<{ itemId: number; index: number } | null>(null)

  // Edit meeting metadata state
  const [editingMeeting, setEditingMeeting] = useState(false)
  const [meetingForm, setMeetingForm] = useState({ title: '', date: '', place: '' })

  // Chairperson state
  const [editingChairperson, setEditingChairperson] = useState(false)
  const [chairpersonId, setChairpersonId] = useState(0)

  // Edit agenda item state
  const [editingItemId, setEditingItemId] = useState<number | null>(null)
  const [itemForm, setItemForm] = useState({ text: '', speaker_ids: [] as number[] })

  // Add agenda item state
  const [showAddItem, setShowAddItem] = useState(false)
  const [newItem, setNewItem] = useState({ text: '', speaker_ids: [] as number[] })

  // Add person search state
  const [personQuery, setPersonQuery] = useState('')
  const [debouncedPersonQuery, setDebouncedPersonQuery] = useState('')
  const [personError, setPersonError] = useState<string | null>(null)

  const { data: meeting, isLoading, isError } = useQuery({
    queryKey: ['meeting', id],
    queryFn: () => getMeeting(id!),
    enabled: !!id,
  })

  useEffect(() => {
    if (meeting) {
      setAgendaItems(meeting.agenda_items)
      setPeople(meeting.people)
    }
  }, [meeting])

  useEffect(() => {
    const t = setTimeout(() => setDebouncedPersonQuery(personQuery), 300)
    return () => clearTimeout(t)
  }, [personQuery])

  const { data: searchResults = [], isFetching: isSearching } = useQuery({
    queryKey: ['people', 'search', debouncedPersonQuery],
    queryFn: () => getPeople(debouncedPersonQuery),
    enabled: debouncedPersonQuery.trim().length > 0,
  })

  function setMeetingData(updated: Meeting) {
    queryClient.setQueryData(['meeting', id], updated)
  }

  // DnD mutations
  const agendaMutation = useMutation({
    mutationFn: (ids: number[]) => reorderAgendaItems(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => { if (meeting) setAgendaItems(meeting.agenda_items) },
  })

  const peopleMutation = useMutation({
    mutationFn: (ids: number[]) => reorderPeople(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => { if (meeting) setPeople(meeting.people) },
  })

  const updateMeetingMutation = useMutation({
    mutationFn: (data: { title: string; date: string; place?: string }) => updateMeeting(id!, data),
    onSuccess: (updated) => { setMeetingData(updated); setEditingMeeting(false) },
  })

  const setChairpersonMutation = useMutation({
    mutationFn: (personId: number) => setChairperson(id!, personId),
    onSuccess: (updated) => { setMeetingData(updated); setEditingChairperson(false) },
    onError: (e) => {
      if (e instanceof ApiError) alert(e.message)
    },
  })

  const deleteMeetingMutation = useMutation({
    mutationFn: () => deleteMeeting(id!),
    onSuccess: () => navigate('/'),
  })

  const addPersonMutation = useMutation({
    mutationFn: (personId: number) => addMeetingPerson(id!, personId),
    onSuccess: (updated) => {
      setMeetingData(updated)
      setPersonQuery('')
      setPersonError(null)
    },
    onError: (e) => {
      if (e instanceof ApiError) setPersonError(e.message)
    },
  })

  const removePersonMutation = useMutation({
    mutationFn: (personId: number) => removeMeetingPerson(id!, personId),
    onSuccess: (updated) => setMeetingData(updated),
    onError: (e) => {
      if (e instanceof ApiError) alert(e.message)
    },
  })

  const addAgendaItemMutation = useMutation({
    mutationFn: (data: { text: string; speaker_ids: number[] }) => addAgendaItem(id!, data),
    onSuccess: (updated) => { setMeetingData(updated); setShowAddItem(false); setNewItem({ text: '', speaker_ids: [] }) },
  })

  const updateAgendaItemMutation = useMutation({
    mutationFn: ({ itemId, data }: { itemId: number; data: { text: string; speaker_ids: number[] } }) =>
      updateAgendaItem(id!, itemId, data),
    onSuccess: (updated) => { setMeetingData(updated); setEditingItemId(null) },
  })

  const deleteAgendaItemMutation = useMutation({
    mutationFn: (itemId: number) => deleteAgendaItem(id!, itemId),
    onSuccess: (updated) => setMeetingData(updated),
  })

  const reorderSpeakersMutation = useMutation({
    mutationFn: ({ itemId, ids }: { itemId: number; ids: number[] }) =>
      reorderAgendaItemSpeakers(id!, itemId, ids),
    onError: () => { if (meeting) setAgendaItems(meeting.agenda_items) },
  })

  function handleSpeakerDragStart(itemId: number, index: number) {
    speakerDragSrc.current = { itemId, index }
  }
  function handleSpeakerDragOver(e: React.DragEvent, itemId: number, index: number) {
    e.preventDefault()
    setSpeakerDragOver({ itemId, index })
  }
  function handleSpeakerDrop(toIndex: number) {
    const src = speakerDragSrc.current
    speakerDragSrc.current = null
    setSpeakerDragOver(null)
    if (!src || src.index === toIndex) return
    setAgendaItems(prev => prev.map(item => {
      if (item.id !== src.itemId) return item
      const next = [...item.speakers]
      const [moved] = next.splice(src.index, 1)
      next.splice(toIndex, 0, moved)
      reorderSpeakersMutation.mutate({ itemId: item.id, ids: next.map(s => s.id) })
      return { ...item, speakers: next }
    }))
  }
  function handleSpeakerDragEnd() {
    speakerDragSrc.current = null
    setSpeakerDragOver(null)
  }

  // DnD hooks
  const agendaDnd = useDragReorder(agendaItems, (reordered) => {
    setAgendaItems(reordered)
    agendaMutation.mutate(reordered.map(i => i.id))
  })

  const peopleDnd = useDragReorder(people, (reordered) => {
    setPeople(reordered)
    peopleMutation.mutate(reordered.map(p => p.id))
  })

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', year: 'numeric',
      hour: '2-digit', minute: '2-digit', timeZone: 'UTC',
    })
  }

  function toDatetimeLocal(iso: string) {
    return iso.slice(0, 16)
  }

  async function handleDownload(type: 'agenda' | 'participants') {
    if (!id) return
    setDownloading(type)
    try {
      if (type === 'agenda') await downloadAgenda(id)
      else await downloadParticipants(id)
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        alert('Совещание не готово к экспорту: назначьте председателя, добавьте участников и повестку')
      }
    } finally {
      setDownloading(null)
    }
  }

  if (isLoading) return <div className="max-w-2xl mx-auto px-4 py-6 text-gray-500 text-sm">Загрузка...</div>
  if (isError || !meeting) return <div className="max-w-2xl mx-auto px-4 py-6 text-red-500 text-sm">Совещание не найдено</div>

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">

      {/* Header */}
      <div className="flex items-start gap-3">
        <Link to="/" className="text-gray-400 hover:text-gray-600 mt-1">←</Link>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h1 className="text-lg font-semibold text-gray-900 leading-snug">{meeting.title}</h1>
            {meeting.status === 'incomplete' && (
              <span className="shrink-0 text-xs bg-yellow-100 text-yellow-700 px-2 py-0.5 rounded-full">
                Не готово
              </span>
            )}
          </div>
        </div>
        <div className="flex gap-2 shrink-0">
          <button
            onClick={() => {
              setMeetingForm({ title: meeting.title, date: toDatetimeLocal(meeting.date), place: meeting.place ?? '' })
              setEditingMeeting(true)
            }}
            className="text-xs text-gray-500 hover:text-green-600 border rounded px-2 py-1"
          >
            Изменить
          </button>
          <button
            onClick={() => { if (confirm('Удалить совещание?')) deleteMeetingMutation.mutate() }}
            className="text-xs text-gray-500 hover:text-red-600 border rounded px-2 py-1"
          >
            Удалить
          </button>
        </div>
      </div>

      {/* Meeting info / edit form */}
      {editingMeeting ? (
        <div className="bg-white border rounded-lg p-4 space-y-3">
          <p className="text-sm font-medium text-gray-700">Редактирование темы и даты</p>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Тема</label>
            <textarea
              value={meetingForm.title}
              onChange={e => setMeetingForm(f => ({ ...f, title: e.target.value }))}
              rows={2}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 resize-none"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Дата и время</label>
            <input
              type="datetime-local"
              value={meetingForm.date}
              onChange={e => setMeetingForm(f => ({ ...f, date: e.target.value }))}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Место</label>
            <input
              value={meetingForm.place}
              onChange={e => setMeetingForm(f => ({ ...f, place: e.target.value }))}
              placeholder="г. Москва, ул. Тверская, д. 13"
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
            />
          </div>
          {updateMeetingMutation.isError && (
            <p className="text-xs text-red-500">
              {updateMeetingMutation.error instanceof ApiError
                ? updateMeetingMutation.error.message
                : 'Ошибка сохранения'}
            </p>
          )}
          <div className="flex gap-2">
            <button
              onClick={() => setEditingMeeting(false)}
              className="flex-1 border text-gray-700 py-2 rounded-lg text-sm hover:bg-gray-50"
            >
              Отмена
            </button>
            <button
              disabled={!meetingForm.title || !meetingForm.date || updateMeetingMutation.isPending}
              onClick={() => updateMeetingMutation.mutate({
                title: meetingForm.title,
                date: meetingForm.date + ':00.000Z',
                place: meetingForm.place,
              })}
              className="flex-1 bg-green-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50"
            >
              {updateMeetingMutation.isPending ? 'Сохранение...' : 'Сохранить'}
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white border rounded-lg p-4 space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">Дата</span>
            <span className="font-medium">{formatDate(meeting.date)}</span>
          </div>
          {meeting.place && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-500">Место</span>
              <span className="font-medium">{meeting.place}</span>
            </div>
          )}
        </div>
      )}

      {/* Chairperson */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">Председательствующий</h2>
          {!editingChairperson && (
            <button
              onClick={() => {
                setChairpersonId(meeting.chairperson?.id ?? 0)
                setEditingChairperson(true)
              }}
              className="text-xs text-gray-500 hover:text-green-600 border rounded px-2 py-1"
            >
              {meeting.chairperson ? 'Изменить' : 'Назначить'}
            </button>
          )}
        </div>
        {editingChairperson ? (
          <div className="bg-white border rounded-lg p-3 space-y-2">
            <select
              value={chairpersonId}
              onChange={e => setChairpersonId(Number(e.target.value))}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 bg-white"
            >
              <option value={0}>Выберите председателя</option>
              {people.map(p => (
                <option key={p.id} value={p.id}>{fullName(p)}</option>
              ))}
            </select>
            <div className="flex gap-2">
              <button
                onClick={() => setEditingChairperson(false)}
                className="flex-1 border text-gray-700 py-1.5 rounded-lg text-xs hover:bg-gray-50"
              >
                Отмена
              </button>
              <button
                disabled={!chairpersonId || setChairpersonMutation.isPending}
                onClick={() => setChairpersonMutation.mutate(chairpersonId)}
                className="flex-1 bg-green-600 text-white py-1.5 rounded-lg text-xs font-medium hover:bg-green-700 disabled:opacity-50"
              >
                {setChairpersonMutation.isPending ? 'Сохранение...' : 'Сохранить'}
              </button>
            </div>
          </div>
        ) : (
          <div className="bg-white border rounded-lg p-3 text-sm">
            {meeting.chairperson ? (
              <div>
                <p className="font-medium">{fullName(meeting.chairperson)}</p>
                {meeting.chairperson.info && <p className="text-xs text-gray-500 mt-0.5">{meeting.chairperson.info}</p>}
              </div>
            ) : (
              <p className="text-gray-400 italic">Не назначен</p>
            )}
          </div>
        )}
      </div>

      {/* People */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">Участники ({people.length})</h2>
          {peopleMutation.isPending && <span className="text-xs text-gray-400">Сохранение...</span>}
        </div>
        <div className="space-y-1">
          {people.map((p, i) => (
            <div
              key={p.id}
              draggable
              onDragStart={() => peopleDnd.handleDragStart(i)}
              onDragOver={(e) => peopleDnd.handleDragOver(e, i)}
              onDrop={() => peopleDnd.handleDrop(i)}
              onDragEnd={peopleDnd.handleDragEnd}
              className={[
                'bg-white border rounded-lg p-3 flex items-center gap-3 transition-opacity',
                peopleDnd.dragOverIndex === i && peopleDnd.dragIndex.current !== i ? 'border-green-400 bg-green-50' : '',
                peopleDnd.dragIndex.current === i ? 'opacity-40' : 'opacity-100',
              ].join(' ')}
            >
              <span className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none text-lg leading-none">⠿</span>
              <span className="text-xs text-gray-400 w-5 shrink-0">{i + 1}.</span>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{fullName(p)}</p>
                {p.info && <p className="text-xs text-gray-500 mt-0.5 truncate">{p.info}</p>}
              </div>
              {meeting.chairperson?.id === p.id && (
                <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full shrink-0">Пред.</span>
              )}
              <button
                onClick={() => { if (confirm(`Удалить ${p.last_name} из совещания?`)) removePersonMutation.mutate(p.id) }}
                className="shrink-0 text-gray-300 hover:text-red-500 text-lg leading-none"
                title="Удалить из совещания"
              >
                ×
              </button>
            </div>
          ))}
        </div>

        {/* Add person search */}
        <div className="mt-3 space-y-2">
          <input
            value={personQuery}
            onChange={e => { setPersonQuery(e.target.value); setPersonError(null) }}
            placeholder="Добавить участника..."
            className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
          />
          {debouncedPersonQuery && !isSearching && searchResults.length > 0 && (
            <div className="border rounded-lg divide-y max-h-48 overflow-y-auto">
              {searchResults.map(p => {
                const alreadyIn = people.some(mp => mp.id === p.id)
                return (
                  <div key={p.id} className="flex items-center justify-between px-3 py-2 bg-white hover:bg-gray-50">
                    <div className="min-w-0">
                      <p className="text-sm font-medium truncate">{fullName(p)}</p>
                      {p.info && <p className="text-xs text-gray-500 truncate">{p.info}</p>}
                    </div>
                    {alreadyIn ? (
                      <span className="text-xs text-gray-400 shrink-0 ml-3">В списке</span>
                    ) : (
                      <button
                        onClick={() => addPersonMutation.mutate(p.id)}
                        disabled={addPersonMutation.isPending}
                        className="shrink-0 ml-3 bg-green-600 text-white px-3 py-1 rounded-lg text-xs font-medium hover:bg-green-700 disabled:opacity-50"
                      >
                        Добавить
                      </button>
                    )}
                  </div>
                )
              })}
            </div>
          )}
          {debouncedPersonQuery && !isSearching && searchResults.length === 0 && (
            <p className="text-xs text-gray-500 px-1">Никого не найдено</p>
          )}
          {personError && <p className="text-xs text-red-500 px-1">{personError}</p>}
        </div>
      </div>

      {/* Agenda items */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">Повестка ({agendaItems.length})</h2>
          {agendaMutation.isPending && <span className="text-xs text-gray-400">Сохранение...</span>}
        </div>
        <div className="space-y-1">
          {agendaItems.map((item, i) => (
            <div key={item.id}>
              {editingItemId === item.id ? (
                <div className="bg-white border rounded-lg p-3 space-y-2">
                  <input
                    value={itemForm.text}
                    onChange={e => setItemForm(f => ({ ...f, text: e.target.value }))}
                    placeholder="Тема пункта повестки"
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                  />
                  <p className="text-xs text-gray-500 font-medium">Докладчики:</p>
                  <SpeakerPicker
                    people={people}
                    speakerIds={itemForm.speaker_ids}
                    onChange={ids => setItemForm(f => ({ ...f, speaker_ids: ids }))}
                  />
                  <div className="flex gap-2">
                    <button
                      onClick={() => setEditingItemId(null)}
                      className="flex-1 border text-gray-700 py-1.5 rounded-lg text-xs hover:bg-gray-50"
                    >
                      Отмена
                    </button>
                    <button
                      disabled={!itemForm.text || itemForm.speaker_ids.length === 0 || updateAgendaItemMutation.isPending}
                      onClick={() => updateAgendaItemMutation.mutate({ itemId: item.id, data: { text: itemForm.text, speaker_ids: itemForm.speaker_ids } })}
                      className="flex-1 bg-green-600 text-white py-1.5 rounded-lg text-xs font-medium hover:bg-green-700 disabled:opacity-50"
                    >
                      Сохранить
                    </button>
                  </div>
                </div>
              ) : (
                <div
                  draggable
                  onDragStart={() => agendaDnd.handleDragStart(i)}
                  onDragOver={(e) => agendaDnd.handleDragOver(e, i)}
                  onDrop={() => agendaDnd.handleDrop(i)}
                  onDragEnd={agendaDnd.handleDragEnd}
                  className={[
                    'bg-white border rounded-lg p-3 flex items-start gap-3 transition-opacity',
                    agendaDnd.dragOverIndex === i && agendaDnd.dragIndex.current !== i ? 'border-green-400 bg-green-50' : '',
                    agendaDnd.dragIndex.current === i ? 'opacity-40' : 'opacity-100',
                  ].join(' ')}
                >
                  <span className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none text-lg leading-none mt-0.5">⠿</span>
                  <span className="text-xs text-gray-400 w-5 shrink-0 mt-0.5">{i + 1}.</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium">{item.text}</p>
                    {item.speakers.length === 1 ? (
                      <p className="text-xs text-gray-500 mt-1">Докладчик: {fullName(item.speakers[0])}</p>
                    ) : (
                      <div className="mt-1">
                        <p className="text-xs text-gray-400 mb-0.5">Докладчики:</p>
                        {item.speakers.map((s, si) => (
                          <div
                            key={s.id}
                            draggable
                            onDragStart={() => handleSpeakerDragStart(item.id, si)}
                            onDragOver={(e) => handleSpeakerDragOver(e, item.id, si)}
                            onDrop={() => handleSpeakerDrop(si)}
                            onDragEnd={handleSpeakerDragEnd}
                            className={[
                              'flex items-center gap-1.5 text-xs text-gray-500 rounded px-1 py-0.5 transition-colors cursor-grab active:cursor-grabbing',
                              speakerDragOver?.itemId === item.id && speakerDragOver?.index === si
                                ? 'bg-blue-50 text-blue-700'
                                : '',
                            ].join(' ')}
                          >
                            <span className="text-gray-300 select-none">⠿</span>
                            {fullName(s)}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                  <div className="flex gap-1 shrink-0">
                    <button
                      onClick={() => {
                        setEditingItemId(item.id)
                        setItemForm({ text: item.text, speaker_ids: item.speakers.map(s => s.id) })
                      }}
                      className="text-xs text-gray-400 hover:text-green-600 px-1.5 py-1 rounded"
                    >
                      ✎
                    </button>
                    <button
                      onClick={() => { if (confirm('Удалить пункт повестки?')) deleteAgendaItemMutation.mutate(item.id) }}
                      className="text-gray-300 hover:text-red-500 text-lg leading-none px-1"
                    >
                      ×
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Add agenda item */}
        {showAddItem ? (
          <div className="mt-2 bg-white border rounded-lg p-3 space-y-2">
            <input
              value={newItem.text}
              onChange={e => setNewItem(f => ({ ...f, text: e.target.value }))}
              placeholder="Тема пункта повестки"
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
            />
            <p className="text-xs text-gray-500 font-medium">Докладчики:</p>
            <SpeakerPicker
              people={people}
              speakerIds={newItem.speaker_ids}
              onChange={ids => setNewItem(f => ({ ...f, speaker_ids: ids }))}
            />
            <div className="flex gap-2">
              <button
                onClick={() => { setShowAddItem(false); setNewItem({ text: '', speaker_ids: [] }) }}
                className="flex-1 border text-gray-700 py-1.5 rounded-lg text-xs hover:bg-gray-50"
              >
                Отмена
              </button>
              <button
                disabled={!newItem.text || newItem.speaker_ids.length === 0 || addAgendaItemMutation.isPending}
                onClick={() => addAgendaItemMutation.mutate({ text: newItem.text, speaker_ids: newItem.speaker_ids })}
                className="flex-1 bg-green-600 text-white py-1.5 rounded-lg text-xs font-medium hover:bg-green-700 disabled:opacity-50"
              >
                Добавить
              </button>
            </div>
          </div>
        ) : (
          <button
            onClick={() => setShowAddItem(true)}
            className="mt-2 w-full border-2 border-dashed border-gray-300 rounded-lg py-2 text-sm text-gray-500 hover:border-blue-400 hover:text-blue-500"
          >
            + Добавить пункт повестки
          </button>
        )}
      </div>

      {/* Export */}
      <div className="flex gap-3 pt-2">
        <button
          onClick={() => handleDownload('agenda')}
          disabled={downloading === 'agenda'}
          className="flex-1 bg-white border border-gray-300 text-gray-700 px-4 py-3 rounded-lg text-sm font-medium hover:bg-gray-50 disabled:opacity-50"
        >
          {downloading === 'agenda' ? 'Загрузка...' : '↓ Повестка (.docx)'}
        </button>
        <button
          onClick={() => handleDownload('participants')}
          disabled={downloading === 'participants'}
          className="flex-1 bg-white border border-gray-300 text-gray-700 px-4 py-3 rounded-lg text-sm font-medium hover:bg-gray-50 disabled:opacity-50"
        >
          {downloading === 'participants' ? 'Загрузка...' : '↓ Список участников (.docx)'}
        </button>
      </div>
    </div>
  )
}
