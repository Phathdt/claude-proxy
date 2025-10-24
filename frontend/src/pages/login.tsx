import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Lock, Mail } from 'lucide-react'
import { authApi } from '@/lib/api'

export function LoginPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const { token } = await authApi.login(email, password)
      localStorage.setItem('auth_token', token)
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
          <h2 className="text-card-foreground mb-6 text-2xl font-semibold">Sign in</h2>

          {error && (
            <div className="bg-destructive/10 text-destructive mb-4 rounded-md p-3 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="email" className="text-foreground mb-2 block text-sm font-medium">
                Email
              </label>
              <div className="relative">
                <Mail className="text-muted-foreground absolute top-1/2 left-3 h-5 w-5 -translate-y-1/2" />
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border py-2 pr-3 pl-10 focus:ring-2 focus:outline-none"
                  placeholder="admin@example.com"
                />
              </div>
            </div>

            <div>
              <label htmlFor="password" className="text-foreground mb-2 block text-sm font-medium">
                Password
              </label>
              <div className="relative">
                <Lock className="text-muted-foreground absolute top-1/2 left-3 h-5 w-5 -translate-y-1/2" />
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border py-2 pr-3 pl-10 focus:ring-2 focus:outline-none"
                  placeholder="••••••••"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="bg-primary text-primary-foreground hover:bg-primary/90 w-full rounded-md px-4 py-2 font-medium transition-colors disabled:opacity-50"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </form>

          <p className="text-muted-foreground mt-4 text-center text-xs">
            Mock login - use any email/password
          </p>
        </div>
      </div>
    </div>
  )
}
