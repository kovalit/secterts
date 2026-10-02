import {
  KeyRound,
  Ticket,
  IdCard,
  Server,
  Database,
  SquareTerminal,
  FileBadge,
  Variable,
  Lock,
  type LucideIcon,
} from 'lucide-react'
import type { SecretNoteType } from '../../types'

export type NoteTypeMeta = {
  type: SecretNoteType
  label: string
  icon: LucideIcon
}

// Each secret-note type carries a universal icon. Order drives the selector.
export const NOTE_TYPES: NoteTypeMeta[] = [
  { type: 'api_key', label: 'API-ключ', icon: KeyRound },
  { type: 'token', label: 'Токен', icon: Ticket },
  { type: 'credentials', label: 'Учётные данные (ID/логин + секрет)', icon: IdCard },
  { type: 'server', label: 'Доступ к серверу / VPS', icon: Server },
  { type: 'database', label: 'Доступ к БД', icon: Database },
  { type: 'ssh_key', label: 'SSH-ключ', icon: SquareTerminal },
  { type: 'certificate', label: 'Сертификат / private key', icon: FileBadge },
  { type: 'environment', label: 'Переменные окружения', icon: Variable },
  { type: 'generic', label: 'Произвольный секрет', icon: Lock },
]

const BY_TYPE = new Map(NOTE_TYPES.map((t) => [t.type, t]))

export function noteTypeMeta(type: SecretNoteType | string): NoteTypeMeta {
  return BY_TYPE.get(type as SecretNoteType) ?? NOTE_TYPES[NOTE_TYPES.length - 1]
}
