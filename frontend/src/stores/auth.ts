import { atom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

// Use atomWithStorage to automatically handle SSR hydration
export const tokenAtom = atomWithStorage<string | null>('api_token', null)

export const isAuthenticatedAtom = atom(
  (get) => get(tokenAtom) !== null
)

// Remove manual localStorage operations - Jotai handles it automatically
export const setTokenAtom = atom(
  null,
  (_get, set, newToken: string | null) => {
    set(tokenAtom, newToken)
  }
)
