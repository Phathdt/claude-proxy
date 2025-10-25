export interface Token {
  id: string
  name: string
  key: string
  status: 'active' | 'inactive'
  createdAt: number
  updatedAt: number
  usageCount: number
  lastUsedAt?: number
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
