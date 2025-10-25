import { useState, useEffect } from 'react'
import { authApi } from '@/lib/api'

interface User {
  id: string
  email: string
  name: string
  role: string
}

interface AuthState {
  isAuthenticated: boolean
  isLoading: boolean
  user: User | null
  error: string | null
}

export function useAuth() {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    user: null,
    error: null,
  })

  useEffect(() => {
    validateToken()
  }, [])

  const validateToken = async () => {
    const token = localStorage.getItem('auth_token')

    // Check if token exists and is valid (not 'undefined' or 'null' strings)
    if (!token || token === 'undefined' || token === 'null') {
      setAuthState({
        isAuthenticated: false,
        isLoading: false,
        user: null,
        error: null,
      })
      return
    }

    try {
      const { valid, user } = await authApi.validate(token)

      if (valid && user) {
        // Update stored user info
        localStorage.setItem('user', JSON.stringify(user))
        setAuthState({
          isAuthenticated: true,
          isLoading: false,
          user,
          error: null,
        })
      } else {
        // Invalid token, clear storage
        localStorage.removeItem('auth_token')
        localStorage.removeItem('user')
        setAuthState({
          isAuthenticated: false,
          isLoading: false,
          user: null,
          error: 'Invalid authentication token',
        })
      }
    } catch (error) {
      // Validation failed, clear storage
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user')
      setAuthState({
        isAuthenticated: false,
        isLoading: false,
        user: null,
        error: error instanceof Error ? error.message : 'Authentication failed',
      })
    }
  }

  const logout = async () => {
    await authApi.logout()
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user')
    setAuthState({
      isAuthenticated: false,
      isLoading: false,
      user: null,
      error: null,
    })
  }

  return {
    ...authState,
    logout,
    revalidate: validateToken,
  }
}
