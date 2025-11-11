import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/hooks/useAuth'
import { authApi } from '@/api/auth'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { LogOut, Info } from 'lucide-react'

export function SettingsPage() {
  const { logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    try {
      await authApi.logout()
      logout()
      toast.success('Logged out successfully')
      navigate('/login')
    } catch (error) {
      console.error('Logout failed:', error)
      // Even if API call fails, clear local state
      logout()
      navigate('/login')
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Settings</h1>

      {/* Account Section */}
      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            You are currently logged in to ViBox.
          </p>
          <Button
            onClick={handleLogout}
            variant="outline"
            className="gap-2"
          >
            <LogOut className="h-4 w-4" />
            Logout
          </Button>
        </CardContent>
      </Card>

      {/* About Section */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            <CardTitle>About</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="space-y-1">
            <p className="text-sm font-medium">ViBox v1.2.0</p>
            <p className="text-xs text-muted-foreground">
              Container-based workspace management platform
            </p>
          </div>
          <div className="border-t pt-3 space-y-1">
            <p className="text-xs text-muted-foreground">
              <span className="font-medium">Backend:</span> Go 1.25
            </p>
            <p className="text-xs text-muted-foreground">
              <span className="font-medium">Frontend:</span> React 19 + Vite 7
            </p>
            <p className="text-xs text-muted-foreground">
              <span className="font-medium">Terminal:</span> xterm.js 5.5
            </p>
          </div>
          <div className="border-t pt-3">
            <p className="text-xs text-muted-foreground">
              Built with ❤️ using modern web technologies
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
