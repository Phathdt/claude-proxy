import { Activity, Key, TrendingUp, Users, Loader2 } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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
          <Card key={stat.title}>
            <CardContent className="pt-6">
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
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Recent Tokens */}
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-4">
            <h2 className="text-card-foreground text-xl font-semibold">Recent Tokens</h2>
            {isLoading ? (
              <div className="py-8 text-center">
                <Loader2 className="text-primary mx-auto h-8 w-8 animate-spin" />
                <p className="text-muted-foreground mt-2">Loading tokens...</p>
              </div>
            ) : tokens && tokens.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Usage</TableHead>
                    <TableHead>Last Used</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tokens.slice(0, 5).map((token) => (
                    <TableRow key={token.id}>
                      <TableCell className="font-medium">{token.name}</TableCell>
                      <TableCell>
                        <span
                          className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                            token.status === 'active'
                              ? 'bg-green-500/10 text-green-500'
                              : 'bg-gray-500/10 text-gray-500'
                          }`}
                        >
                          {token.status}
                        </span>
                      </TableCell>
                      <TableCell>{token.usageCount.toLocaleString()}</TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {token.lastUsedAt
                          ? new Date(token.lastUsedAt * 1000).toLocaleDateString()
                          : 'Never used'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <p className="text-muted-foreground py-8 text-center text-sm">No tokens found</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
