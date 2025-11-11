// Workspaces Hook with auto-refresh functionality
import { useEffect } from 'react'
import { useAtom, useSetAtom } from 'jotai'
import { workspacesAtom } from '@/stores/workspaces'
import { isLoadingAtom } from '@/stores/ui'

// This hook will be enhanced in Module 3 (API Integration)
// For now, it provides the basic structure
export const useWorkspaces = (options?: { autoRefresh?: boolean; interval?: number }) => {
  const [workspaces, setWorkspaces] = useAtom(workspacesAtom)
  const setIsLoading = useSetAtom(isLoadingAtom)

  const { autoRefresh = true, interval = 5000 } = options || {}

  // Fetch workspaces function (will be implemented in Module 3)
  const fetchWorkspaces = async () => {
    try {
      setIsLoading(true)
      // TODO: Implement API call in Module 3
      // const { data } = await workspaceApi.list()
      // setWorkspaces(data)
      console.log('Fetching workspaces...')
    } catch (error) {
      console.error('Failed to fetch workspaces:', error)
    } finally {
      setIsLoading(false)
    }
  }

  // Auto-refresh with polling
  useEffect(() => {
    if (!autoRefresh) return

    // Initial fetch
    fetchWorkspaces()

    // Set up interval for polling
    const intervalId = setInterval(fetchWorkspaces, interval)

    // Cleanup on unmount
    return () => {
      clearInterval(intervalId)
    }
  }, [autoRefresh, interval])

  return {
    workspaces,
    refetch: fetchWorkspaces,
    setWorkspaces,
  }
}
