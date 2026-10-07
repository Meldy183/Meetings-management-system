import type { AttendanceMode } from '../api/types'

interface Props {
  value: AttendanceMode
  onChange: (value: AttendanceMode) => void
  disabled?: boolean
}

export function AttendanceModeToggle({ value, onChange, disabled = false }: Props) {
  return (
    <div role="group" aria-label="Формат участия" className="inline-flex shrink-0 rounded-md border border-gray-300 p-0.5"
      onPointerDown={event => event.stopPropagation()}
      onDragStart={event => { event.preventDefault(); event.stopPropagation() }}>
      {(['in_person', 'vcs'] as const).map(mode => (
        <button key={mode} type="button" disabled={disabled} aria-pressed={value === mode}
          onClick={event => { event.stopPropagation(); onChange(mode) }}
          className={`rounded px-2 py-1 text-xs ${value === mode ? 'bg-green-600 text-white' : 'text-gray-600 hover:bg-gray-100'} disabled:opacity-50`}>
          {mode === 'in_person' ? 'Очно' : 'ВКС'}
        </button>
      ))}
    </div>
  )
}
