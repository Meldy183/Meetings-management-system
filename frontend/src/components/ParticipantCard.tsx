import type { Participant } from '../api/types'

interface Props {
  participant: Participant
  onRemove?: () => void
  onEdit?: () => void
  badge?: string // e.g. "Председатель"
}

export function ParticipantCard({ participant, onRemove, onEdit, badge }: Props) {
  const fullName = [participant.last_name, participant.first_name, participant.middle_name]
    .filter(Boolean).join(' ')
  return (
    <div className="flex items-center justify-between p-3 bg-white border rounded-lg">
      <div>
        <div className="flex items-center gap-2">
          <span className="font-medium text-sm">{fullName}</span>
          {badge && <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full">{badge}</span>}
        </div>
        {participant.info && <p className="text-xs text-gray-500 mt-0.5">{participant.info}</p>}
      </div>
      <div className="flex gap-2">
        {onEdit && (
          <button onClick={onEdit} className="text-xs text-gray-500 hover:text-blue-600 px-2 py-1 rounded">
            Изменить
          </button>
        )}
        {onRemove && (
          <button onClick={onRemove} className="text-xs text-gray-500 hover:text-red-600 px-2 py-1 rounded">
            Удалить
          </button>
        )}
      </div>
    </div>
  )
}
