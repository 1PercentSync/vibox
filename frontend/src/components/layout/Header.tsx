import { Link } from 'react-router-dom'

export function Header() {
  return (
    <header className="border-b border-border bg-background/60 backdrop-blur-md backdrop-saturate-150 sticky top-0 z-50 shadow-sm">
      <div className="container mx-auto px-6 py-4 flex items-center justify-between max-w-7xl">
        <Link to="/" className="text-xl font-bold hover:opacity-80 transition-opacity">
          ViBox
        </Link>
        <nav className="flex gap-6">
          <Link to="/" className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors">
            Workspaces
          </Link>
          <Link to="/settings" className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors">
            Settings
          </Link>
        </nav>
      </div>
    </header>
  )
}
