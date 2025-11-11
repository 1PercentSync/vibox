import { Link } from 'react-router-dom'

export function Header() {
  return (
    <header className="border-b">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold">
          ViBox
        </Link>
        <nav className="flex gap-4">
          <Link to="/" className="hover:underline">
            Workspaces
          </Link>
          <Link to="/settings" className="hover:underline">
            Settings
          </Link>
        </nav>
      </div>
    </header>
  )
}
