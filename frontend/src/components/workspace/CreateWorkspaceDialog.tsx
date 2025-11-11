import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { Script } from '@/api/types'

interface CreateWorkspaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: {
    name: string
    image: string
    scripts: Script[]
    ports: Record<string, string>
  }) => Promise<void>
}

const commonImages = [
  'ubuntu:22.04',
  'ubuntu:24.04',
  'alpine:latest',
  'node:20',
  'python:3.11',
  'golang:1.22',
]

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
  onSubmit,
}: CreateWorkspaceDialogProps) {
  const [name, setName] = useState('')
  const [image, setImage] = useState('ubuntu:22.04')
  const [scripts, setScripts] = useState<{ name: string; content: string }[]>([])
  const [ports, setPorts] = useState<{ port: string; label: string }[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleAddScript = () => {
    setScripts([...scripts, { name: '', content: '' }])
  }

  const handleRemoveScript = (index: number) => {
    setScripts(scripts.filter((_, i) => i !== index))
  }

  const handleScriptChange = (
    index: number,
    field: 'name' | 'content',
    value: string
  ) => {
    const updated = [...scripts]
    updated[index][field] = value
    setScripts(updated)
  }

  const handleAddPort = () => {
    setPorts([...ports, { port: '', label: '' }])
  }

  const handleRemovePort = (index: number) => {
    setPorts(ports.filter((_, i) => i !== index))
  }

  const handlePortChange = (
    index: number,
    field: 'port' | 'label',
    value: string
  ) => {
    const updated = [...ports]
    updated[index][field] = value
    setPorts(updated)
  }

  const handleSubmit = async () => {
    // Validation
    if (!name.trim()) {
      setError('Workspace name is required')
      return
    }

    // Validate scripts
    const validScripts: Script[] = scripts
      .filter((s) => s.name.trim() && s.content.trim())
      .map((s, index) => ({
        name: s.name.trim(),
        content: s.content.trim(),
        order: index + 1,
      }))

    // Validate ports
    const validPorts: Record<string, string> = {}
    for (const p of ports) {
      if (p.port.trim() && p.label.trim()) {
        const portNum = parseInt(p.port.trim())
        if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
          setError(`Invalid port number: ${p.port}`)
          return
        }
        validPorts[p.port.trim()] = p.label.trim()
      }
    }

    try {
      setLoading(true)
      setError('')
      await onSubmit({
        name: name.trim(),
        image,
        scripts: validScripts,
        ports: validPorts,
      })

      // Reset form
      setName('')
      setImage('ubuntu:22.04')
      setScripts([])
      setPorts([])
      onOpenChange(false)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create workspace')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Workspace</DialogTitle>
          <DialogDescription>
            Create a new containerized development environment
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Workspace Name */}
          <div className="space-y-2">
            <Label htmlFor="name">
              Workspace Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="name"
              placeholder="my-workspace"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          {/* Docker Image */}
          <div className="space-y-2">
            <Label htmlFor="image">Docker Image</Label>
            <Input
              id="image"
              list="common-images"
              placeholder="ubuntu:22.04"
              value={image}
              onChange={(e) => setImage(e.target.value)}
            />
            <datalist id="common-images">
              {commonImages.map((img) => (
                <option key={img} value={img} />
              ))}
            </datalist>
            <p className="text-xs text-muted-foreground">
              Enter any Docker image (e.g., ubuntu:rolling, node:20-alpine). Common images appear as suggestions.
            </p>
          </div>

          {/* Initialization Scripts */}
          <div className="space-y-2">
            <Label>Initialization Scripts (Optional)</Label>
            {scripts.map((script, index) => (
              <div key={index} className="space-y-2 p-3 border rounded-lg">
                <div className="flex items-center gap-2">
                  <Input
                    placeholder="Script name"
                    value={script.name}
                    onChange={(e) =>
                      handleScriptChange(index, 'name', e.target.value)
                    }
                  />
                  <Button
                    type="button"
                    size="sm"
                    variant="destructive"
                    onClick={() => handleRemoveScript(index)}
                  >
                    Remove
                  </Button>
                </div>
                <Textarea
                  placeholder="#!/bin/bash&#10;apt-get update && apt-get install -y curl"
                  value={script.content}
                  onChange={(e) =>
                    handleScriptChange(index, 'content', e.target.value)
                  }
                  rows={4}
                />
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddScript}
            >
              + Add Script
            </Button>
          </div>

          {/* Port Labels */}
          <div className="space-y-2">
            <Label>Port Labels (Optional)</Label>
            <p className="text-sm text-muted-foreground">
              Give friendly names to ports for quick access
            </p>
            {ports.map((port, index) => (
              <div key={index} className="flex items-center gap-2">
                <Input
                  placeholder="Port (e.g., 8080)"
                  value={port.port}
                  onChange={(e) =>
                    handlePortChange(index, 'port', e.target.value)
                  }
                  className="w-32"
                />
                <Input
                  placeholder="Label (e.g., VS Code)"
                  value={port.label}
                  onChange={(e) =>
                    handlePortChange(index, 'label', e.target.value)
                  }
                  className="flex-1"
                />
                <Button
                  type="button"
                  size="sm"
                  variant="destructive"
                  onClick={() => handleRemovePort(index)}
                >
                  Remove
                </Button>
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddPort}
            >
              + Add Port
            </Button>
          </div>

          {/* Error Message */}
          {error && (
            <p className="text-sm text-destructive" role="alert">
              {error}
            </p>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={loading}>
            {loading ? 'Creating...' : 'Create Workspace'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
