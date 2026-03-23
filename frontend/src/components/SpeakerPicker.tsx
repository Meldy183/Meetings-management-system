import { useRef, useState } from 'react'
import type { Person } from '../api/types'

function fullName(p: Person) {
  return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ')
}

export function SpeakerPicker({
  people,
  speakerIds,
  onChange,
}: {
  people: Person[]
  speakerIds: number[]
  onChange: (ids: number[]) => void
}) {
  const selected = speakerIds.map(id => people.find(p => p.id === id)!).filter(Boolean)
  const available = people.filter(p => !speakerIds.includes(p.id))

  const dragIndex = useRef<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  function handleDragStart(i: number) { dragIndex.current = i }
  function handleDragOver(e: React.DragEvent, i: number) { e.preventDefault(); setDragOverIndex(i) }
  function handleDragEnd() { dragIndex.current = null; setDragOverIndex(null) }
  function handleDrop(toIndex: number) {
    const from = dragIndex.current
    dragIndex.current = null
    setDragOverIndex(null)
    if (from === null || from === toIndex) return
    const next = [...speakerIds]
    const [moved] = next.splice(from, 1)
    next.splice(toIndex, 0, moved)
    onChange(next)
  }

  return (
    <div className="space-y-2">
      {selected.length > 0 && (
        <div className="space-y-1">
          {selected.map((p, i) => (
            <div
              key={p.id}
              draggable
              onDragStart={() => handleDragStart(i)}
              onDragOver={(e) => handleDragOver(e, i)}
              onDrop={() => handleDrop(i)}
              onDragEnd={handleDragEnd}
              className={[
                'flex items-center gap-2 px-2 py-1.5 bg-white border rounded-lg text-sm transition-opacity',
                dragOverIndex === i && dragIndex.current !== i ? 'border-blue-400 bg-blue-50' : '',
                dragIndex.current === i ? 'opacity-40' : '',
              ].join(' ')}
            >
              <span className="text-gray-300 hover:text-gray-500 cursor-grab active:cursor-grabbing select-none">⠿</span>
              <span className="flex-1 truncate">{fullName(p)}</span>
              {p.info && <span className="text-xs text-gray-400 truncate max-w-[40%]">{p.info}</span>}
              <button
                type="button"
                onClick={() => onChange(speakerIds.filter(id => id !== p.id))}
                className="text-red-400 hover:text-red-600 text-lg leading-none ml-1"
              >×</button>
            </div>
          ))}
        </div>
      )}
      {available.length > 0 && (
        <div className="border rounded-lg divide-y max-h-40 overflow-y-auto">
          {available.map(p => (
            <div
              key={p.id}
              onClick={() => onChange([...speakerIds, p.id])}
              className="flex items-center px-3 py-2 bg-white hover:bg-gray-50 cursor-pointer text-sm"
            >
              <span className="flex-1 truncate">{fullName(p)}</span>
              {p.info && <span className="text-xs text-gray-400 ml-2 truncate max-w-[40%]">{p.info}</span>}
            </div>
          ))}
        </div>
      )}
      {people.length === 0 && (
        <p className="text-xs text-gray-400 italic">Сначала добавьте участников совещания</p>
      )}
    </div>
  )
}
