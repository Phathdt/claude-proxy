import { useState } from 'react'
import { Edit2, Plus, Trash2, Shield } from 'lucide-react'
import { useAppTokens, useDeleteAppToken } from '@/hooks/use-app-tokens'
import { AppTokenFormModal } from '@/components/app-tokens/app-token-form-modal'
import type { AppToken } from '@/types/app-token'

export function AppTokensPage() {
  const { data: appTokens, isLoading } = useAppTokens()
  const deleteMutation = useDeleteAppToken()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingToken, setEditingToken] = useState<AppToken | undefined>()

  const handleCreate = () => {
    setEditingToken(undefined)
    setIsModalOpen(true)
  }

  const handleEdit = (token: AppToken) => {
    setEditingToken(token)
    setIsModalOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (confirm('Are you sure you want to delete this app token?')) {
      await deleteMutation.mutateAsync(id)
    }
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingToken(undefined)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold tracking-tight">App Tokens (OAuth)</h1>
          <p className="text-muted-foreground">
            Manage OAuth 2.0 applications and client credentials
          </p>
        </div>
        <button
          onClick={handleCreate}
          className="bg-primary text-primary-foreground hover:bg-primary/90 flex items-center gap-2 rounded-md px-4 py-2 font-medium transition-colors"
        >
          <Plus className="h-5 w-5" />
          Create App Token
        </button>
      </div>

      {/* Table */}
      <div className="border-border bg-card rounded-lg border shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-border bg-muted/50 border-b">
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Application
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Email / Org
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Type / Account
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Status
                </th>
                <th className="text-muted-foreground px-6 py-3 text-right text-sm font-medium">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-border divide-y">
              {isLoading ? (
                <tr>
                  <td colSpan={5} className="text-muted-foreground px-6 py-8 text-center">
                    Loading...
                  </td>
                </tr>
              ) : appTokens && appTokens.length > 0 ? (
                appTokens.map((token) => (
                  <tr key={token.id} className="hover:bg-muted/50 transition-colors">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <Shield className="text-primary h-4 w-4" />
                        <p className="text-foreground font-medium">{token.name}</p>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-foreground text-sm">{token.email}</p>
                      <code className="text-muted-foreground mt-1 text-xs">{token.orgId}</code>
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex rounded-full bg-purple-500/10 px-2 py-0.5 text-xs font-medium text-purple-500">
                        {token.type}
                      </span>
                      <p className="text-muted-foreground mt-1 text-xs capitalize">
                        {token.accountType}
                      </p>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${
                          token.status === 'active'
                            ? 'bg-green-500/10 text-green-500'
                            : 'bg-gray-500/10 text-gray-500'
                        }`}
                      >
                        {token.status}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => handleEdit(token)}
                          className="text-muted-foreground hover:bg-muted hover:text-foreground rounded p-1 transition-colors"
                        >
                          <Edit2 className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(token.id)}
                          disabled={deleteMutation.isPending}
                          className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive rounded p-1 transition-colors"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={5} className="text-muted-foreground px-6 py-8 text-center">
                    No app tokens found. Create your first OAuth application to get started.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      <AppTokenFormModal open={isModalOpen} onClose={handleCloseModal} token={editingToken} />
    </div>
  )
}
