import { useState, useEffect } from 'react'
import { Plus, Trash2, ExternalLink, Loader2, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:5201'

interface AppAccount {
  id: string
  name: string
  organization_uuid: string
  status: string
  created_at: number
  updated_at: number
  expires_at: number
}

export function AppTokensPage() {
  const [accounts, setAccounts] = useState<AppAccount[]>([])
  const [loading, setLoading] = useState(false)
  const [showModal, setShowModal] = useState(false)

  // Modal state
  const [step, setStep] = useState(1)
  const [appName, setAppName] = useState('')
  const [orgId, setOrgId] = useState('')
  const [authUrl, setAuthUrl] = useState('')
  const [state, setState] = useState('')
  const [codeVerifier, setCodeVerifier] = useState('')
  const [authCode, setAuthCode] = useState('')
  const [error, setError] = useState('')

  // Load accounts
  const loadAccounts = async () => {
    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch(`${API_BASE_URL}/api/app-accounts`, {
        headers: { 'X-API-Key': token || '' },
      })
      if (!response.ok) throw new Error('Failed to load accounts')
      const data = await response.json()
      setAccounts(data.accounts || [])
    } catch (err) {
      console.error('Failed to load accounts:', err)
    }
  }

  // Step 1: Start OAuth flow
  const handleStartOAuth = async () => {
    if (!appName.trim()) {
      setError('Please enter an app name')
      return
    }

    setLoading(true)
    setError('')

    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch(`${API_BASE_URL}/api/app-accounts`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': token || '',
        },
        body: JSON.stringify({ name: appName, org_id: orgId || undefined }),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error?.message || 'Failed to start OAuth')
      }

      const data = await response.json()
      setAuthUrl(data.authorization_url)
      setState(data.state)
      setCodeVerifier(data.code_verifier)
      setStep(2)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start OAuth')
    } finally {
      setLoading(false)
    }
  }

  // Step 2: Complete OAuth flow
  const handleCompleteOAuth = async () => {
    if (!authCode.trim()) {
      setError('Please enter the authorization code')
      return
    }

    setLoading(true)
    setError('')

    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch(`${API_BASE_URL}/api/app-accounts/complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': token || '',
        },
        body: JSON.stringify({
          name: appName,
          code: authCode,
          state,
          code_verifier: codeVerifier,
          org_id: orgId || undefined,
        }),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error?.message || 'Failed to complete OAuth')
      }

      // Success! Reload accounts and close modal
      await loadAccounts()
      resetModal()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to complete OAuth')
    } finally {
      setLoading(false)
    }
  }

  // Delete account
  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this app account?')) return

    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch(`${API_BASE_URL}/api/app-accounts/${id}`, {
        method: 'DELETE',
        headers: { 'X-API-Key': token || '' },
      })

      if (!response.ok) throw new Error('Failed to delete account')
      await loadAccounts()
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

  // Load accounts on mount
  useEffect(() => {
    loadAccounts()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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

      {/* Accounts list */}
      <Card>
        <CardContent className="p-6">
          {accounts.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground">No app accounts yet. Create one to get started!</p>
            </div>
          ) : (
            <div className="space-y-4">
              {accounts.map((account) => (
                <div
                  key={account.id}
                  className="flex items-center justify-between p-4 border rounded-lg"
                >
                  <div>
                    <h3 className="font-medium">{account.name}</h3>
                    <p className="text-sm text-muted-foreground">{account.organization_uuid}</p>
                    <span
                      className={`inline-flex mt-2 rounded-full px-2 py-1 text-xs font-medium ${
                        account.status === 'active'
                          ? 'bg-green-500/10 text-green-500'
                          : 'bg-gray-500/10 text-gray-500'
                      }`}
                    >
                      {account.status}
                    </span>
                  </div>
                  <Button variant="destructive" size="sm" onClick={() => handleDelete(account.id)}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-lg mx-4">
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
                    />
                  </div>
                  <div>
                    <Label htmlFor="orgId">Organization ID (Optional)</Label>
                    <Input
                      id="orgId"
                      value={orgId}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => setOrgId(e.target.value)}
                      placeholder="org_..."
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
                    <Button variant="outline" onClick={resetModal} className="flex-1">
                      Cancel
                    </Button>
                    <Button onClick={handleStartOAuth} disabled={loading} className="flex-1">
                      {loading ? (
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
                    <Button variant="outline" onClick={resetModal} className="flex-1">
                      Cancel
                    </Button>
                    <Button onClick={handleCompleteOAuth} disabled={loading} className="flex-1">
                      {loading ? (
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
