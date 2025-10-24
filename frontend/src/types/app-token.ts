export interface AppToken {
  id: string
  name: string
  email: string
  orgId: string
  type: 'oauth' | 'cookies'
  state?: string
  accountType: 'pro' | 'max'
  status: 'active' | 'inactive'
  createdAt: string
  updatedAt: string
  usageCount: number
  lastUsedAt?: string
  resetTime?: string
}

export interface CreateAppTokenDto {
  name: string
  email: string
  orgId: string
  type: 'oauth' | 'cookies'
  accountType: 'pro' | 'max'
  status: 'active' | 'inactive'
}

export interface UpdateAppTokenDto {
  id: string
  name: string
  email: string
  orgId: string
  type: 'oauth' | 'cookies'
  accountType: 'pro' | 'max'
  status: 'active' | 'inactive'
}
