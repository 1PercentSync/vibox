// Authentication Hook
import { useAtom, useAtomValue } from 'jotai'
import { tokenAtom, setTokenAtom, isAuthenticatedAtom } from '@/stores/auth'

export const useAuth = () => {
  const [token] = useAtom(tokenAtom)
  const isAuthenticated = useAtomValue(isAuthenticatedAtom)
  const [, setToken] = useAtom(setTokenAtom)

  const logout = () => {
    setToken(null)
  }

  return {
    token,
    isAuthenticated,
    setToken,
    logout,
  }
}
