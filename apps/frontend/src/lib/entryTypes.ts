import type { EntryType } from '../types'

// Human-readable (Russian) labels for each entry type, plus a short plural used
// on the inventory screen ("17 API-ключей").
export const ENTRY_TYPES: { value: EntryType; label: string; plural: string }[] = [
  { value: 'password', label: 'Пароль', plural: 'Пароли' },
  { value: 'api_key', label: 'API-ключ', plural: 'API-ключи' },
  { value: 'ssh_key', label: 'SSH-ключ', plural: 'SSH-ключи' },
  { value: 'certificate', label: 'Сертификат', plural: 'Сертификаты' },
  { value: 'token', label: 'Токен', plural: 'Токены' },
  { value: 'license', label: 'Лицензия', plural: 'Лицензии' },
  { value: 'database', label: 'Доступ к БД', plural: 'Доступы к БД' },
  { value: 'secret_note', label: 'Секретная заметка', plural: 'Секретные заметки' },
  { value: 'other', label: 'Другое', plural: 'Другое' },
]

const BY_VALUE = new Map(ENTRY_TYPES.map((t) => [t.value, t]))

export function entryTypeLabel(value?: string | null): string {
  if (!value) return 'Пароль'
  return BY_VALUE.get(value as EntryType)?.label ?? value
}

export function entryTypePlural(value?: string | null): string {
  if (!value) return 'Пароли'
  return BY_VALUE.get(value as EntryType)?.plural ?? value
}

// Entry types for which an expiration date is meaningful (feature 2).
export const EXPIRABLE_TYPES: EntryType[] = [
  'api_key',
  'certificate',
  'token',
  'license',
  'password',
]

// Strength labels for the 0..4 score computed by the backend.
export const STRENGTH_LABELS = ['Очень слабый', 'Слабый', 'Средний', 'Надёжный', 'Очень надёжный']

export function strengthLabel(score?: number | null): string {
  if (score == null) return 'Неизвестно'
  return STRENGTH_LABELS[Math.max(0, Math.min(4, score))]
}
