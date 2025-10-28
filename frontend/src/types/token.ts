export interface Token {
  id: string
  name: string
  key: string
  status: 'active' | 'inactive'
  role: 'user' | 'admin'
  createdAt: string // RFC3339/ISO 8601 datetime
  updatedAt: string // RFC3339/ISO 8601 datetime
  usageCount: number
  lastUsedAt?: string // RFC3339/ISO 8601 datetime
}

export interface CreateTokenDto {
  name: string
  key: string
  status: 'active' | 'inactive'
  role?: 'user' | 'admin' // Optional, defaults to 'user'
}

export interface UpdateTokenDto {
  id: string
  name: string
  key: string
  status: 'active' | 'inactive'
  role: 'user' | 'admin'
}

export interface TokenQueryParams {
  role?: 'user' | 'admin' | ''
  status?: 'active' | 'inactive' | ''
  search?: string
  page?: number
  limit?: number
}

export interface Paging {
  page: number
  limit: number
  total: number
  cursor?: string
  next_cursor?: string
}

export interface TokenListResponse {
  tokens: Token[]
  paging: Paging
}
