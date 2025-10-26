import { z } from 'zod'

export const loginSchema = z.object({
  apiKey: z.string().min(1, 'API Key is required'),
})

export type LoginFormData = z.infer<typeof loginSchema>
