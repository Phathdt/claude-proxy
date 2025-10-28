/**
 * Session entity representing an active user session
 */
export interface Session {
  id: string;
  accountId: string;
  tokenId: string;
  userAgent: string;
  ipAddress: string;
  createdAt: string;
  lastSeenAt: string;
  expiresAt: string;
  isActive: boolean;
  requestPath: string;
}

/**
 * Response from list all sessions endpoint
 */
export interface ListSessionsResponse {
  sessions: Session[];
  total: number;
}

/**
 * Response from list account sessions endpoint
 */
export interface ListAccountSessionsResponse {
  accountId: string;
  sessions: Session[];
  total: number;
}

/**
 * Response from revoke session endpoint
 */
export interface RevokeSessionResponse {
  sessionId: string;
  message: string;
}

/**
 * Response from revoke account sessions endpoint
 */
export interface RevokeAccountSessionsResponse {
  accountId: string;
  revokedCount: number;
  message: string;
}
