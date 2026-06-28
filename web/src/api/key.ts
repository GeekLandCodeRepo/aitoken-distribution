import { alovaInstance } from '../lib/alova'

export interface ApiKey {
  id: string
  name: string
  key_prefix: string
  key_suffix: string
  status: number
  quota_limit: number
  used_quota: number
  rate_limit: number
  allowed_models: string[] | null
  allowed_ips: string[] | null
  expires_at: string | null
  last_used_at: string | null
  created_at: string
}

export interface CreateKeyRequest {
  name: string
  quota_limit?: number
  rate_limit?: number
  allowed_models?: string[]
  allowed_ips?: string[]
  expires_at?: string
}

export interface CreateKeyResponse extends ApiKey {
  key: string
}

export const apiKeyApi = {
  list: () =>
    alovaInstance.Get<ApiKey[]>('/api-keys'),
  
  create: (data: CreateKeyRequest) =>
    alovaInstance.Post<CreateKeyResponse>('/api-keys', data),
  
  update: (id: string, data: Partial<CreateKeyRequest>) =>
    alovaInstance.Put<ApiKey>(`/api-keys/${id}`, data),
  
  delete: (id: string) =>
    alovaInstance.Delete(`/api-keys/${id}`),
  
  toggle: (id: string) =>
    alovaInstance.Post<ApiKey>(`/api-keys/${id}/toggle`),
}
