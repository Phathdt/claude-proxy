import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { appTokenApi } from '@/lib/api'
import type { CreateAppTokenDto, UpdateAppTokenDto } from '@/types/app-token'

export const useAppTokens = () => {
  return useQuery({
    queryKey: ['appTokens'],
    queryFn: appTokenApi.getAll,
  })
}

export const useAppToken = (id: string) => {
  return useQuery({
    queryKey: ['appTokens', id],
    queryFn: () => appTokenApi.getById(id),
    enabled: !!id,
  })
}

export const useCreateAppToken = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateAppTokenDto) => appTokenApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appTokens'] })
    },
  })
}

export const useUpdateAppToken = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateAppTokenDto) => appTokenApi.update(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appTokens'] })
    },
  })
}

export const useDeleteAppToken = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => appTokenApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appTokens'] })
    },
  })
}
