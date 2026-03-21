import { useForm } from 'react-hook-form'
import type { PersonCreate } from '../api/types'

interface Props {
  defaultValues?: Partial<PersonCreate>
  onSubmit: (data: PersonCreate) => void
  onCancel?: () => void
  submitLabel?: string
  isLoading?: boolean
}

export function ParticipantForm({ defaultValues, onSubmit, onCancel, submitLabel = 'Сохранить', isLoading }: Props) {
  const { register, handleSubmit, formState: { errors } } = useForm<PersonCreate>({ defaultValues })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">Фамилия *</label>
          <input
            {...register('last_name', { required: 'Обязательное поле' })}
            className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Иванов"
          />
          {errors.last_name && <p className="text-xs text-red-500 mt-1">{errors.last_name.message}</p>}
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">Имя *</label>
          <input
            {...register('first_name', { required: 'Обязательное поле' })}
            className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Иван"
          />
          {errors.first_name && <p className="text-xs text-red-500 mt-1">{errors.first_name.message}</p>}
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700 mb-1">Отчество</label>
          <input
            {...register('middle_name')}
            className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Иванович"
          />
        </div>
      </div>
      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">Должность / роль</label>
        <input
          {...register('info')}
          className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Руководитель отдела"
        />
      </div>
      <div className="flex gap-3 pt-1">
        <button
          type="submit"
          disabled={isLoading}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
        >
          {isLoading ? 'Сохранение...' : submitLabel}
        </button>
        {onCancel && (
          <button type="button" onClick={onCancel} className="text-gray-500 px-4 py-2 rounded-lg text-sm hover:text-gray-700">
            Отмена
          </button>
        )}
      </div>
    </form>
  )
}
