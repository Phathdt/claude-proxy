import { useState } from 'react'
import { useAllSessions, useRevokeSession } from '@/hooks/use-sessions'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { AlertCircle, Loader2, Trash2, RefreshCw } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Dialog } from '@/components/ui/dialog'
import type { Session } from '@/types/session'

export default function SessionsPage() {
  const { data, isLoading, error, refetch } = useAllSessions()
  const revokeSessionMutation = useRevokeSession()
  const [sessionToRevoke, setSessionToRevoke] = useState<string | null>(null)

  const handleRevokeSession = async () => {
    if (!sessionToRevoke) return

    try {
      await revokeSessionMutation.mutateAsync(sessionToRevoke)
      setSessionToRevoke(null)
    } catch (error) {
      console.error('Failed to revoke session:', error)
    }
  }

  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins} min ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
  }

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>
          Failed to load sessions: {error instanceof Error ? error.message : 'Unknown error'}
        </AlertDescription>
      </Alert>
    )
  }

  const sessions = data?.sessions || []
  const totalSessions = data?.total || 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold tracking-tight">Active Sessions</h1>
          <p className="text-muted-foreground">
            Monitor and manage active client sessions (per IP + User-Agent)
          </p>
        </div>
        <Button onClick={() => refetch()} disabled={isLoading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Sessions ({totalSessions})</CardTitle>
          <CardDescription>All active sessions with concurrent usage tracking</CardDescription>
        </CardHeader>
        <CardContent>
          {sessions.length === 0 ? (
            <div className="text-muted-foreground flex h-32 items-center justify-center">
              No active sessions
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Session ID</TableHead>
                  <TableHead>Token ID</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead>User Agent</TableHead>
                  <TableHead>Last Seen</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {sessions.map((session: Session) => (
                  <TableRow key={session.id}>
                    <TableCell className="font-mono text-xs">
                      <code className="bg-muted text-foreground rounded px-2 py-1">
                        {session.id.substring(0, 8)}...
                      </code>
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      <code className="bg-muted text-foreground rounded px-2 py-1">
                        {session.tokenId.substring(0, 8)}...
                      </code>
                    </TableCell>
                    <TableCell className="text-foreground text-sm">{session.ipAddress}</TableCell>
                    <TableCell
                      className="text-foreground/70 max-w-[200px] truncate text-xs"
                      title={session.userAgent}
                    >
                      {session.userAgent}
                    </TableCell>
                    <TableCell className="text-foreground text-sm">
                      {formatTimeAgo(session.lastSeenAt)}
                    </TableCell>
                    <TableCell className="text-foreground text-sm">
                      {formatTimeAgo(session.expiresAt)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={session.isActive ? 'default' : 'secondary'}>
                        {session.isActive ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSessionToRevoke(session.id)}
                        disabled={revokeSessionMutation.isPending}
                      >
                        <Trash2 className="text-destructive h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={!!sessionToRevoke}
        onClose={() => setSessionToRevoke(null)}
        title="Revoke Session?"
      >
        <p className="text-muted-foreground mb-4 text-sm">
          This will immediately terminate the session. The user will need to create a new session to
          continue using the API.
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setSessionToRevoke(null)}>
            Cancel
          </Button>
          <Button onClick={handleRevokeSession} variant="destructive">
            {revokeSessionMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Revoking...
              </>
            ) : (
              'Revoke Session'
            )}
          </Button>
        </div>
      </Dialog>
    </div>
  )
}
