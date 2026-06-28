import { alovaInstance } from '../lib/alova'

export interface UsageOverview {
  balance: number
  used_quota: number
  request_count: number
  today: {
    requests: number
    tokens: number
    cost: number
  }
  this_month: {
    requests: number
    tokens: number
    cost: number
  }
}

export interface UsageStat {
  date: string
  requests: number
  tokens: number
  prompt_tokens: number
  completion_tokens: number
  cost: number
}

export interface UsageByModel {
  model: string
  requests: number
  tokens: number
  cost: number
}

export interface ModelUsageStats {
  model: string
  requests: number
  tokens: number
  cost: number
  percentage: number
}

export interface UserUsageStats {
  user_id: string
  username: string
  email: string
  requests: number
  tokens: number
  cost: number
  percentage: number
}

export interface UsageStatsResponse {
  stats: UsageStat[]
  by_model: UsageByModel[]
}

export interface TokenTrendPoint {
  label: string
  date: string
  hour: number
  requests: number
  tokens: number
  prompt_tokens: number
  completion_tokens: number
  reasoning_tokens: number
  cache_tokens: number
  cost: number
}

export interface RequestLog {
  id: string
  user_id: string
  api_key_id: string
  channel_id: string
  endpoint: string
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  reasoning_tokens: number
  cost: number
  cache_hit: boolean
  cache_tokens: number
  status_code: number
  is_stream: boolean
  first_byte_ms: number
  latency_ms: number
  error_message: string | null
  request_id: string
  ip_address: string
  created_at: string
  username?: string
  email?: string
  key_name?: string
  key_prefix?: string
  key_suffix?: string
  channel?: string
}

export interface LogListResponse {
  total: number
  page: number
  size: number
  items: RequestLog[]
}

export const usageApi = {
  overview: () =>
    alovaInstance.Get<UsageOverview>('/usage/overview'),
  
  stats: (params?: { days?: number; group_by?: string; start?: string; end?: string; model?: string }) =>
    alovaInstance.Get<UsageStatsResponse>('/usage/stats', { params }),
  
  logs: (params?: { page?: number; size?: number; model?: string; key?: string; start?: string; end?: string }) =>
    alovaInstance.Get<LogListResponse>('/usage/logs', { params }),

  adminLogs: (params?: { page?: number; size?: number; model?: string; key?: string; start?: string; end?: string }) =>
    alovaInstance.Get<LogListResponse>('/admin/usage/logs', { params }),

  adminOverview: () =>
    alovaInstance.Get<UsageOverview>('/admin/usage/overview'),

  adminDaily: (params?: { date?: string }) =>
    alovaInstance.Get<UsageOverview['today']>('/admin/usage/daily', { params }),

  adminTokenTrend: (params?: { granularity?: 'hour' | 'day'; date?: string; days?: number }) =>
    alovaInstance.Get<TokenTrendPoint[]>('/admin/usage/token-trend', { params }),

  adminTopModels: (params?: { limit?: number }) =>
    alovaInstance.Get<ModelUsageStats[]>('/admin/usage/top-models', { params }),

  adminTopUsers: (params?: { limit?: number }) =>
    alovaInstance.Get<UserUsageStats[]>('/admin/usage/top-users', { params }),
}
