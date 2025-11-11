import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface ResetConfirmDialogProps {
  open: boolean
  workspaceName?: string
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
  loading?: boolean
}

export function ResetConfirmDialog({
  open,
  workspaceName,
  onOpenChange,
  onConfirm,
  loading = false,
}: ResetConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Reset Workspace</DialogTitle>
          <DialogDescription>
            Are you sure you want to reset <strong className="text-foreground">{workspaceName}</strong>? All data will be lost and the workspace will be recreated with a fresh container.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={loading}>
            {loading ? 'Resetting...' : 'Reset'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
