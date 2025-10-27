import { useState } from 'react'
import { Edit2, Plus, Trash2, Copy, Check, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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
        <Button onClick={handleCreate}>
          <Plus className="mr-2 h-4 w-4" />
          Create Token
        </Button>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="py-12 text-center">
              <Loader2 className="text-primary mx-auto h-8 w-8 animate-spin" />
              <p className="text-muted-foreground mt-2">Loading tokens...</p>
            </div>
          ) : tokens && tokens.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>API Key</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Usage</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tokens.map((token) => (
                  <TableRow key={token.id}>
                    <TableCell className="font-medium">{token.name}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <code className="bg-muted text-foreground rounded px-2 py-1 text-xs">
                          {token.key.substring(0, 20)}...
                        </code>
                        <button
                          onClick={() => handleCopyKey(token.key, token.id)}
                          className="text-foreground/60 hover:bg-muted hover:text-foreground rounded p-1 transition-colors"
                        >
                          {copiedId === token.id ? (
                            <Check className="h-4 w-4 text-green-500" />
                          ) : (
                            <Copy className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                          token.status === 'active'
                            ? 'bg-green-500/10 text-green-500'
                            : 'bg-gray-500/10 text-gray-500'
                        }`}
                      >
                        {token.status}
                      </span>
                    </TableCell>
                    <TableCell>{token.usageCount.toLocaleString()}</TableCell>
                    <TableCell className="text-foreground/70 text-sm">
                      {token.lastUsedAt
                        ? new Date(token.lastUsedAt * 1000).toLocaleString()
                        : 'Never'}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => handleEdit(token)}
                          className="text-foreground/60 hover:bg-muted hover:text-foreground rounded p-1 transition-colors"
                        >
                          <Edit2 className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(token.id)}
                          disabled={deleteMutation.isPending}
                          className="text-foreground/60 hover:bg-destructive/10 hover:text-destructive rounded p-1 transition-colors"
                        >
                          {deleteMutation.isPending ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="py-12 text-center">
              <p className="text-muted-foreground">
                No tokens found. Create your first token to get started.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <TokenFormModal
        open={isModalOpen}
        onClose={handleCloseModal}
        token={editingToken}
        existingTokens={tokens || []}
      />
    </div>
  )
}
