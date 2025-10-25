import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Key } from 'lucide-react'
import { authApi } from '@/lib/api'

export function LoginPage() {
  const navigate = useNavigate()
  const [apiKey, setApiKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const { token, user } = await authApi.login(apiKey)
      // Store both the API key (as auth_token) and user info
      localStorage.setItem('auth_token', token)
      localStorage.setItem('user', JSON.stringify(user))
      navigate('/admin/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h1 className="text-foreground text-4xl font-bold tracking-tight">Claude Proxy</h1>
          <p className="text-muted-foreground mt-2 text-sm">Admin Panel</p>
        </div>

        <div className="border-border bg-card rounded-lg border p-8 shadow-lg">
          <h2 className="text-card-foreground mb-6 text-2xl font-semibold">Admin Sign In</h2>

          {error && (
            <div className="bg-destructive/10 text-destructive mb-4 rounded-md p-3 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="apiKey" className="text-foreground mb-2 block text-sm font-medium">
                API Key
              </label>
              <div className="relative">
                <Key className="text-muted-foreground absolute top-1/2 left-3 h-5 w-5 -translate-y-1/2" />
                <input
                  id="apiKey"
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  required
                  className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border py-2 pr-3 pl-10 focus:ring-2 focus:outline-none"
                  placeholder="Enter your API key"
                />
              </div>
              <p className="text-muted-foreground mt-1 text-xs">
                Use the API key from your config.yaml
              </p>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="bg-primary text-primary-foreground hover:bg-primary/90 w-full rounded-md px-4 py-2 font-medium transition-colors disabled:opacity-50"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </form>

          <div className="text-muted-foreground mt-4 text-center text-xs">
            <p>Configure your API key in <code className="bg-muted px-1 rounded">config.yaml</code></p>
            <p className="mt-1">under <code className="bg-muted px-1 rounded">auth.api_key</code></p>
          </div>
        </div>
      </div>
    </div>
  )
}
