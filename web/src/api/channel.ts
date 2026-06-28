import { alovaInstance } from '../lib/alova'

export interface Channel {
  id: string
  name: string
  type: number
  type_name: string
  base_url: string
  status: number
  priority: number
  weight: number
  balance: number
  models: string[]
  model_mapping: Record<string, string> | null
  groups: string[]
  used_quota: number
  request_count: number
  success_count: number
  success_rate: number
  config: Record<string, any> | null
  created_at: string
}

export interface ChannelListResponse {
  total: number
  page: number
  size: number
  items: Channel[]
}

export interface CreateChannelRequest {
  name: string
  type: number
  base_url: string
  api_key: string
  models: string[]
  model_mapping?: Record<string, string>
  priority?: number
  weight?: number
  groups?: string[]
  config?: Record<string, any>
}

export interface ChannelTestResult {
  success: boolean
  latency_ms: number
  tested_at: string
}

export const channelApi = {
  list: (params?: { page?: number; size?: number; status?: number; type?: number }) =>
    alovaInstance.Get<ChannelListResponse>('/admin/channels', { params }),
  
  create: (data: CreateChannelRequest) =>
    alovaInstance.Post<Channel>('/admin/channels', data),
  
  update: (id: string, data: Partial<CreateChannelRequest>) =>
    alovaInstance.Put<Channel>(`/admin/channels/${id}`, data),
  
  delete: (id: string) =>
    alovaInstance.Delete(`/admin/channels/${id}`),
  
  test: (id: string) =>
    alovaInstance.Post<ChannelTestResult>(`/admin/channels/${id}/test`),

  updateStatus: (id: string, enabled: boolean) =>
    alovaInstance.Put<Channel>(`/admin/channels/${id}/status`, { enabled }),
  
  batchCreate: (channels: CreateChannelRequest[]) =>
    alovaInstance.Post<{ created: number; channels: Channel[] }>('/admin/channels/batch', { channels }),
}
