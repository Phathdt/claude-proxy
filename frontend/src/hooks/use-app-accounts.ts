import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { appAccountsApi } from '@/lib/api'
import type {
  CreateAppAccountRequest,
  CompleteAppAccountRequest,
  UpdateAppAccountRequest,
} from '@/lib/api'

// Query keys
const QUERY_KEYS = {
  appAccounts: ['appAccounts'] as const,
  appAccount: (id: string) => ['appAccount', id] as const,
}

// List all app accounts
export function useAppAccounts() {
  return useQuery({
    queryKey: QUERY_KEYS.appAccounts,
    queryFn: () => appAccountsApi.list(),
  })
}

// Get single app account
export function useAppAccount(id: string) {
  return useQuery({
    queryKey: QUERY_KEYS.appAccount(id),
    queryFn: () => appAccountsApi.get(id),
    enabled: !!id,
  })
}

// Create app account (start OAuth)
export function useCreateAppAccount() {
  return useMutation({
    mutationFn: (data: CreateAppAccountRequest) => appAccountsApi.create(data),
    onSuccess: () => {
      // Don't invalidate yet - wait for completion
    },
  })
}

// Complete OAuth flow
export function useCompleteAppAccount() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CompleteAppAccountRequest) => appAccountsApi.complete(data),
    onSuccess: () => {
      // Invalidate and refetch accounts list
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.appAccounts })
    },
  })
}

// Update app account
export function useUpdateAppAccount() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAppAccountRequest }) =>
      appAccountsApi.update(id, data),
    onSuccess: (_, variables) => {
      // Invalidate both list and single account
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.appAccounts })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.appAccount(variables.id) })
    },
  })
}

// Delete app account
export function useDeleteAppAccount() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => appAccountsApi.delete(id),
    onSuccess: () => {
      // Invalidate accounts list
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.appAccounts })
    },
  })
}
