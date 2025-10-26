import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accountsApi } from '@/lib/api'
import type { CreateAccountRequest, CompleteAccountRequest, UpdateAccountRequest } from '@/lib/api'

// Query keys
const QUERY_KEYS = {
  accounts: ['accounts'] as const,
  account: (id: string) => ['account', id] as const,
}

// List all accounts
export function useAccounts() {
  return useQuery({
    queryKey: QUERY_KEYS.accounts,
    queryFn: () => accountsApi.list(),
  })
}

// Get single account
export function useAccount(id: string) {
  return useQuery({
    queryKey: QUERY_KEYS.account(id),
    queryFn: () => accountsApi.get(id),
    enabled: !!id,
  })
}

// Create account (start OAuth)
export function useCreateAccount() {
  return useMutation({
    mutationFn: (data: CreateAccountRequest) => accountsApi.create(data),
    onSuccess: () => {
      // Don't invalidate yet - wait for completion
    },
  })
}

// Complete OAuth flow
export function useCompleteAccount() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CompleteAccountRequest) => accountsApi.complete(data),
    onSuccess: () => {
      // Invalidate and refetch accounts list
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.accounts })
    },
  })
}

// Update account
export function useUpdateAccount() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAccountRequest }) =>
      accountsApi.update(id, data),
    onSuccess: (_, variables) => {
      // Invalidate both list and single account
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.accounts })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.account(variables.id) })
    },
  })
}

// Delete account
export function useDeleteAccount() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => accountsApi.delete(id),
    onSuccess: () => {
      // Invalidate accounts list
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.accounts })
    },
  })
}
