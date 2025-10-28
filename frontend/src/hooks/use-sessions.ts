import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { sessionApi } from '@/lib/api'
import type { ListSessionsResponse, RevokeSessionResponse } from '@/types/session'

/**
 * Hook to fetch all sessions (admin)
 * Sessions track concurrent requests per client (IP + User-Agent)
 */
export function useAllSessions() {
  return useQuery<ListSessionsResponse>({
    queryKey: ['sessions', 'all'],
    queryFn: () => sessionApi.listAll(),
  })
}

/**
 * Hook to revoke a session
 */
export function useRevokeSession() {
  const queryClient = useQueryClient()

  return useMutation<RevokeSessionResponse, Error, string>({
    mutationFn: (sessionId: string) => sessionApi.revoke(sessionId),
    onSuccess: () => {
      // Invalidate all session queries to refetch
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
    },
  })
}
