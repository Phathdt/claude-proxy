import { useState, useEffect } from 'react'
import { Dialog } from '@/components/ui/dialog'
import { useCreateAppToken, useUpdateAppToken } from '@/hooks/use-app-tokens'
import type { AppToken } from '@/types/app-token'

interface AppTokenFormModalProps {
  open: boolean
  onClose: () => void
  token?: AppToken
}

export function AppTokenFormModal({ open, onClose, token }: AppTokenFormModalProps) {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [orgId, setOrgId] = useState('')
  const [type, setType] = useState<'oauth' | 'cookies'>('oauth')
  const [accountType, setAccountType] = useState<'pro' | 'max'>('pro')
  const [status, setStatus] = useState<'active' | 'inactive'>('active')

  const createMutation = useCreateAppToken()
  const updateMutation = useUpdateAppToken()

  useEffect(() => {
    if (token) {
      setName(token.name)
      setEmail(token.email)
      setOrgId(token.orgId)
      setType(token.type)
      setAccountType(token.accountType)
      setStatus(token.status)
    } else {
      setName('')
      setEmail('')
      setOrgId('')
      setType('oauth')
      setAccountType('pro')
      setStatus('active')
    }
  }, [token, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      if (token) {
        await updateMutation.mutateAsync({
          id: token.id,
          name,
          email,
          orgId,
          type,
          accountType,
          status,
        })
      } else {
        await createMutation.mutateAsync({
          name,
          email,
          orgId,
          type,
          accountType,
          status,
        })
      }
      onClose()
    } catch (error) {
      console.error('Failed to save app token:', error)
    }
  }

  const isLoading = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={open} onClose={onClose} title={token ? 'Edit App Token' : 'Create App Token'}>
      <form onSubmit={handleSubmit} className="max-h-[70vh] space-y-4 overflow-y-auto">
        {/* Name */}
        <div>
          <label htmlFor="name" className="text-foreground mb-2 block text-sm font-medium">
            Application Name
          </label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
            placeholder="My Web Application"
          />
        </div>

        {/* Email */}
        <div>
          <label htmlFor="email" className="text-foreground mb-2 block text-sm font-medium">
            Email
          </label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
            placeholder="developer@example.com"
          />
        </div>

        {/* Organization ID */}
        <div>
          <label htmlFor="orgId" className="text-foreground mb-2 block text-sm font-medium">
            Organization ID
          </label>
          <input
            id="orgId"
            type="text"
            value={orgId}
            onChange={(e) => setOrgId(e.target.value)}
            required
            className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 font-mono text-sm focus:ring-2 focus:outline-none"
            placeholder="org_abc123xyz"
          />
        </div>

        {/* Type */}
        <div>
          <label htmlFor="type" className="text-foreground mb-2 block text-sm font-medium">
            Type
          </label>
          <select
            id="type"
            value={type}
            onChange={(e) => setType(e.target.value as 'oauth' | 'cookies')}
            className="border-input bg-background text-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
          >
            <option value="oauth">OAuth</option>
            <option value="cookies">Cookies</option>
          </select>
        </div>

        {/* Account Type */}
        <div>
          <label htmlFor="accountType" className="text-foreground mb-2 block text-sm font-medium">
            Account Type
          </label>
          <select
            id="accountType"
            value={accountType}
            onChange={(e) => setAccountType(e.target.value as 'pro' | 'max')}
            className="border-input bg-background text-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
          >
            <option value="pro">Pro</option>
            <option value="max">Max</option>
          </select>
        </div>

        {/* Status */}
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

        {/* Actions */}
        <div className="bg-card sticky bottom-0 flex gap-3 pt-4">
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
