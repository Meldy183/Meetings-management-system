import { useRef, useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getMeeting, downloadAgenda, downloadParticipants,
  reorderParticipants, reorderAgendaItems,
  updateMeeting, deleteMeeting,
  addMeetingParticipant, removeMeetingParticipant,
  addAgendaItem, updateAgendaItem, deleteAgendaItem,
} from '../api/meetings'
import { getParticipants } from '../api/participants'
import { ApiError } from '../api/client'
import type { Participant, AgendaItem, Meeting } from '../api/types'

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

function fullName(p: Participant) {
  return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ')
}

export function MeetingDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [downloading, setDownloading] = useState<'agenda' | 'participants' | null>(null)

  // DnD local state
  const [agendaItems, setAgendaItems] = useState<AgendaItem[]>([])
  const [participants, setParticipants] = useState<Participant[]>([])

  // Edit meeting state
  const [editingMeeting, setEditingMeeting] = useState(false)
  const [meetingForm, setMeetingForm] = useState({ title: '', date: '', chairperson_id: 0 })

  // Edit agenda item state
  const [editingItemId, setEditingItemId] = useState<number | null>(null)
  const [itemForm, setItemForm] = useState({ text: '', speaker_id: 0 })

  // Add agenda item state
  const [showAddItem, setShowAddItem] = useState(false)
  const [newItem, setNewItem] = useState({ text: '', speaker_id: 0 })

  // Add participant search state
  const [participantQuery, setParticipantQuery] = useState('')
  const [debouncedParticipantQuery, setDebouncedParticipantQuery] = useState('')
  const [participantError, setParticipantError] = useState<string | null>(null)

  const { data: meeting, isLoading, isError } = useQuery({
    queryKey: ['meeting', id],
    queryFn: () => getMeeting(id!),
    enabled: !!id,
  })

  useEffect(() => {
    if (meeting) {
      setAgendaItems(meeting.agenda_items)
      setParticipants(meeting.participants)
    }
  }, [meeting])

  useEffect(() => {
    const t = setTimeout(() => setDebouncedParticipantQuery(participantQuery), 300)
    return () => clearTimeout(t)
  }, [participantQuery])

  const { data: searchResults = [], isFetching: isSearching } = useQuery({
    queryKey: ['participants', 'search', debouncedParticipantQuery],
    queryFn: () => getParticipants(debouncedParticipantQuery),
    enabled: debouncedParticipantQuery.trim().length > 0,
  })

  function setMeetingData(updated: Meeting) {
    queryClient.setQueryData(['meeting', id], updated)
  }

  // DnD mutations (existing)
  const agendaMutation = useMutation({
    mutationFn: (ids: number[]) => reorderAgendaItems(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => { if (meeting) setAgendaItems(meeting.agenda_items) },
  })

  const participantsMutation = useMutation({
    mutationFn: (ids: number[]) => reorderParticipants(id!, ids),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['meeting', id] }),
    onError: () => { if (meeting) setParticipants(meeting.participants) },
  })

  // New mutations
  const updateMeetingMutation = useMutation({
    mutationFn: (data: { title: string; date: string; chairperson_id: number }) =>
      updateMeeting(id!, data),
    onSuccess: (updated) => { setMeetingData(updated); setEditingMeeting(false) },
  })

  const deleteMeetingMutation = useMutation({
    mutationFn: () => deleteMeeting(id!),
    onSuccess: () => navigate('/'),
  })

  const addParticipantMutation = useMutation({
    mutationFn: (participantId: number) => addMeetingParticipant(id!, participantId),
    onSuccess: (updated) => {
      setMeetingData(updated)
      setParticipantQuery('')
      setParticipantError(null)
    },
    onError: (e) => {
      if (e instanceof ApiError) setParticipantError(e.message)
    },
  })

  const removeParticipantMutation = useMutation({
    mutationFn: (participantId: number) => removeMeetingParticipant(id!, participantId),
    onSuccess: (updated) => setMeetingData(updated),
    onError: (e) => {
      if (e instanceof ApiError) alert(e.message)
    },
  })

  const addAgendaItemMutation = useMutation({
    mutationFn: (data: { text: string; speaker_id: number }) => addAgendaItem(id!, data),
    onSuccess: (updated) => { setMeetingData(updated); setShowAddItem(false); setNewItem({ text: '', speaker_id: 0 }) },
  })

  const updateAgendaItemMutation = useMutation({
    mutationFn: ({ itemId, data }: { itemId: number; data: { text: string; speaker_id: number } }) =>
      updateAgendaItem(id!, itemId, data),
    onSuccess: (updated) => { setMeetingData(updated); setEditingItemId(null) },
  })

  const deleteAgendaItemMutation = useMutation({
    mutationFn: (itemId: number) => deleteAgendaItem(id!, itemId),
    onSuccess: (updated) => setMeetingData(updated),
  })

  // DnD hooks
  const agendaDnd = useDragReorder(agendaItems, (reordered) => {
    setAgendaItems(reordered)
    agendaMutation.mutate(reordered.map(i => i.id))
  })

  const participantsDnd = useDragReorder(participants, (reordered) => {
    setParticipants(reordered)
    participantsMutation.mutate(reordered.map(p => p.id))
  })

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', year: 'numeric',
      hour: '2-digit', minute: '2-digit',
    })
  }

  function toDatetimeLocal(iso: string) {
    return new Date(iso).toISOString().slice(0, 16)
  }

  async function handleDownload(type: 'agenda' | 'participants') {
    if (!id) return
    setDownloading(type)
    try {
      if (type === 'agenda') await downloadAgenda(id)
      else await downloadParticipants(id)
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
        <h1 className="text-lg font-semibold text-gray-900 leading-snug flex-1">{meeting.title}</h1>
        <div className="flex gap-2 shrink-0">
          <button
            onClick={() => {
              setMeetingForm({ title: meeting.title, date: toDatetimeLocal(meeting.date), chairperson_id: meeting.chairperson.id })
              setEditingMeeting(true)
            }}
            className="text-xs text-gray-500 hover:text-blue-600 border rounded px-2 py-1"
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
          <p className="text-sm font-medium text-gray-700">Редактирование совещания</p>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Тема</label>
            <textarea
              value={meetingForm.title}
              onChange={e => setMeetingForm(f => ({ ...f, title: e.target.value }))}
              rows={2}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Дата и время</label>
            <input
              type="datetime-local"
              value={meetingForm.date}
              onChange={e => setMeetingForm(f => ({ ...f, date: e.target.value }))}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Председательствующий</label>
            <select
              value={meetingForm.chairperson_id}
              onChange={e => setMeetingForm(f => ({ ...f, chairperson_id: Number(e.target.value) }))}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              <option value={0}>Выберите председателя</option>
              {participants.map(p => (
                <option key={p.id} value={p.id}>{fullName(p)}</option>
              ))}
            </select>
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
              disabled={!meetingForm.title || !meetingForm.date || !meetingForm.chairperson_id || updateMeetingMutation.isPending}
              onClick={() => updateMeetingMutation.mutate({
                title: meetingForm.title,
                date: new Date(meetingForm.date).toISOString(),
                chairperson_id: meetingForm.chairperson_id,
              })}
              className="flex-1 bg-blue-600 text-white py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
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
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">Председательствующий</span>
            <span className="font-medium text-right">{fullName(meeting.chairperson)}</span>
          </div>
        </div>
      )}

      {/* Participants */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-gray-700">Участники ({participants.length})</h2>
          {participantsMutation.isPending && <span className="text-xs text-gray-400">Сохранение...</span>}
        </div>
        <div className="space-y-1">
          {participants.map((p, i) => (
            <div
              key={p.id}
              draggable
              onDragStart={() => participantsDnd.handleDragStart(i)}
              onDragOver={(e) => participantsDnd.handleDragOver(e, i)}
              onDrop={() => participantsDnd.handleDrop(i)}
              onDragEnd={participantsDnd.handleDragEnd}
              className={[
                'bg-white border rounded-lg p-3 flex items-center gap-3 transition-opacity',
                participantsDnd.dragOverIndex === i && participantsDnd.dragIndex.current !== i ? 'border-blue-400 bg-blue-50' : '',
                participantsDnd.dragIndex.current === i ? 'opacity-40' : 'opacity-100',
              ].join(' ')}
            >
              <span className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none text-lg leading-none">⠿</span>
              <span className="text-xs text-gray-400 w-5 shrink-0">{i + 1}.</span>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{fullName(p)}</p>
                {p.info && <p className="text-xs text-gray-500 mt-0.5 truncate">{p.info}</p>}
              </div>
              <button
                onClick={() => { if (confirm(`Удалить ${p.last_name} из совещания?`)) removeParticipantMutation.mutate(p.id) }}
                className="shrink-0 text-gray-300 hover:text-red-500 text-lg leading-none"
                title="Удалить из совещания"
              >
                ×
              </button>
            </div>
          ))}
        </div>

        {/* Add participant search */}
        <div className="mt-3 space-y-2">
          <input
            value={participantQuery}
            onChange={e => { setParticipantQuery(e.target.value); setParticipantError(null) }}
            placeholder="Добавить участника..."
            className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {debouncedParticipantQuery && !isSearching && searchResults.length > 0 && (
            <div className="border rounded-lg divide-y max-h-48 overflow-y-auto">
              {searchResults.map(p => {
                const alreadyIn = participants.some(mp => mp.id === p.id)
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
                        onClick={() => addParticipantMutation.mutate(p.id)}
                        disabled={addParticipantMutation.isPending}
                        className="shrink-0 ml-3 bg-blue-600 text-white px-3 py-1 rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50"
                      >
                        Добавить
                      </button>
                    )}
                  </div>
                )
              })}
            </div>
          )}
          {debouncedParticipantQuery && !isSearching && searchResults.length === 0 && (
            <p className="text-xs text-gray-500 px-1">Никого не найдено</p>
          )}
          {participantError && <p className="text-xs text-red-500 px-1">{participantError}</p>}
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
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <select
                    value={itemForm.speaker_id}
                    onChange={e => setItemForm(f => ({ ...f, speaker_id: Number(e.target.value) }))}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
                  >
                    <option value={0}>Выберите докладчика</option>
                    {participants.map(p => <option key={p.id} value={p.id}>{fullName(p)}</option>)}
                  </select>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setEditingItemId(null)}
                      className="flex-1 border text-gray-700 py-1.5 rounded-lg text-xs hover:bg-gray-50"
                    >
                      Отмена
                    </button>
                    <button
                      disabled={!itemForm.text || !itemForm.speaker_id || updateAgendaItemMutation.isPending}
                      onClick={() => updateAgendaItemMutation.mutate({ itemId: item.id, data: { text: itemForm.text, speaker_id: itemForm.speaker_id } })}
                      className="flex-1 bg-blue-600 text-white py-1.5 rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50"
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
                    agendaDnd.dragOverIndex === i && agendaDnd.dragIndex.current !== i ? 'border-blue-400 bg-blue-50' : '',
                    agendaDnd.dragIndex.current === i ? 'opacity-40' : 'opacity-100',
                  ].join(' ')}
                >
                  <span className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none text-lg leading-none mt-0.5">⠿</span>
                  <span className="text-xs text-gray-400 w-5 shrink-0 mt-0.5">{i + 1}.</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium">{item.text}</p>
                    <p className="text-xs text-gray-500 mt-1">
                      Докладчик: {fullName(item.speaker)}
                    </p>
                  </div>
                  <div className="flex gap-1 shrink-0">
                    <button
                      onClick={() => { setEditingItemId(item.id); setItemForm({ text: item.text, speaker_id: item.speaker.id }) }}
                      className="text-xs text-gray-400 hover:text-blue-600 px-1.5 py-1 rounded"
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
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <select
              value={newItem.speaker_id}
              onChange={e => setNewItem(f => ({ ...f, speaker_id: Number(e.target.value) }))}
              className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              <option value={0}>Выберите докладчика</option>
              {participants.map(p => <option key={p.id} value={p.id}>{fullName(p)}</option>)}
            </select>
            <div className="flex gap-2">
              <button
                onClick={() => { setShowAddItem(false); setNewItem({ text: '', speaker_id: 0 }) }}
                className="flex-1 border text-gray-700 py-1.5 rounded-lg text-xs hover:bg-gray-50"
              >
                Отмена
              </button>
              <button
                disabled={!newItem.text || !newItem.speaker_id || addAgendaItemMutation.isPending}
                onClick={() => addAgendaItemMutation.mutate({ text: newItem.text, speaker_id: newItem.speaker_id })}
                className="flex-1 bg-blue-600 text-white py-1.5 rounded-lg text-xs font-medium hover:bg-blue-700 disabled:opacity-50"
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
