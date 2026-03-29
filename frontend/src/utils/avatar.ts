export const DEFAULT_AVATAR_COLORS = [
  '#6366f1',
  '#8b5cf6',
  '#ec4899',
  '#f59e0b',
  '#10b981',
  '#3b82f6',
  '#ef4444',
  '#14b8a6',
  '#2a4b7c',
  '#e2725b',
  '#5b7c7c',
  '#d4a373'
]

export function hashAvatarColor(seed: string): string {
  let hash = 0
  for (let i = 0; i < seed.length; i++) hash = seed.charCodeAt(i) + ((hash << 5) - hash)
  return DEFAULT_AVATAR_COLORS[Math.abs(hash) % DEFAULT_AVATAR_COLORS.length]
}

/** Sunucudan gelen #RRGGBB veya boş — boşsa seed ile paletten renk */
export function resolveAvatarBackground(saved: string | undefined | null, fallbackSeed: string): string {
  const s = saved?.trim()
  if (s && /^#[0-9A-Fa-f]{6}$/.test(s)) return s
  return hashAvatarColor(fallbackSeed || 'x')
}

export function usernameInitial(username: string | undefined | null): string {
  const u = String(username ?? '').trim()
  if (!u) return '?'
  const ch = u.match(/^./u)?.[0] ?? u[0]
  return ch.toUpperCase()
}

export function absoluteApiUrl(apiBase: string, path: string | undefined | null): string {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return `${apiBase.replace(/\/$/, '')}${path.startsWith('/') ? path : '/' + path}`
}
