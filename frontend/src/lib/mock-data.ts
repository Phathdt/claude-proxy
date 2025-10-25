import type { Token } from '@/types/token'

export const mockTokens: Token[] = [
  {
    id: '1',
    name: 'Production API Key',
    key: 'sk_prod_1a2b3c4d5e6f7g8h9i0j',
    status: 'active',
    createdAt: Math.floor(new Date('2025-01-15T10:30:00Z').getTime() / 1000),
    updatedAt: Math.floor(new Date('2025-01-20T14:20:00Z').getTime() / 1000),
    usageCount: 1542,
    lastUsedAt: Math.floor(new Date('2025-10-23T12:00:00Z').getTime() / 1000),
  },
  {
    id: '2',
    name: 'Development API Key',
    key: 'sk_dev_9i8h7g6f5e4d3c2b1a0j',
    status: 'active',
    createdAt: Math.floor(new Date('2025-01-10T09:15:00Z').getTime() / 1000),
    updatedAt: Math.floor(new Date('2025-01-18T11:45:00Z').getTime() / 1000),
    usageCount: 342,
    lastUsedAt: Math.floor(new Date('2025-10-22T18:30:00Z').getTime() / 1000),
  },
  {
    id: '3',
    name: 'Testing API Key',
    key: 'sk_test_0j9i8h7g6f5e4d3c2b1a',
    status: 'inactive',
    createdAt: Math.floor(new Date('2025-01-05T08:00:00Z').getTime() / 1000),
    updatedAt: Math.floor(new Date('2025-01-12T16:30:00Z').getTime() / 1000),
    usageCount: 89,
    lastUsedAt: Math.floor(new Date('2025-01-12T10:15:00Z').getTime() / 1000),
  },
  {
    id: '4',
    name: 'Staging API Key',
    key: 'sk_staging_a1b2c3d4e5f6g7h8i9j0',
    status: 'active',
    createdAt: Math.floor(new Date('2025-01-08T13:45:00Z').getTime() / 1000),
    updatedAt: Math.floor(new Date('2025-01-19T09:00:00Z').getTime() / 1000),
    usageCount: 756,
    lastUsedAt: Math.floor(new Date('2025-10-23T08:45:00Z').getTime() / 1000),
  },
]
