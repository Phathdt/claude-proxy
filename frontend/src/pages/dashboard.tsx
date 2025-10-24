import { Activity, Key, TrendingUp, Users } from 'lucide-react'
import { useTokens } from '@/hooks/use-tokens'

export function DashboardPage() {
  const { data: tokens, isLoading } = useTokens()

  const stats = {
    totalTokens: tokens?.length || 0,
    activeTokens: tokens?.filter((t) => t.status === 'active').length || 0,
    totalUsage: tokens?.reduce((acc, t) => acc + t.usageCount, 0) || 0,
    avgUsage: tokens?.length
      ? Math.round(tokens.reduce((acc, t) => acc + t.usageCount, 0) / tokens.length)
      : 0,
  }

  const statCards = [
    {
      title: 'Total Tokens',
      value: stats.totalTokens,
      icon: Key,
      color: 'text-primary',
      bgColor: 'bg-primary/10',
    },
    {
      title: 'Active Tokens',
      value: stats.activeTokens,
      icon: Activity,
      color: 'text-green-500',
      bgColor: 'bg-green-500/10',
    },
    {
      title: 'Total Usage',
      value: stats.totalUsage.toLocaleString(),
      icon: TrendingUp,
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
    },
    {
      title: 'Avg. Usage/Token',
      value: stats.avgUsage.toLocaleString(),
      icon: Users,
      color: 'text-orange-500',
      bgColor: 'bg-orange-500/10',
    },
  ]

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-foreground text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">Overview of your API tokens and usage</p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => (
          <div key={stat.title} className="border-border bg-card rounded-lg border p-6 shadow-sm">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-muted-foreground text-sm font-medium">{stat.title}</p>
                <p className="text-card-foreground mt-2 text-3xl font-bold">
                  {isLoading ? '-' : stat.value}
                </p>
              </div>
              <div className={`rounded-lg ${stat.bgColor} p-3`}>
                <stat.icon className={`h-6 w-6 ${stat.color}`} />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="border-border bg-card rounded-lg border p-6 shadow-sm">
        <h2 className="text-card-foreground mb-4 text-xl font-semibold">Recent Tokens</h2>
        <div className="space-y-4">
          {isLoading ? (
            <p className="text-muted-foreground text-sm">Loading...</p>
          ) : tokens && tokens.length > 0 ? (
            tokens.slice(0, 5).map((token) => (
              <div
                key={token.id}
                className="border-border flex items-center justify-between rounded-md border p-4"
              >
                <div className="flex items-center gap-4">
                  <div
                    className={`h-2 w-2 rounded-full ${
                      token.status === 'active' ? 'bg-green-500' : 'bg-gray-400'
                    }`}
                  />
                  <div>
                    <p className="text-foreground font-medium">{token.name}</p>
                    <p className="text-muted-foreground text-sm">{token.usageCount} requests</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-foreground text-sm font-medium">
                    {token.status === 'active' ? 'Active' : 'Inactive'}
                  </p>
                  <p className="text-muted-foreground text-xs">
                    {token.lastUsedAt
                      ? new Date(token.lastUsedAt).toLocaleDateString()
                      : 'Never used'}
                  </p>
                </div>
              </div>
            ))
          ) : (
            <p className="text-muted-foreground text-sm">No tokens found</p>
          )}
        </div>
      </div>
    </div>
  )
}
