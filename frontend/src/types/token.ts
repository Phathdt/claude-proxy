export interface Token {
  id: string
  name: string
  key: string
  status: 'active' | 'inactive'
  createdAt: string
  updatedAt: string
  usageCount: number
  lastUsedAt?: string
}

export interface CreateTokenDto {
  name: string
  key: string
  status: 'active' | 'inactive'
}

export interface UpdateTokenDto {
  id: string
  name: string
  key: string
  status: 'active' | 'inactive'
}
