import type { Permission, SessionUser } from '../types'

const API_BASE = import.meta.env.VITE_IMSI_API_BASE ?? 'http://127.0.0.1:8080'
const useMock = import.meta.env.MODE === 'test' || import.meta.env.VITE_IMSI_DATA_MODE === 'mock'

export const mockUser: SessionUser = {
  id: 'usr_demo',
  bank_id: 'bank-demo',
  username: 'admin',
  email: 'admin@imsi.local',
  display_name: 'Operations Admin',
  roles: ['platform_admin'],
  permissions: [
    'dashboard:read',
    'transactions:read',
    'transactions:trace',
    'providers:manage',
    'policy:draft',
    'policy:approve',
    'policy:activate',
    'incidents:manage',
    'reconciliation:manage',
    'fx:read',
    'audit:read',
    'audit:export',
    'users:manage',
    'identity:manage',
  ],
  auth_provider: 'mock',
}

export async function currentUser() {
  if (useMock) return mockUser
  const response = await fetch(`${API_BASE}/v1/auth/me`, { credentials: 'include' })
  if (response.status === 401) return null
  if (!response.ok) throw new Error('Unable to load current user')
  return (await response.json()) as SessionUser
}

export async function loginLocal(bankId: string, username: string, password: string) {
  if (useMock) return mockUser
  const response = await fetch(`${API_BASE}/v1/auth/login`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bank_id: bankId, username, password }),
  })
  if (!response.ok) throw new Error('Login failed')
  return (await response.json()) as SessionUser
}

export async function loginLDAP(bankId: string, username: string, password: string) {
  if (useMock) return { ...mockUser, auth_provider: 'ldap:mock' }
  const response = await fetch(`${API_BASE}/v1/auth/ldap/login`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bank_id: bankId, username, password }),
  })
  if (!response.ok) throw new Error('LDAP login failed')
  return (await response.json()) as SessionUser
}

export async function logout() {
  if (useMock) return
  await fetch(`${API_BASE}/v1/auth/logout`, { method: 'POST', credentials: 'include' })
}

export function hasPermission(user: SessionUser | null, permission: Permission) {
  return Boolean(user?.permissions.includes(permission))
}

