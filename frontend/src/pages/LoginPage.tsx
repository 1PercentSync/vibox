import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAtom } from 'jotai'
import { setTokenAtom } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card'
import { toast } from 'sonner'

export function LoginPage() {
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [, saveToken] = useAtom(setTokenAtom)
  const navigate = useNavigate()

  const handleLogin = async () => {
    // Validate input
    if (!token.trim()) {
      setError('Token is required')
      return
    }

    setLoading(true)
    setError('')

    try {
      // Call login API
      await authApi.login(token)
      // Save token to state (also saves to localStorage)
      saveToken(token)
      toast.success('Login successful!')
      // Redirect to home page
      navigate('/')
    } catch (err: any) {
      // Handle error
      if (err.response?.status === 401) {
        setError('Invalid token')
      } else {
        setError('Login failed. Please try again.')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !loading) {
      handleLogin()
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>ViBox Login</CardTitle>
          <CardDescription>
            Enter your API token to access ViBox
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="token" className="text-sm font-medium">
                API Token
              </label>
              <Input
                id="token"
                type="password"
                placeholder="Enter your API token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                onKeyDown={handleKeyDown}
                disabled={loading}
              />
            </div>
            {error && (
              <p className="text-sm text-destructive" role="alert">
                {error}
              </p>
            )}
          </div>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <Button
            onClick={handleLogin}
            disabled={loading}
            className="w-full"
          >
            {loading ? 'Logging in...' : 'Login'}
          </Button>
          <div className="text-sm text-muted-foreground">
            <p className="mb-1">Generate a token with:</p>
            <code className="bg-muted px-2 py-1 rounded text-xs">
              openssl rand -hex 32
            </code>
          </div>
        </CardFooter>
      </Card>
    </div>
  )
}
