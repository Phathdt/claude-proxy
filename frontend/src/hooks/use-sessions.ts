import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { sessionApi } from '@/lib/api'
import type {
  ListSessionsResponse,
  ListAccountSessionsResponse,
  RevokeSessionResponse,
  RevokeAccountSessionsResponse,
} from '@/types/session'

/**
 * Hook to fetch all sessions (admin)
 */
export function useAllSessions() {
  return useQuery<ListSessionsResponse>({
    queryKey: ['sessions', 'all'],
    queryFn: () => sessionApi.listAll(),
  })
}

/**
 * Hook to fetch sessions by account
 */
export function useAccountSessions(accountId: string | undefined) {
  return useQuery<ListAccountSessionsResponse>({
    queryKey: ['sessions', 'account', accountId],
    queryFn: () => sessionApi.listByAccount(accountId!),
    enabled: !!accountId,
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

/**
 * Hook to revoke all sessions for an account
 */
export function useRevokeAccountSessions() {
  const queryClient = useQueryClient()

  return useMutation<RevokeAccountSessionsResponse, Error, string>({
    mutationFn: (accountId: string) => sessionApi.revokeAccountSessions(accountId),
    onSuccess: (_, accountId) => {
      // Invalidate all session queries
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      // Also invalidate specific account sessions
      queryClient.invalidateQueries({ queryKey: ['sessions', 'account', accountId] })
    },
  })
}
