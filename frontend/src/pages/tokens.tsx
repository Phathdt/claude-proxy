import { useState } from 'react'
import { Edit2, Plus, Trash2, Copy, Check } from 'lucide-react'
import { useTokens, useDeleteToken } from '@/hooks/use-tokens'
import { TokenFormModal } from '@/components/tokens/token-form-modal'
import type { Token } from '@/types/token'

export function TokensPage() {
  const { data: tokens, isLoading } = useTokens()
  const deleteMutation = useDeleteToken()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingToken, setEditingToken] = useState<Token | undefined>()
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const handleCreate = () => {
    setEditingToken(undefined)
    setIsModalOpen(true)
  }

  const handleEdit = (token: Token) => {
    setEditingToken(token)
    setIsModalOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (confirm('Are you sure you want to delete this token?')) {
      await deleteMutation.mutateAsync(id)
    }
  }

  const handleCopyKey = async (key: string, id: string) => {
    await navigator.clipboard.writeText(key)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingToken(undefined)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold tracking-tight">Tokens</h1>
          <p className="text-muted-foreground">Manage your API tokens</p>
        </div>
        <button
          onClick={handleCreate}
          className="bg-primary text-primary-foreground hover:bg-primary/90 flex items-center gap-2 rounded-md px-4 py-2 font-medium transition-colors"
        >
          <Plus className="h-5 w-5" />
          Create Token
        </button>
      </div>

      {/* Table */}
      <div className="border-border bg-card rounded-lg border shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-border bg-muted/50 border-b">
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Name
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  API Key
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Status
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Usage
                </th>
                <th className="text-muted-foreground px-6 py-3 text-left text-sm font-medium">
                  Last Used
                </th>
                <th className="text-muted-foreground px-6 py-3 text-right text-sm font-medium">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-border divide-y">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="text-muted-foreground px-6 py-8 text-center">
                    Loading...
                  </td>
                </tr>
              ) : tokens && tokens.length > 0 ? (
                tokens.map((token) => (
                  <tr key={token.id} className="hover:bg-muted/50 transition-colors">
                    <td className="px-6 py-4">
                      <p className="text-foreground font-medium">{token.name}</p>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <code className="bg-muted text-foreground rounded px-2 py-1 text-sm">
                          {token.key.substring(0, 20)}...
                        </code>
                        <button
                          onClick={() => handleCopyKey(token.key, token.id)}
                          className="text-muted-foreground hover:bg-muted hover:text-foreground rounded p-1 transition-colors"
                        >
                          {copiedId === token.id ? (
                            <Check className="h-4 w-4 text-green-500" />
                          ) : (
                            <Copy className="h-4 w-4" />
                          )}
                        </button>
                      </div>
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
                    <td className="text-foreground px-6 py-4">
                      {token.usageCount.toLocaleString()}
                    </td>
                    <td className="text-muted-foreground px-6 py-4 text-sm">
                      {token.lastUsedAt ? new Date(token.lastUsedAt * 1000).toLocaleString() : 'Never'}
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
                  <td colSpan={6} className="text-muted-foreground px-6 py-8 text-center">
                    No tokens found. Create your first token to get started.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      <TokenFormModal open={isModalOpen} onClose={handleCloseModal} token={editingToken} />
    </div>
  )
}
