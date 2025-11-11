import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { LoginPage } from './pages/LoginPage'
import { WorkspacesPage } from './pages/WorkspacesPage'
import { WorkspaceDetailPage } from './pages/WorkspaceDetailPage'
import { SettingsPage } from './pages/SettingsPage'
import { Layout } from './components/layout/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'

const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <Layout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <WorkspacesPage />,
      },
      {
        path: 'workspace/:id',
        element: <WorkspaceDetailPage />,
      },
      {
        path: 'settings',
        element: <SettingsPage />,
      },
    ],
  },
])

function App() {
  return <RouterProvider router={router} />
}

export default App
