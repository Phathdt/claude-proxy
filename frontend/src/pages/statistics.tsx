import {
  Activity,
  AlertCircle,
  CheckCircle2,
  Clock,
  Loader2,
  RefreshCw,
  TrendingUp,
  Users,
  XCircle,
  PauseCircle,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useStatistics } from '@/hooks/use-statistics'
import type { SystemHealth } from '@/types/statistics'

export function StatisticsPage() {
  const { data: statistics, isLoading, refetch, dataUpdatedAt } = useStatistics()

  const handleRefresh = () => {
    refetch()
  }

  const getSystemHealthConfig = (
    health: SystemHealth
  ): {
    label: string
    color: string
    bgColor: string
    icon: typeof CheckCircle2
  } => {
    switch (health) {
      case 'healthy':
        return {
          label: 'Healthy',
          color: 'text-green-500',
          bgColor: 'bg-green-500/10',
          icon: CheckCircle2,
        }
      case 'degraded':
        return {
          label: 'Degraded',
          color: 'text-amber-500',
          bgColor: 'bg-amber-500/10',
          icon: AlertCircle,
        }
      case 'unhealthy':
        return {
          label: 'Unhealthy',
          color: 'text-red-500',
          bgColor: 'bg-red-500/10',
          icon: XCircle,
        }
    }
  }

  const healthConfig = statistics
    ? getSystemHealthConfig(statistics.systemHealth)
    : getSystemHealthConfig('healthy')

  const accountStatusCards = [
    {
      title: 'Total Accounts',
      value: statistics?.totalAccounts || 0,
      icon: Users,
      color: 'text-primary',
      bgColor: 'bg-primary/10',
    },
    {
      title: 'Active Accounts',
      value: statistics?.activeAccounts || 0,
      icon: CheckCircle2,
      color: 'text-green-500',
      bgColor: 'bg-green-500/10',
    },
    {
      title: 'Rate Limited',
      value: statistics?.rateLimitedAccounts || 0,
      icon: AlertCircle,
      color: 'text-amber-500',
      bgColor: 'bg-amber-500/10',
    },
    {
      title: 'Invalid Accounts',
      value: statistics?.invalidAccounts || 0,
      icon: XCircle,
      color: 'text-red-500',
      bgColor: 'bg-red-500/10',
    },
    {
      title: 'Inactive Accounts',
      value: statistics?.inactiveAccounts || 0,
      icon: PauseCircle,
      color: 'text-gray-500',
      bgColor: 'bg-gray-500/10',
    },
  ]

  const tokenHealthCards = [
    {
      title: 'Accounts Needing Refresh',
      value: statistics?.accountsNeedingRefresh || 0,
      icon: RefreshCw,
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
      suffix: '',
    },
    {
      title: 'Oldest Token Age',
      value: statistics?.oldestTokenAgeHours ? statistics.oldestTokenAgeHours.toFixed(1) : '0.0',
      icon: Clock,
      color: 'text-purple-500',
      bgColor: 'bg-purple-500/10',
      suffix: 'hours',
    },
  ]

  const lastUpdated = dataUpdatedAt ? new Date(dataUpdatedAt).toLocaleTimeString() : 'Never'

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-foreground text-3xl font-bold tracking-tight">Statistics</h1>
          <p className="text-muted-foreground">System health and account statistics</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="text-muted-foreground text-sm">
            Last updated: <span className="font-medium">{lastUpdated}</span>
          </div>
          <Button onClick={handleRefresh} disabled={isLoading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* System Health Card */}
      <Card className="border-2">
        <CardHeader>
          <CardTitle className="text-foreground flex items-center gap-2 text-xl font-bold">
            <Activity className="h-6 w-6" />
            System Health
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="py-8 text-center">
              <Loader2 className="text-primary mx-auto h-12 w-12 animate-spin" />
            </div>
          ) : (
            <div className="flex items-center gap-6">
              <div className={`rounded-2xl ${healthConfig.bgColor} p-6`}>
                <healthConfig.icon className={`h-12 w-12 ${healthConfig.color}`} />
              </div>
              <div>
                <p className="text-foreground/70 mb-1 text-sm font-medium uppercase tracking-wide">
                  Current Status
                </p>
                <p className={`text-4xl font-bold ${healthConfig.color}`}>{healthConfig.label}</p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Account Status Grid */}
      <div>
        <h2 className="text-foreground mb-4 text-xl font-semibold">Account Status</h2>
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-5">
          {accountStatusCards.map((stat) => (
            <Card key={stat.title}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-foreground/60 mb-1 text-xs font-medium">{stat.title}</p>
                    <p className="text-foreground text-3xl font-bold">
                      {isLoading ? '-' : stat.value}
                    </p>
                  </div>
                  <div className={`rounded-lg ${stat.bgColor} p-3`}>
                    <stat.icon className={`h-6 w-6 ${stat.color}`} />
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Token Health Grid */}
      <div>
        <h2 className="text-foreground mb-4 text-xl font-semibold">Token Health</h2>
        <div className="grid gap-6 sm:grid-cols-2">
          {tokenHealthCards.map((stat) => (
            <Card key={stat.title}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-foreground/60 mb-1 text-sm font-medium">{stat.title}</p>
                    <p className="text-foreground mt-2 text-4xl font-bold">
                      {isLoading ? '-' : stat.value}
                      {!isLoading && stat.suffix && (
                        <span className="text-foreground/60 ml-2 text-lg font-normal">
                          {stat.suffix}
                        </span>
                      )}
                    </p>
                  </div>
                  <div className={`rounded-lg ${stat.bgColor} p-4`}>
                    <stat.icon className={`h-8 w-8 ${stat.color}`} />
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Info Card */}
      <Card className="bg-muted/50 border-dashed">
        <CardContent className="pt-6">
          <div className="flex items-start gap-3">
            <TrendingUp className="text-primary mt-0.5 h-5 w-5 flex-shrink-0" />
            <div className="text-foreground/80 text-sm">
              <p className="font-medium">Auto-refresh enabled</p>
              <p className="text-muted-foreground mt-1">
                Statistics automatically refresh every 30 seconds. Use the refresh button above to
                manually update the data.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
