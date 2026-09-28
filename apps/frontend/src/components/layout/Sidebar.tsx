import { NavLink, useLocation } from 'react-router-dom'
import {
  KeyRound,
  User as UserIcon,
  Building2,
  Boxes,
  ShieldCheck,
  Puzzle,
  List,
} from 'lucide-react'
import type { ReactNode } from 'react'

type Item = { to: string; label: string; icon: ReactNode; end?: boolean; match?: (search: string) => boolean }

function NavItem({ item }: { item: Item }) {
  const location = useLocation()
  const active =
    item.match && location.pathname === item.to.split('?')[0]
      ? item.match(location.search)
      : undefined

  return (
    <NavLink
      to={item.to}
      end={item.end}
      className={({ isActive }) => {
        const on = active !== undefined ? active : isActive
        return (
          'flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg text-[14px] transition-colors ' +
          (on
            ? 'bg-accent-50 text-accent-text font-semibold shadow-[inset_0_0_0_1px_theme(colors.accent.100)]'
            : 'text-ink hover:bg-ink-strong/5 hover:text-ink-strong')
        )
      }}
    >
      <span className="shrink-0">{item.icon}</span>
      {item.label}
    </NavLink>
  )
}

function Section({ title, items }: { title: string; items: Item[] }) {
  return (
    <div className="mb-5">
      <div className="px-2.5 pb-1.5 text-[11px] font-semibold uppercase tracking-[.08em] text-ink-faint">
        {title}
      </div>
      <nav className="flex flex-col gap-0.5">
        {items.map((it) => (
          <NavItem key={it.to} item={it} />
        ))}
      </nav>
    </div>
  )
}

export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const passwordItems: Item[] = [
    {
      to: '/passwords',
      label: 'Все пароли',
      icon: <KeyRound size={16} />,
      end: true,
      match: (s) => !s.includes('scope='),
    },
    {
      to: '/passwords?scope=personal',
      label: 'Личные',
      icon: <UserIcon size={16} />,
      match: (s) => s.includes('scope=personal'),
    },
    {
      to: '/passwords?scope=commercial',
      label: 'Коммерческие',
      icon: <Building2 size={16} />,
      match: (s) => s.includes('scope=commercial'),
    },
    { to: '/companies', label: 'Компании', icon: <Building2 size={16} /> },
  ]

  const secretItems: Item[] = [
    { to: '/app-secrets', label: 'Проекты', icon: <Boxes size={16} /> },
  ]

  const settingsItems: Item[] = [
    { to: '/settings/security', label: 'Безопасность', icon: <ShieldCheck size={16} /> },
    { to: '/settings/extension', label: 'Chrome extension', icon: <Puzzle size={16} /> },
    { to: '/settings/audit', label: 'Журнал действий', icon: <List size={16} /> },
  ]

  return (
    <div className="h-full flex flex-col p-4 pt-5" onClick={onNavigate}>
      <Section title="Пароли" items={passwordItems} />
      <Section title="Секреты приложений" items={secretItems} />
      <Section title="Настройки" items={settingsItems} />
    </div>
  )
}
