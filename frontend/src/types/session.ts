/**
 * Session entity representing an active client session
 * Sessions track concurrent requests per client (IP + User-Agent)
 */
export interface Session {
  id: string
  tokenId: string
  userAgent: string
  ipAddress: string
  createdAt: string
  lastSeenAt: string
  expiresAt: string
  isActive: boolean
  requestPath: string
}

/**
 * Response from list all sessions endpoint
 */
export interface ListSessionsResponse {
  sessions: Session[]
  total: number
}

/**
 * Response from revoke session endpoint
 */
export interface RevokeSessionResponse {
  success: boolean
  message: string
}
