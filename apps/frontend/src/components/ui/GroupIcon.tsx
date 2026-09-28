import { useEffect, useState } from 'react'
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

// Entry icon: a custom uploaded icon or resolved favicon, with a graceful
// fallback to the group icon when there is none or the image fails to load.
export function EntryIcon({
  iconUrl,
  faviconUrl,
  groupIcon,
  size = 20,
}: {
  iconUrl?: string | null
  faviconUrl?: string | null
  groupIcon?: string
  size?: number
}) {
  const src = iconUrl || faviconUrl || null
  const [failed, setFailed] = useState(false)

  // The icon fills the whole box (no padded background), so it reads larger.
  const box = size + 12

  // Reset the error state whenever the source changes.
  useEffect(() => setFailed(false), [src])

  return (
    <span
      className="grid place-items-center text-accent overflow-hidden shrink-0"
      style={{ width: box, height: box }}
    >
      {src && !failed ? (
        <img
          src={src}
          alt=""
          width={box}
          height={box}
          className="object-contain w-full h-full"
          onError={() => setFailed(true)}
        />
      ) : (
        <GroupIcon icon={groupIcon} size={box} />
      )}
    </span>
  )
}
