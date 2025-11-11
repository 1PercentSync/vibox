import { atom } from 'jotai'

export const tokenAtom = atom<string | null>(
  localStorage.getItem('api_token')
)

export const isAuthenticatedAtom = atom(
  (get) => get(tokenAtom) !== null
)

export const setTokenAtom = atom(
  null,
  (_get, set, newToken: string | null) => {
    set(tokenAtom, newToken)
    if (newToken) {
      localStorage.setItem('api_token', newToken)
    } else {
      localStorage.removeItem('api_token')
    }
  }
)
