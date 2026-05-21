import type { CountryIdentity, ProviderIdentity } from '../types'

export const providerIdentities: Record<string, ProviderIdentity> = {
  thunes: {
    id: 'thunes',
    name: 'Thunes',
    shortName: 'TH',
    mark: 'Th',
    category: 'B2B payout network',
    tone: 'healthy',
    color: '#22d3a4',
  },
  remitly: {
    id: 'remitly',
    name: 'Remitly',
    shortName: 'RM',
    mark: 'Re',
    category: 'Digital IMTO',
    tone: 'watch',
    color: '#38bdf8',
  },
  ria: {
    id: 'ria',
    name: 'Ria',
    shortName: 'RIA',
    mark: 'Ri',
    category: 'Legacy IMTO',
    tone: 'degraded',
    color: '#fb7185',
  },
  papss: {
    id: 'papss',
    name: 'PAPSS',
    shortName: 'PS',
    mark: 'Pa',
    category: 'Pan-African rail',
    tone: 'recovery',
    color: '#a78bfa',
  },
  nip: {
    id: 'nip',
    name: 'NIP',
    shortName: 'NIP',
    mark: 'Ni',
    category: 'Local payout rail',
    tone: 'healthy',
    color: '#34d399',
  },
  'manual-review': {
    id: 'manual-review',
    name: 'Manual review',
    shortName: 'MR',
    mark: 'Mr',
    category: 'Operational queue',
    tone: 'unknown',
    color: '#94a3b8',
  },
}

export function flagForCountryCode(code: string) {
  const normalized = code.trim().toUpperCase()
  if (!/^[A-Z]{2}$/.test(normalized) || normalized === 'EU') return normalized || '??'
  return [...normalized].map((letter) => String.fromCodePoint(127397 + letter.charCodeAt(0))).join('')
}

export const countryIdentities: Record<string, CountryIdentity> = {
  eu: { code: 'EU', name: 'European Union', shortName: 'Europe', flag: 'EU' },
  europe: { code: 'EU', name: 'European Union', shortName: 'Europe', flag: 'EU' },
  'european union': { code: 'EU', name: 'European Union', shortName: 'Europe', flag: 'EU' },
  germany: { code: 'DE', name: 'Germany', shortName: 'Germany', flag: flagForCountryCode('DE') },
  nigeria: { code: 'NG', name: 'Nigeria', shortName: 'Nigeria', flag: flagForCountryCode('NG') },
  kenya: { code: 'KE', name: 'Kenya', shortName: 'Kenya', flag: flagForCountryCode('KE') },
  'united kingdom': { code: 'GB', name: 'United Kingdom', shortName: 'United Kingdom', flag: flagForCountryCode('GB') },
  uk: { code: 'GB', name: 'United Kingdom', shortName: 'United Kingdom', flag: flagForCountryCode('GB') },
  gb: { code: 'GB', name: 'United Kingdom', shortName: 'United Kingdom', flag: flagForCountryCode('GB') },
  'united states': { code: 'US', name: 'United States', shortName: 'United States', flag: flagForCountryCode('US') },
  us: { code: 'US', name: 'United States', shortName: 'United States', flag: flagForCountryCode('US') },
  ng: { code: 'NG', name: 'Nigeria', shortName: 'Nigeria', flag: flagForCountryCode('NG') },
  ke: { code: 'KE', name: 'Kenya', shortName: 'Kenya', flag: flagForCountryCode('KE') },
}

const normalize = (value: string) => value.trim().toLowerCase()

export const providerKey = (provider: string) => {
  const firstToken = provider.split('->')[0]?.split('/')[0]?.trim() ?? provider
  const normalized = normalize(firstToken)
  if (normalized.includes('thunes')) return 'thunes'
  if (normalized.includes('remitly')) return 'remitly'
  if (normalized.includes('ria')) return 'ria'
  if (normalized.includes('papss')) return 'papss'
  if (normalized.includes('nip')) return 'nip'
  if (normalized.includes('manual')) return 'manual-review'
  return normalized.replace(/[^a-z0-9]+/g, '-')
}

export const getProviderIdentity = (provider: string): ProviderIdentity => {
  const key = providerKey(provider)
  return (
    providerIdentities[key] ?? {
      id: key,
      name: provider,
      shortName: provider.slice(0, 3).toUpperCase(),
      mark: provider.slice(0, 2),
      category: 'Provider',
      tone: 'unknown',
      color: '#94a3b8',
    }
  )
}

export const getCountryIdentity = (country: string): CountryIdentity => {
  const key = normalize(country)
  const inferredCode = country.trim().slice(0, 2).toUpperCase()
  return countryIdentities[key] ?? { code: inferredCode, name: country, shortName: country, flag: flagForCountryCode(inferredCode) }
}

export const parseCorridor = (corridor: string) => {
  const normalized = corridor.replace(/\s+-\>\s+|\s+to\s+/i, ' -> ')
  const [origin = normalized, destination = ''] = normalized.split('->').map((part) => part.trim())
  return { origin, destination }
}

