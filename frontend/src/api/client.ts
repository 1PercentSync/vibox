import axios from 'axios'
import { flushSync } from 'react-dom'
import { getDefaultStore } from 'jotai'
import { setTokenAtom } from '@/stores/auth'
import { toast } from 'sonner'

const store = getDefaultStore()

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
  withCredentials: true, // Automatically send Cookie
})

// Response interceptor: handle errors globally
client.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle 401 Unauthorized
    if (error.response?.status === 401) {
      // Use flushSync to ensure synchronous state update
      // Prevents race condition when component unmounts before state update
      flushSync(() => {
        store.set(setTokenAtom, null)
      })
      toast.error('Session expired. Please login again.')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    // Handle other HTTP errors
    const errorMessage = error.response?.data?.error || error.message || 'An unexpected error occurred'

    // Don't show toast for cancelled requests
    if (!axios.isCancel(error)) {
      // Show different messages based on status code
      if (error.response?.status === 404) {
        toast.error('Resource not found')
      } else if (error.response?.status === 500) {
        toast.error(`Server error: ${errorMessage}`)
      } else if (error.code === 'ECONNABORTED') {
        toast.error('Request timeout. Please try again.')
      } else if (error.code === 'ERR_NETWORK') {
        toast.error('Network error. Please check your connection.')
      } else {
        toast.error(errorMessage)
      }
    }

    return Promise.reject(error)
  }
)

export default client
