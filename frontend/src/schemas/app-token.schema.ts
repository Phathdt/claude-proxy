import { z } from 'zod'

export const appTokenType = z.enum(['oauth', 'cookies'])
export const appTokenAccountType = z.enum(['pro', 'max'])

export const createAppTokenSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .min(3, 'Name must be at least 3 characters')
    .max(100, 'Name must be at most 100 characters'),
  email: z.string().email('Invalid email address').min(1, 'Email is required'),
  orgId: z
    .string()
    .min(1, 'Organization ID is required')
    .min(3, 'Organization ID must be at least 3 characters'),
  type: appTokenType.default('oauth'),
  accountType: appTokenAccountType.default('pro'),
})

export const updateAppTokenSchema = createAppTokenSchema.extend({
  id: z.string().min(1, 'ID is required'),
})

export type CreateAppTokenFormData = z.infer<typeof createAppTokenSchema>
export type UpdateAppTokenFormData = z.infer<typeof updateAppTokenSchema>
export type AppTokenType = z.infer<typeof appTokenType>
export type AppTokenAccountType = z.infer<typeof appTokenAccountType>

// Validation helpers for unique constraints
export function createAppTokenSchemaWithUniqueCheck(
  existingAccounts: Array<{ id: string; name: string; organizationUuid: string }>,
  editingAccountId?: string
) {
  return createAppTokenSchema
    .refine(
      (data) => {
        const isDuplicate = existingAccounts.some(
          (a) => a.name.toLowerCase() === data.name.toLowerCase() && a.id !== editingAccountId
        )
        return !isDuplicate
      },
      {
        message: 'Account name must be unique',
        path: ['name'],
      }
    )
    .refine(
      (data) => {
        const isDuplicate = existingAccounts.some(
          (a) => a.organizationUuid === data.orgId && a.id !== editingAccountId
        )
        return !isDuplicate
      },
      {
        message: 'Organization ID must be unique',
        path: ['orgId'],
      }
    )
}

export function updateAppTokenSchemaWithUniqueCheck(
  existingAccounts: Array<{ id: string; name: string; organizationUuid: string }>,
  editingAccountId: string
) {
  return updateAppTokenSchema
    .refine(
      (data) => {
        const isDuplicate = existingAccounts.some(
          (a) => a.name.toLowerCase() === data.name.toLowerCase() && a.id !== editingAccountId
        )
        return !isDuplicate
      },
      {
        message: 'Account name must be unique',
        path: ['name'],
      }
    )
    .refine(
      (data) => {
        const isDuplicate = existingAccounts.some(
          (a) => a.organizationUuid === data.orgId && a.id !== editingAccountId
        )
        return !isDuplicate
      },
      {
        message: 'Organization ID must be unique',
        path: ['orgId'],
      }
    )
}
