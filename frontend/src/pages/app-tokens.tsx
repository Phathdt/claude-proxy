import { useState } from 'react'
import { Plus, Trash2, ExternalLink, Loader2, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  useAppAccounts,
  useCreateAppAccount,
  useCompleteAppAccount,
  useDeleteAppAccount,
} from '@/hooks/use-app-accounts'

export function AppTokensPage() {
  // React Query hooks
  const { data: accounts = [], isLoading } = useAppAccounts()
  const createMutation = useCreateAppAccount()
  const completeMutation = useCompleteAppAccount()
  const deleteMutation = useDeleteAppAccount()

  // Modal state
  const [showModal, setShowModal] = useState(false)
  const [step, setStep] = useState(1)
  const [appName, setAppName] = useState('')
  const [orgId, setOrgId] = useState('')
  const [authUrl, setAuthUrl] = useState('')
  const [state, setState] = useState('')
  const [codeVerifier, setCodeVerifier] = useState('')
  const [authCode, setAuthCode] = useState('')
  const [error, setError] = useState('')

  // Step 1: Start OAuth flow
  const handleStartOAuth = async () => {
    if (!appName.trim()) {
      setError('Please enter an app name')
      return
    }

    setError('')
    try {
      const data = await createMutation.mutateAsync({
        name: appName,
        org_id: orgId || undefined,
      })
      setAuthUrl(data.authorization_url)
      setState(data.state)
      setCodeVerifier(data.code_verifier)
      setStep(2)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start OAuth')
    }
  }

  // Step 2: Complete OAuth flow
  const handleCompleteOAuth = async () => {
    if (!authCode.trim()) {
      setError('Please enter the authorization code')
      return
    }

    setError('')
    try {
      await completeMutation.mutateAsync({
        name: appName,
        code: authCode,
        state,
        code_verifier: codeVerifier,
        org_id: orgId || undefined,
      })
      // Success! Close modal
      resetModal()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete OAuth')
    }
  }

  // Delete account
  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this app account?')) return

    try {
      await deleteMutation.mutateAsync(id)
    } catch (err) {
      alert('Failed to delete account')
    }
  }

  const resetModal = () => {
    setShowModal(false)
    setStep(1)
    setAppName('')
    setOrgId('')
    setAuthUrl('')
    setState('')
    setCodeVerifier('')
    setAuthCode('')
    setError('')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">App Tokens (OAuth)</h1>
          <p className="text-muted-foreground">Manage Claude OAuth applications</p>
        </div>
        <Button onClick={() => setShowModal(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create App Token
        </Button>
      </div>

      {/* Accounts Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="text-center py-12">
              <Loader2 className="mx-auto h-8 w-8 animate-spin text-primary" />
              <p className="mt-2 text-muted-foreground">Loading accounts...</p>
            </div>
          ) : accounts.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground">No app accounts yet. Create one to get started!</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Organization UUID</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Expires At</TableHead>
                  <TableHead>Created At</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell className="font-medium">{account.name}</TableCell>
                    <TableCell>
                      <code className="text-xs bg-muted px-2 py-1 rounded">
                        {account.organization_uuid}
                      </code>
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                          account.status === 'active'
                            ? 'bg-green-500/10 text-green-500'
                            : 'bg-gray-500/10 text-gray-500'
                        }`}
                      >
                        {account.status}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm">
                      {new Date(account.expires_at * 1000).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(account.created_at * 1000).toLocaleString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(account.id)}
                        disabled={deleteMutation.isPending}
                      >
                        {deleteMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4 text-destructive" />
                        )}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <Card className="w-full max-w-lg">
            <CardHeader>
              <CardTitle>Create App Token</CardTitle>
              <CardDescription>Step {step} of 2</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {step === 1 && (
                <>
                  <div>
                    <Label htmlFor="appName">App Name *</Label>
                    <Input
                      id="appName"
                      value={appName}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAppName(e.target.value)}
                      placeholder="My Claude App"
                      disabled={createMutation.isPending}
                    />
                  </div>
                  <div>
                    <Label htmlFor="orgId">Organization ID (Optional)</Label>
                    <Input
                      id="orgId"
                      value={orgId}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setOrgId(e.target.value)}
                      placeholder="org_..."
                      disabled={createMutation.isPending}
                    />
                    <p className="text-xs text-muted-foreground mt-1">
                      If not provided, will be fetched from your account
                    </p>
                  </div>

                  {error && (
                    <Alert variant="destructive">
                      <XCircle className="h-4 w-4" />
                      <AlertDescription>{error}</AlertDescription>
                    </Alert>
                  )}

                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      onClick={resetModal}
                      disabled={createMutation.isPending}
                      className="flex-1"
                    >
                      Cancel
                    </Button>
                    <Button
                      onClick={handleStartOAuth}
                      disabled={createMutation.isPending}
                      className="flex-1"
                    >
                      {createMutation.isPending ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Generating...
                        </>
                      ) : (
                        'Authenticate with Claude'
                      )}
                    </Button>
                  </div>
                </>
              )}

              {step === 2 && (
                <>
                  <Alert>
                    <CheckCircle2 className="h-4 w-4" />
                    <AlertDescription>
                      <strong>Step 1:</strong> Click the button below to authorize with Claude
                    </AlertDescription>
                  </Alert>

                  <Button asChild className="w-full">
                    <a href={authUrl} target="_blank" rel="noopener noreferrer">
                      Open Claude Authorization
                      <ExternalLink className="ml-2 h-4 w-4" />
                    </a>
                  </Button>

                  <div className="border-t pt-4">
                    <Label htmlFor="authCode">
                      <strong>Step 2:</strong> Paste Authorization Code
                    </Label>
                    <Input
                      id="authCode"
                      value={authCode}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAuthCode(e.target.value)}
                      placeholder="Paste the code from the callback URL"
                      className="mt-2"
                      disabled={completeMutation.isPending}
                    />
                    <p className="text-xs text-muted-foreground mt-1">
                      After authorizing, copy the <code>code</code> parameter from the redirect URL
                    </p>
                  </div>

                  {error && (
                    <Alert variant="destructive">
                      <XCircle className="h-4 w-4" />
                      <AlertDescription>{error}</AlertDescription>
                    </Alert>
                  )}

                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      onClick={resetModal}
                      disabled={completeMutation.isPending}
                      className="flex-1"
                    >
                      Cancel
                    </Button>
                    <Button
                      onClick={handleCompleteOAuth}
                      disabled={completeMutation.isPending}
                      className="flex-1"
                    >
                      {completeMutation.isPending ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Completing...
                        </>
                      ) : (
                        'Complete Setup'
                      )}
                    </Button>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
