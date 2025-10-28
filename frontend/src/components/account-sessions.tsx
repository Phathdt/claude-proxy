import { useState } from 'react'
import { useAccountSessions, useRevokeSession, useRevokeAccountSessions } from '@/hooks/use-sessions'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import type { Session } from '@/types/session'

interface AccountSessionsProps {
  accountId: string
}

export function AccountSessions({ accountId }: AccountSessionsProps) {
  const { data, isLoading, error, refetch } = useAccountSessions(accountId)
  const revokeSessionMutation = useRevokeSession()
  const revokeAllMutation = useRevokeAccountSessions()
  const [sessionToRevoke, setSessionToRevoke] = useState<string | null>(null)
  const [showRevokeAllDialog, setShowRevokeAllDialog] = useState(false)

  const handleRevokeSession = async () => {
    if (!sessionToRevoke) return

    try {
      await revokeSessionMutation.mutateAsync(sessionToRevoke)
      setSessionToRevoke(null)
    } catch (error) {
      console.error('Failed to revoke session:', error)
    }
  }

  const handleRevokeAll = async () => {
    try {
      await revokeAllMutation.mutateAsync(accountId)
      setShowRevokeAllDialog(false)
    } catch (error) {
      console.error('Failed to revoke all sessions:', error)
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
      <div className="flex h-32 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load sessions: {error instanceof Error ? error.message : 'Unknown error'}
        </AlertDescription>
      </Alert>
    )
  }

  const sessions = data?.sessions || []
  const totalSessions = data?.total || 0

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Active Sessions ({totalSessions})</CardTitle>
              <CardDescription>
                Sessions currently active for this account
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <Button onClick={() => refetch()} size="sm" disabled={isLoading}>
                <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                Refresh
              </Button>
              {totalSessions > 0 && (
                <Button
                  onClick={() => setShowRevokeAllDialog(true)}
                  variant="destructive"
                  size="sm"
                  disabled={revokeAllMutation.isPending}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Revoke All
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {sessions.length === 0 ? (
            <div className="flex h-32 items-center justify-center text-muted-foreground">
              No active sessions for this account
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Session ID</TableHead>
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
                        {session.id.substring(0, 12)}...
                      </code>
                    </TableCell>
                    <TableCell className="text-foreground text-sm">{session.ipAddress}</TableCell>
                    <TableCell className="text-foreground/70 max-w-[200px] truncate text-xs" title={session.userAgent}>
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

      {/* Revoke single session dialog */}
      <AlertDialog open={!!sessionToRevoke} onOpenChange={() => setSessionToRevoke(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke Session?</AlertDialogTitle>
            <AlertDialogDescription>
              This will immediately terminate the session. The user will need to create a new session
              to continue using the API.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleRevokeSession}>
              {revokeSessionMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Revoking...
                </>
              ) : (
                'Revoke Session'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Revoke all sessions dialog */}
      <AlertDialog open={showRevokeAllDialog} onOpenChange={setShowRevokeAllDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke All Sessions?</AlertDialogTitle>
            <AlertDialogDescription>
              This will immediately terminate all {totalSessions} active session
              {totalSessions !== 1 ? 's' : ''} for this account. Users will need to create new
              sessions to continue using the API.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleRevokeAll} className="bg-destructive text-destructive-foreground">
              {revokeAllMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Revoking...
                </>
              ) : (
                `Revoke ${totalSessions} Session${totalSessions !== 1 ? 's' : ''}`
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
