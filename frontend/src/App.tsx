import { useState, useEffect } from 'react'
import { Activity, CheckCircle2, XCircle } from 'lucide-react'

interface HealthResponse {
  status: string
  timestamp: number
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/health')
      .then((res) => res.json())
      .then((data) => {
        setHealth(data)
        setLoading(false)
      })
      .catch((err) => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  return (
    <div className="min-h-screen bg-background font-sans">
      <div className="container mx-auto px-4 py-16">
        <div className="mx-auto max-w-2xl space-y-8">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-5xl font-bold tracking-tight text-foreground">
              Claude Proxy
            </h1>
            <p className="mt-4 text-lg text-muted-foreground">
              API Gateway & Health Monitoring
            </p>
          </div>

          {/* Health Status Card */}
          <div className="rounded-lg border border-border bg-card p-6 shadow-lg">
            <div className="mb-4 flex items-center gap-3">
              <Activity className="h-6 w-6 text-primary" />
              <h2 className="text-2xl font-semibold text-card-foreground">
                Backend Health
              </h2>
            </div>

            {loading && (
              <div className="flex items-center gap-2 text-muted-foreground">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                <span>Checking backend status...</span>
              </div>
            )}

            {error && (
              <div className="flex items-start gap-3 rounded-md bg-destructive/10 p-4">
                <XCircle className="h-5 w-5 text-destructive" />
                <div>
                  <p className="font-medium text-destructive">Connection Error</p>
                  <p className="mt-1 text-sm text-destructive-foreground">{error}</p>
                </div>
              </div>
            )}

            {health && (
              <div className="space-y-4">
                <div className="flex items-center gap-3 rounded-md bg-primary/10 p-4">
                  <CheckCircle2 className="h-5 w-5 text-primary" />
                  <div>
                    <p className="font-medium text-foreground">Status: Online</p>
                    <p className="text-sm text-muted-foreground">
                      All systems operational
                    </p>
                  </div>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="rounded-md border border-border bg-muted/50 p-4">
                    <p className="text-sm font-medium text-muted-foreground">
                      Backend Status
                    </p>
                    <p className="mt-1 text-xl font-semibold text-foreground">
                      {health.status}
                    </p>
                  </div>
                  <div className="rounded-md border border-border bg-muted/50 p-4">
                    <p className="text-sm font-medium text-muted-foreground">
                      Last Check
                    </p>
                    <p className="mt-1 text-xl font-semibold text-foreground">
                      {new Date(health.timestamp * 1000).toLocaleTimeString()}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Info Card */}
          <div className="rounded-lg border border-border bg-card p-6 shadow-md">
            <h3 className="mb-3 text-lg font-semibold text-card-foreground">
              Development Info
            </h3>
            <div className="space-y-2 text-sm text-muted-foreground">
              <div className="flex justify-between">
                <span>Frontend:</span>
                <span className="font-mono text-foreground">
                  http://localhost:5173
                </span>
              </div>
              <div className="flex justify-between">
                <span>Backend API:</span>
                <span className="font-mono text-foreground">
                  http://localhost:4000/api
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
