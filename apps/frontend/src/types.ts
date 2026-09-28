// Shared domain types (mirrors the backend API responses and docs/03).

export type PasswordScope = 'personal' | 'commercial'

export type User = {
  id: string
  email: string
  email_verified: boolean
  email_2fa_enabled: boolean
  created_at: string
}

export type PasswordGroup = {
  id: string
  slug: string
  label: string
  icon: string
  sort_order: number
}

export type Company = {
  id: string
  name: string
  created_at: string
  updated_at: string
}

export type PasswordEntry = {
  id: string
  scope: PasswordScope
  company_id?: string | null
  group_id: string
  title: string
  site_url?: string | null
  domain?: string | null
  favicon_url?: string | null
  icon_source: 'favicon' | 'group' | 'custom'
  custom_icon?: string | null
  login?: string | null
  has_password: boolean
  has_comment: boolean
  created_at: string
  updated_at: string
}

export type AppProject = {
  id: string
  company_id?: string | null
  name: string
  description?: string | null
  env_count: number
  secret_count: number
  created_at: string
  updated_at: string
}

export type AppEnvironment = {
  id: string
  project_id: string
  name: string
  sort_order: number
}

export type AppSecret = {
  id: string
  project_id: string
  environment_id: string
  key: string
  has_value: boolean
  has_comment: boolean
  created_at: string
  updated_at: string
}

export type ExtensionToken = {
  id: string
  name: string
  scopes: string[]
  created_at: string
  expires_at: string
  last_used_at?: string | null
  revoked_at?: string | null
}

export type Session = {
  id: string
  user_agent?: string | null
  ip?: string | null
  current: boolean
  created_at: string
  expires_at: string
  revoked_at?: string | null
}

export type AuditLog = {
  id: string
  action: string
  entity_type: string
  entity_id?: string | null
  ip?: string | null
  created_at: string
}
