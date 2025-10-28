import { useEffect } from 'react'
import { useForm, type Resolver } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Dialog } from '@/components/ui/dialog'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@/components/ui/form'
import { useCreateToken, useUpdateToken } from '@/hooks/use-tokens'
import {
  createTokenSchemaWithUniqueCheck,
  updateTokenSchemaWithUniqueCheck,
  type CreateTokenFormData,
} from '@/schemas/token.schema'
import { setFormErrors, getErrorMessage } from '@/lib/form-utils'
import type { Token } from '@/types/token'

interface TokenFormModalProps {
  open: boolean
  onClose: () => void
  token?: Token
  existingTokens?: Token[]
}

export function TokenFormModal({ open, onClose, token, existingTokens = [] }: TokenFormModalProps) {
  const isEditing = !!token
  const createSchema = createTokenSchemaWithUniqueCheck(existingTokens)
  const updateSchema = token
    ? updateTokenSchemaWithUniqueCheck(existingTokens, token.id)
    : createSchema

  const form = useForm<CreateTokenFormData>({
    resolver: zodResolver(isEditing ? updateSchema : createSchema) as Resolver<CreateTokenFormData>,
    defaultValues: {
      name: '',
      key: '',
      status: 'active',
      role: 'user',
    },
  })

  const createMutation = useCreateToken()
  const updateMutation = useUpdateToken()

  useEffect(() => {
    if (open) {
      if (token) {
        form.reset({
          name: token.name,
          key: token.key,
          status: token.status,
          role: token.role,
        })
      } else {
        form.reset({
          name: '',
          key: '',
          status: 'active',
          role: 'user',
        })
      }
    }
  }, [token, open, form])

  const handleSubmit = async (data: CreateTokenFormData) => {
    try {
      if (token) {
        await updateMutation.mutateAsync({
          id: token.id,
          name: data.name,
          key: data.key,
          status: data.status,
          role: data.role || 'user',
        })
      } else {
        await createMutation.mutateAsync(data)
      }
      onClose()
    } catch (error) {
      const errorMessage = getErrorMessage(error)
      setFormErrors(error, form.setError, errorMessage)
    }
  }

  const isLoading = createMutation.isPending || updateMutation.isPending

  return (
    <Dialog open={open} onClose={onClose} title={token ? 'Edit Token' : 'Create Token'}>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
          <FormField
            control={form.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <input
                    type="text"
                    placeholder="Production API Key"
                    className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="key"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Token Value</FormLabel>
                <FormControl>
                  <input
                    type="text"
                    placeholder="sk_prod_1a2b3c4d5e6f7g8h9i0j"
                    className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 font-mono text-sm focus:ring-2 focus:outline-none"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="role"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Role</FormLabel>
                <FormControl>
                  <select
                    className="border-input bg-background text-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
                    {...field}
                  >
                    <option value="user">User</option>
                    <option value="admin">Admin</option>
                  </select>
                </FormControl>
                <p className="text-muted-foreground mt-1 text-xs">
                  Admin tokens can access the admin UI
                </p>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="status"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Status</FormLabel>
                <FormControl>
                  <select
                    className="border-input bg-background text-foreground focus:border-ring focus:ring-ring w-full rounded-md border px-3 py-2 focus:ring-2 focus:outline-none"
                    {...field}
                  >
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="border-input bg-background text-foreground hover:bg-muted flex-1 rounded-md border px-4 py-2 font-medium transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="bg-primary text-primary-foreground hover:bg-primary/90 flex-1 rounded-md px-4 py-2 font-medium transition-colors disabled:opacity-50"
            >
              {isLoading ? 'Saving...' : token ? 'Update' : 'Create'}
            </button>
          </div>
        </form>
      </Form>
    </Dialog>
  )
}
