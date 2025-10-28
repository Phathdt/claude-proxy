import { z } from 'zod'

export const statusEnum = z.enum(['active', 'inactive'])

export const createTokenSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .min(3, 'Name must be at least 3 characters')
    .max(100, 'Name must be at most 100 characters'),
  key: z
    .string()
    .min(1, 'Token value is required'),
  status: statusEnum.default('active'),
})

export const updateTokenSchema = createTokenSchema.extend({
  id: z.string().min(1, 'ID is required'),
})

export type CreateTokenFormData = {
  name: string
  key: string
  status: 'active' | 'inactive'
}

export type UpdateTokenFormData = CreateTokenFormData & {
  id: string
}

// Validation helpers for unique constraints
export function createTokenSchemaWithUniqueCheck(
  existingTokens: Array<{ id: string; name: string; key: string }>,
  editingTokenId?: string
) {
  return createTokenSchema
    .refine(
      (data) => {
        const isDuplicate = existingTokens.some(
          (t) => t.name.toLowerCase() === data.name.toLowerCase() && t.id !== editingTokenId
        )
        return !isDuplicate
      },
      {
        message: 'Token name must be unique',
        path: ['name'],
      }
    )
    .refine(
      (data) => {
        const isDuplicate = existingTokens.some(
          (t) => t.key === data.key && t.id !== editingTokenId
        )
        return !isDuplicate
      },
      {
        message: 'API key must be unique',
        path: ['key'],
      }
    )
}

export function updateTokenSchemaWithUniqueCheck(
  existingTokens: Array<{ id: string; name: string; key: string }>,
  editingTokenId: string
) {
  return updateTokenSchema
    .refine(
      (data) => {
        const isDuplicate = existingTokens.some(
          (t) => t.name.toLowerCase() === data.name.toLowerCase() && t.id !== editingTokenId
        )
        return !isDuplicate
      },
      {
        message: 'Token name must be unique',
        path: ['name'],
      }
    )
    .refine(
      (data) => {
        const isDuplicate = existingTokens.some(
          (t) => t.key === data.key && t.id !== editingTokenId
        )
        return !isDuplicate
      },
      {
        message: 'API key must be unique',
        path: ['key'],
      }
    )
}
