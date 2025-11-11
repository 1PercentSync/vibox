import { useParams } from 'react-router-dom'

export function WorkspaceDetailPage() {
  const { id } = useParams()

  return (
    <div className="container mx-auto py-6">
      <h1 className="text-3xl font-bold">Workspace Detail: {id}</h1>
    </div>
  )
}
