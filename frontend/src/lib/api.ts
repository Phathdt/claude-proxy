import type { Token, CreateTokenDto, UpdateTokenDto } from '@/types/token'
import type { AppToken, CreateAppTokenDto, UpdateAppTokenDto } from '@/types/app-token'
import { mockTokens } from './mock-data'
import { mockAppTokens } from './mock-app-tokens'

// Simulated delay for API calls
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

let tokens = [...mockTokens]
let appTokens = [...mockAppTokens]

export const tokenApi = {
  getAll: async (): Promise<Token[]> => {
    await delay(500)
    return [...tokens]
  },

  getById: async (id: string): Promise<Token | undefined> => {
    await delay(300)
    return tokens.find((t) => t.id === id)
  },

  create: async (data: CreateTokenDto): Promise<Token> => {
    await delay(800)
    const newToken: Token = {
      id: Math.random().toString(36).substring(7),
      name: data.name,
      key: data.key,
      status: data.status,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      usageCount: 0,
    }
    tokens.push(newToken)
    return newToken
  },

  update: async (data: UpdateTokenDto): Promise<Token> => {
    await delay(600)
    const index = tokens.findIndex((t) => t.id === data.id)
    if (index === -1) throw new Error('Token not found')

    const updatedToken: Token = {
      ...tokens[index],
      name: data.name,
      key: data.key,
      status: data.status,
      updatedAt: new Date().toISOString(),
    }
    tokens[index] = updatedToken
    return updatedToken
  },

  delete: async (id: string): Promise<void> => {
    await delay(500)
    tokens = tokens.filter((t) => t.id !== id)
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
    const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ api_key: apiKey }),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Login failed')
    }

    return response.json()
  },

  // Validate API key
  validate: async (
    apiKey: string
  ): Promise<{ valid: boolean; user?: { id: string; email: string; name: string; role: string } }> => {
    const response = await fetch(`${API_BASE_URL}/api/auth/validate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ api_key: apiKey }),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Validation failed')
    }

    return response.json()
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
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000'

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
    const response = await fetch(`${API_BASE_URL}/oauth/authorize`)
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to get authorization URL')
    }
    return response.json()
  },

  // Exchange authorization code for tokens
  exchangeCode: async (data: OAuthExchangeRequest): Promise<OAuthExchangeResponse> => {
    const response = await fetch(`${API_BASE_URL}/oauth/exchange`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to exchange code')
    }
    return response.json()
  },

  // Get health status (includes account info)
  getHealth: async (): Promise<HealthResponse> => {
    const response = await fetch(`${API_BASE_URL}/health`)
    if (!response.ok) {
      throw new Error('Failed to get health status')
    }
    return response.json()
  },
}
