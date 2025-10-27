import { useQuery } from '@tanstack/react-query'
import { statisticsApi } from '@/lib/api'

// Query keys
const QUERY_KEYS = {
  statistics: ['statistics'] as const,
}

// Get system statistics with 30-second auto-refresh
export function useStatistics() {
  return useQuery({
    queryKey: QUERY_KEYS.statistics,
    queryFn: () => statisticsApi.getStatistics(),
    refetchInterval: 30000, // Auto-refresh every 30 seconds
    staleTime: 25000, // Consider data stale after 25 seconds
  })
}
