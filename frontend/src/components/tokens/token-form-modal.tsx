import { useState, useEffect } from 'react'
import { Dialog } from '@/components/ui/dialog'
import { useCreateToken, useUpdateToken } from '@/hooks/use-tokens'
import type { Token } from '@/types/token'

interface TokenFormModalProps {
  open: boolean
  onClose: () => void
  token?: Token
}

export function TokenFormModal({ open, onClose, token }: TokenFormModalProps) {
  const [name, setName] = useState('')
  const [key, setKey] = useState('')
  const [status, setStatus] = useState<'active' | 'inactive'>('active')

  const createMutation = useCreateToken()
  const updateMutation = useUpdateToken()

  useEffect(() => {
    if (token) {
      setName(token.name)
      setKey(token.key)
      setStatus(token.status)
    } else {
      setName('')
      setKey('')
      setStatus('active')
    }
  }, [token, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      if (token) {
        await updateMutation.mutateAsync({ id: token.id, name, key, status })
      } else {
        await createMutation.mutateAsync({ name, key, status })
      }
      onClose()
    } catch (error) {
      console.error('Failed to save token:', error)
    }
  }

  const isLoading = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={open} onClose={onClose} title={token ? 'Edit Token' : 'Create Token'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="name" className="text-foreground mb-2 block text-sm font-medium">
            Name
          </label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
            placeholder="Production API Key"
          />
        </div>

        <div>
          <label htmlFor="key" className="text-foreground mb-2 block text-sm font-medium">
            Token Value
          </label>
          <input
            id="key"
            type="text"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            required
            className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 font-mono text-sm focus:ring-2 focus:outline-none"
            placeholder="sk_prod_1a2b3c4d5e6f7g8h9i0j"
          />
        </div>

        <div>
          <label htmlFor="status" className="text-foreground mb-2 block text-sm font-medium">
            Status
          </label>
          <select
            id="status"
            value={status}
            onChange={(e) => setStatus(e.target.value as 'active' | 'inactive')}
            className="border-input bg-background text-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
          >
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
        </div>

        <div className="flex gap-3">
          <button
            type="button"
            onClick={onClose}
            className="border-input bg-background text-foreground hover:bg-muted flex-1 rounded-md border px-4 py-2 font-medium transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isLoading}
            className="bg-primary text-primary-foreground hover:bg-primary/90 flex-1 rounded-md px-4 py-2 font-medium transition-colors disabled:opacity-50"
          >
            {isLoading ? 'Saving...' : token ? 'Update' : 'Create'}
          </button>
        </div>
      </form>
    </Dialog>
  )
}
