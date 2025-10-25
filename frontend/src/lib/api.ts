import axios from 'axios'
import type { Token, CreateTokenDto, UpdateTokenDto } from '@/types/token'
import type { AppToken, CreateAppTokenDto, UpdateAppTokenDto } from '@/types/app-token'
import { mockAppTokens } from './mock-app-tokens'
import { convertKeysToSnake, convertKeysToCamel } from './case-converter'

// Simulated delay for API calls
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

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
    if (token) {
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

let appTokens = [...mockAppTokens]

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
  // Login with API key
  login: async (apiKey: string): Promise<LoginResponse> => {
    const response = await apiClient.post('/api/auth/login', { api_key: apiKey })
    return response.data
  },

  // Validate API key
  validate: async (
    apiKey: string
  ): Promise<{ valid: boolean; user?: { id: string; email: string; name: string; role: string } }> => {
    const response = await apiClient.post('/api/auth/validate', { api_key: apiKey })
    return response.data
  },

  logout: async (): Promise<void> => {
    // Just clear local storage, no backend call needed
    await delay(300)
  },
}

// App Token API
export const appTokenApi = {
  getAll: async (): Promise<AppToken[]> => {
    await delay(500)
    return [...appTokens]
  },

  getById: async (id: string): Promise<AppToken | undefined> => {
    await delay(300)
    return appTokens.find((t) => t.id === id)
  },

  create: async (data: CreateAppTokenDto): Promise<AppToken> => {
    await delay(800)
    const newToken: AppToken = {
      id: Math.random().toString(36).substring(7),
      name: data.name,
      email: data.email,
      orgId: data.orgId,
      type: data.type,
      accountType: data.accountType,
      status: data.status,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      usageCount: 0,
    }
    appTokens.push(newToken)
    return newToken
  },

  update: async (data: UpdateAppTokenDto): Promise<AppToken> => {
    await delay(600)
    const index = appTokens.findIndex((t) => t.id === data.id)
    if (index === -1) throw new Error('App token not found')

    const updatedToken: AppToken = {
      ...appTokens[index],
      name: data.name,
      email: data.email,
      orgId: data.orgId,
      type: data.type,
      accountType: data.accountType,
      status: data.status,
      updatedAt: new Date().toISOString(),
    }
    appTokens[index] = updatedToken
    return updatedToken
  },

  delete: async (id: string): Promise<void> => {
    await delay(500)
    appTokens = appTokens.filter((t) => t.id !== id)
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

// App Accounts API (multi-account OAuth management)
export interface AppAccount {
  id: string
  name: string
  organization_uuid: string
  access_token?: string
  refresh_token?: string
  expires_at: number
  status: string
  created_at: number
  updated_at: number
}

export interface CreateAppAccountRequest {
  name: string
  org_id?: string
}

export interface CreateAppAccountResponse {
  authorization_url: string
  state: string
  code_verifier: string
}

export interface CompleteAppAccountRequest {
  name: string
  code: string
  state: string
  code_verifier: string
  org_id?: string
}

export interface CompleteAppAccountResponse {
  success: boolean
  message: string
  account: AppAccount
}

export interface UpdateAppAccountRequest {
  name?: string
  status?: string
}

export const appAccountsApi = {
  // Start OAuth flow - returns authorization URL
  create: async (data: CreateAppAccountRequest): Promise<CreateAppAccountResponse> => {
    const response = await apiClient.post('/api/app-accounts', data)
    return response.data
  },

  // Complete OAuth flow - exchange code for tokens
  complete: async (data: CompleteAppAccountRequest): Promise<CompleteAppAccountResponse> => {
    const response = await apiClient.post('/api/app-accounts/complete', data)
    return response.data
  },

  // List all app accounts
  list: async (): Promise<AppAccount[]> => {
    const response = await apiClient.get('/api/app-accounts')
    return response.data.accounts || []
  },

  // Get single app account
  get: async (id: string): Promise<AppAccount> => {
    const response = await apiClient.get(`/api/app-accounts/${id}`)
    return response.data.account
  },

  // Update app account
  update: async (id: string, data: UpdateAppAccountRequest): Promise<AppAccount> => {
    const response = await apiClient.put(`/api/app-accounts/${id}`, data)
    return response.data.account
  },

  // Delete app account
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/app-accounts/${id}`)
  },
}
