import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { oauthApi, type OAuthAuthorizeResponse } from '@/lib/api'
import { CheckCircle2, XCircle, ExternalLink, Loader2 } from 'lucide-react'

export default function OAuthSetup() {
  const [orgId, setOrgId] = useState('')
  const [oauthData, setOauthData] = useState<OAuthAuthorizeResponse | null>(null)
  const [authCode, setAuthCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')
  const [accountInfo, setAccountInfo] = useState<{
    organization_uuid: string
    expires_at: number
  } | null>(null)

  const handleGenerateUrl = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await oauthApi.getAuthorizeUrl()
      setOauthData(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate OAuth URL')
    } finally {
      setLoading(false)
    }
  }

  const handleExchangeCode = async () => {
    if (!oauthData || !authCode) {
      setError('Please enter the authorization code')
      return
    }

    setLoading(true)
    setError('')
    try {
      const result = await oauthApi.exchangeCode({
        code: authCode,
        state: oauthData.state,
        code_verifier: oauthData.code_verifier,
        org_id: orgId || undefined,
      })
      
      setSuccess(true)
      setAccountInfo({
        organization_uuid: result.organization_uuid,
        expires_at: result.expires_at,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to exchange code')
    } finally {
      setLoading(false)
    }
  }

  const formatExpiresAt = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString()
  }

  if (success && accountInfo) {
    return (
      <div className="p-8">
        <Card className="max-w-2xl mx-auto">
          <CardHeader>
            <div className="flex items-center gap-2">
              <CheckCircle2 className="h-6 w-6 text-green-500" />
              <CardTitle>OAuth Setup Complete!</CardTitle>
            </div>
            <CardDescription>
              Your Claude account has been successfully configured
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label className="text-muted-foreground">Organization UUID</Label>
              <p className="font-mono text-sm mt-1">{accountInfo.organization_uuid}</p>
            </div>
            <div>
              <Label className="text-muted-foreground">Token Expires At</Label>
              <p className="text-sm mt-1">{formatExpiresAt(accountInfo.expires_at)}</p>
            </div>
            <Alert>
              <AlertDescription>
                Tokens will be automatically refreshed before expiry. You can now use the Claude API!
              </AlertDescription>
            </Alert>
            <Button onClick={() => (window.location.href = '/admin/dashboard')} className="w-full">
              Go to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="p-8">
      <Card className="max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle>OAuth Setup</CardTitle>
          <CardDescription>
            Connect your Claude account using OAuth 2.0 authentication
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Step 1: Optional Organization ID */}
          <div className="space-y-2">
            <Label htmlFor="orgId">
              Organization ID <span className="text-muted-foreground">(Optional)</span>
            </Label>
            <Input
              id="orgId"
              placeholder="org_..."
              value={orgId}
              onChange={(e) => setOrgId(e.target.value)}
              disabled={!!oauthData}
            />
            <p className="text-sm text-muted-foreground">
              If not provided, it will be automatically fetched from your account
            </p>
          </div>

          {/* Step 2: Generate OAuth URL */}
          {!oauthData && (
            <Button onClick={handleGenerateUrl} disabled={loading} className="w-full">
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Generating...
                </>
              ) : (
                'Generate OAuth URL'
              )}
            </Button>
          )}

          {/* Step 3: Show OAuth URL */}
          {oauthData && !success && (
            <div className="space-y-4">
              <Alert>
                <AlertDescription>
                  <strong className="block mb-2">Step 1: Authorize with Claude</strong>
                  Click the button below to open Claude's authorization page. After authorizing, you'll be redirected back with a code in the URL.
                </AlertDescription>
              </Alert>

              <Button asChild className="w-full">
                <a
                  href={oauthData.authorization_url}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Connect to Claude
                  <ExternalLink className="ml-2 h-4 w-4" />
                </a>
              </Button>

              {/* Debug info */}
              <details className="text-sm">
                <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
                  Show technical details
                </summary>
                <div className="mt-2 space-y-2 font-mono text-xs bg-muted p-3 rounded">
                  <div>
                    <strong>State:</strong> {oauthData.state}
                  </div>
                  <div>
                    <strong>Code Verifier:</strong> {oauthData.code_verifier}
                  </div>
                </div>
              </details>

              {/* Step 4: Enter authorization code */}
              <div className="space-y-2">
                <Label htmlFor="authCode">
                  <strong>Step 2:</strong> Paste Authorization Code
                </Label>
                <Input
                  id="authCode"
                  placeholder="Enter the code from the callback URL"
                  value={authCode}
                  onChange={(e) => setAuthCode(e.target.value)}
                />
                <p className="text-sm text-muted-foreground">
                  After authorizing, copy the <code className="px-1 bg-muted">code</code> parameter from the redirect URL
                </p>
              </div>

              <Button 
                onClick={handleExchangeCode} 
                disabled={loading || !authCode} 
                className="w-full"
              >
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Exchanging Code...
                  </>
                ) : (
                  'Complete Setup'
                )}
              </Button>
            </div>
          )}

          {/* Error display */}
          {error && (
            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
