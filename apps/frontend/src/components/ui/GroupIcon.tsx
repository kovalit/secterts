import {
  Code,
  Folder,
  Mail,
  Server,
  Users,
  type LucideIcon,
} from 'lucide-react'

// Map password-group icon slugs (from the DB) to Lucide icons.
const ICONS: Record<string, LucideIcon> = {
  mail: Mail,
  users: Users,
  server: Server,
  code: Code,
}

export function GroupIcon({ icon, size = 16 }: { icon?: string; size?: number }) {
  const Cmp = (icon && ICONS[icon]) || Folder
  return <Cmp size={size} />
}

// Favicon with graceful fallback to the group icon.
export function EntryIcon({
  faviconUrl,
  groupIcon,
  size = 20,
}: {
  faviconUrl?: string | null
  groupIcon?: string
  size?: number
}) {
  return (
    <span
      className="grid place-items-center rounded-md bg-accent-50 text-accent overflow-hidden shrink-0"
      style={{ width: size + 12, height: size + 12 }}
    >
      {faviconUrl ? (
        <img
          src={faviconUrl}
          alt=""
          width={size}
          height={size}
          onError={(e) => {
            // Hide broken favicons; the group icon underneath shows instead.
            ;(e.currentTarget as HTMLImageElement).style.display = 'none'
          }}
        />
      ) : (
        <GroupIcon icon={groupIcon} size={size - 2} />
      )}
    </span>
  )
}
