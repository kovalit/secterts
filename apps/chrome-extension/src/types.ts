export type Settings = {
  apiBaseUrl: string
  extensionToken: string
}

export type LookupItem = {
  id: string
  title: string
  login?: string | null
  favicon_url?: string | null
  scope?: string
}

export type LookupResponse = {
  domain: string
  items: LookupItem[]
}
