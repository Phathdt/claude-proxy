import axios from 'axios'
import type { Token, CreateTokenDto, UpdateTokenDto } from '@/types/token'
import { convertKeysToSnake, convertKeysToCamel } from './case-converter'

// API base URL
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000'

// Axios instance with default config
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token and convert camelCase to snake_case
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    // Only set header if token exists and is not empty
    if (token && token !== 'undefined' && token !== 'null') {
      config.headers['X-API-Key'] = token
    }

    // Convert request data from camelCase to snake_case
    if (config.data) {
      config.data = convertKeysToSnake(config.data)
    }

    // Convert query params from camelCase to snake_case
    if (config.params) {
      config.params = convertKeysToSnake(config.params)
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling and convert snake_case to camelCase
apiClient.interceptors.response.use(
  (response) => {
    // Convert response data from snake_case to camelCase
    if (response.data) {
      response.data = convertKeysToCamel(response.data)
    }
    return response
  },
  (error) => {
    // Convert error response data from snake_case to camelCase
    if (error.response?.data) {
      error.response.data = convertKeysToCamel(error.response.data)
    }

    if (error.response?.data?.error?.message) {
      throw new Error(error.response.data.error.message)
    } else if (error.response?.data?.error) {
      throw new Error(error.response.data.error)
    } else if (error.message) {
      throw new Error(error.message)
    } else {
      throw new Error('An unexpected error occurred')
    }
  }
)

// Token API (real API calls to backend)
export const tokenApi = {
  getAll: async (): Promise<Token[]> => {
    const response = await apiClient.get('/api/tokens')
    return response.data.tokens || []
  },

  getById: async (id: string): Promise<Token | undefined> => {
    const response = await apiClient.get(`/api/tokens/${id}`)
    return response.data.token
  },

  create: async (data: CreateTokenDto): Promise<Token> => {
    const response = await apiClient.post('/api/tokens', data)
    return response.data.token
  },

  update: async (data: UpdateTokenDto): Promise<Token> => {
    const response = await apiClient.put(`/api/tokens/${data.id}`, {
      name: data.name,
      key: data.key,
      status: data.status,
    })
    return response.data.token
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/tokens/${id}`)
  },

  generateKey: async (): Promise<string> => {
    const response = await apiClient.post('/api/tokens/generate-key')
    return response.data.key
  },
}

// Auth API (real API calls to backend)
export interface LoginResponse {
  success: boolean
  token: string
  user: {
    id: string
    email: string
    name: string
    role: string
  }
}

export const authApi = {
  // Login with email and password
  login: async (credentials: { email: string; password: string }): Promise<LoginResponse> => {
    const response = await apiClient.post('/api/auth/login', {
      email: credentials.email,
      password: credentials.password,
    })
    return response.data
  },

  // Validate API key
  validate: async (
    apiKey: string
  ): Promise<{
    valid: boolean
    user?: { id: string; email: string; name: string; role: string }
  }> => {
    const response = await apiClient.post('/api/auth/validate', { api_key: apiKey })
    return response.data
  },

  logout: async (): Promise<void> => {
    // Just clear local storage, no backend call needed
    localStorage.removeItem('auth_token')
  },
}

// OAuth API (real API calls to backend)
export interface OAuthAuthorizeResponse {
  authorization_url: string
  state: string
  code_verifier: string
}

export interface OAuthExchangeRequest {
  code: string
  state: string
  code_verifier: string
  org_id?: string
}

export interface OAuthExchangeResponse {
  success: boolean
  message: string
  organization_uuid: string
  expires_at: number
}

export interface HealthResponse {
  status: string
  timestamp: number
  account: {
    account_valid: boolean
    expires_at?: number
    organization?: string
  }
}

export const oauthApi = {
  // Generate OAuth authorization URL
  getAuthorizeUrl: async (): Promise<OAuthAuthorizeResponse> => {
    const response = await apiClient.get('/oauth/authorize')
    return response.data
  },

  // Exchange authorization code for tokens
  exchangeCode: async (data: OAuthExchangeRequest): Promise<OAuthExchangeResponse> => {
    const response = await apiClient.post('/oauth/exchange', data)
    return response.data
  },

  // Get health status (includes account info)
  getHealth: async (): Promise<HealthResponse> => {
    const response = await apiClient.get('/health')
    return response.data
  },
}

// Accounts API (multi-account OAuth management)
export interface Account {
  id: string
  name: string
  organizationUuid: string
  accessToken?: string
  refreshToken?: string
  expiresAt: number
  status: string
  createdAt: number
  updatedAt: number
}

export interface CreateAccountRequest {
  name: string
  orgId?: string
}

export interface CreateAccountResponse {
  authorizationUrl: string
  state: string
  codeVerifier: string
}

export interface CompleteAccountRequest {
  name: string
  code: string
  state: string
  codeVerifier: string
  orgId?: string
}

export interface CompleteAccountResponse {
  success: boolean
  message: string
  account: Account
}

export interface UpdateAccountRequest {
  name?: string
  status?: string
}

export const accountsApi = {
  // List all accounts
  list: async (): Promise<Account[]> => {
    const response = await apiClient.get('/api/accounts')
    return response.data.accounts || []
  },

  // Get single account
  get: async (id: string): Promise<Account> => {
    const response = await apiClient.get(`/api/accounts/${id}`)
    return response.data.account
  },

  // Create account (start OAuth flow)
  create: async (data: CreateAccountRequest): Promise<CreateAccountResponse> => {
    const params = new URLSearchParams()
    if (data.orgId) {
      params.append('org_id', data.orgId)
    }
    const response = await apiClient.get(`/oauth/authorize?${params.toString()}`)
    return response.data
  },

  // Complete OAuth flow
  complete: async (data: CompleteAccountRequest): Promise<CompleteAccountResponse> => {
    const response = await apiClient.post('/oauth/exchange', data)
    return response.data
  },

  // Update account
  update: async (id: string, data: UpdateAccountRequest): Promise<Account> => {
    const response = await apiClient.put(`/api/accounts/${id}`, data)
    return response.data.account
  },

  // Delete account
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/accounts/${id}`)
  },
}
