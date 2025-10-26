import { z } from 'zod'

export const accountType = z.enum(['oauth', 'cookies'])
export const accountTypeEnum = z.enum(['pro', 'max'])

export const createAccountSchema = z.object({
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
  type: accountType.default('oauth'),
  accountType: accountTypeEnum.default('pro'),
})

export const updateAccountSchema = createAccountSchema.extend({
  id: z.string().min(1, 'ID is required'),
})

export type CreateAccountFormData = z.infer<typeof createAccountSchema>
export type UpdateAccountFormData = z.infer<typeof updateAccountSchema>
export type AccountType = z.infer<typeof accountType>
export type AccountTypeEnum = z.infer<typeof accountTypeEnum>

// Validation helpers for unique constraints
export function createAccountSchemaWithUniqueCheck(
  existingAccounts: Array<{ id: string; name: string; organizationUuid: string }>,
  editingAccountId?: string
) {
  return createAccountSchema
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

export function updateAccountSchemaWithUniqueCheck(
  existingAccounts: Array<{ id: string; name: string; organizationUuid: string }>,
  editingAccountId: string
) {
  return updateAccountSchema
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
