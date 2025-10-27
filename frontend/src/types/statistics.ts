export type SystemHealth = 'healthy' | 'degraded' | 'unhealthy'

export interface Statistics {
  totalAccounts: number
  activeAccounts: number
  inactiveAccounts: number
  rateLimitedAccounts: number
  invalidAccounts: number
  accountsNeedingRefresh: number
  oldestTokenAgeHours: number
  systemHealth: SystemHealth
}
