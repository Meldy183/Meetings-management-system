import { useForm } from 'react-hook-form'
import type { Participant } from '../api/types'

interface AgendaItemData {
  text: string
  speaker_id: number
}

interface Props {
  index: number
  participants: Participant[]
  defaultValues?: Partial<AgendaItemData>
  onChange: (data: AgendaItemData) => void
  onRemove: () => void
}

export function AgendaItemRow({ index, participants, defaultValues, onChange, onRemove }: Props) {
  const { register, watch } = useForm<AgendaItemData>({
    defaultValues: { speaker_id: defaultValues?.speaker_id, text: defaultValues?.text ?? '' },
  })

  // Call onChange on every change
  const text = watch('text')
  const speakerId = watch('speaker_id')

  return (
    <div className="flex gap-2 items-start p-3 bg-white border rounded-lg">
      <span className="text-sm font-medium text-gray-400 mt-2 w-5 shrink-0">{index + 1}.</span>
      <div className="flex-1 space-y-2">
        <input
          {...register('text', { onChange: () => onChange({ text, speaker_id: Number(speakerId) }) })}
          placeholder="Тема пункта повестки"
          className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          {...register('speaker_id', { valueAsNumber: true, onChange: () => onChange({ text, speaker_id: Number(speakerId) }) })}
          className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
        >
          <option value="">Выберите докладчика *</option>
          {participants.map(p => (
            <option key={p.id} value={p.id}>
              {p.last_name} {p.first_name} {p.middle_name ?? ''}
            </option>
          ))}
        </select>
      </div>
      <button onClick={onRemove} className="text-gray-400 hover:text-red-500 mt-2 text-lg leading-none">×</button>
    </div>
  )
}
