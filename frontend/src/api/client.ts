import axios from 'axios'
import { getDefaultStore } from 'jotai'
import { setTokenAtom } from '@/stores/auth'

const store = getDefaultStore()

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
  withCredentials: true, // Automatically send Cookie
})

// Response interceptor: handle 401 errors
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      store.set(setTokenAtom, null)
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
