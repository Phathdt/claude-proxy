import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { tokenApi } from '@/lib/api'
import type { CreateTokenDto, UpdateTokenDto, TokenQueryParams } from '@/types/token'

export const useTokens = (params?: TokenQueryParams) => {
  return useQuery({
    queryKey: ['tokens', params],
    queryFn: () => tokenApi.getAll(params),
  })
}

export const useToken = (id: string) => {
  return useQuery({
    queryKey: ['tokens', id],
    queryFn: () => tokenApi.getById(id),
    enabled: !!id,
  })
}

export const useCreateToken = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateTokenDto) => tokenApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
    },
  })
}

export const useUpdateToken = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateTokenDto) => tokenApi.update(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
    },
  })
}

export const useDeleteToken = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => tokenApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tokens'] })
    },
  })
}
