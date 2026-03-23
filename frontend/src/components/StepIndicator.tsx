interface Props {
  current: number
  total: number
  labels: string[]
}

export function StepIndicator({ current, total, labels }: Props) {
  return (
    <div className="flex items-center gap-1 mb-6">
      {labels.map((label, i) => {
        const step = i + 1
        const active = step === current
        const done = step < current
        return (
          <div key={step} className="flex items-center gap-1">
            <div className={`flex items-center justify-center w-7 h-7 rounded-full text-sm font-medium
              ${done ? 'bg-green-500 text-white' : active ? 'bg-green-600 text-white' : 'bg-gray-200 text-gray-500'}`}>
              {done ? '✓' : step}
            </div>
            <span className={`text-xs hidden sm:block ${active ? 'text-green-600 font-medium' : 'text-gray-400'}`}>
              {label}
            </span>
            {step < total && <div className="w-4 h-px bg-gray-300 mx-1" />}
          </div>
        )
      })}
    </div>
  )
}
