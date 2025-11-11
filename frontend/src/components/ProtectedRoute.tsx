import { Navigate } from 'react-router-dom'
import { useAtom } from 'jotai'
import { isAuthenticatedAtom } from '@/stores/auth'

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const [isAuthenticated] = useAtom(isAuthenticatedAtom)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
