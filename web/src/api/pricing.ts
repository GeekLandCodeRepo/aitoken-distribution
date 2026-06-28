import { alovaInstance } from '../lib/alova'

export interface Pricing {
  id: string
  channel_id: string
  model_name: string
  prompt_price: number
  prompt_unit: number
  completion_price: number
  completion_unit: number
  image_price: number | null
  audio_price: number | null
  cached_prompt_price: number
  currency: string
  enabled: boolean
  created_at: string
}

export interface CreatePricingRequest {
  channel_id: string
  model_name: string
  prompt_price: number
  prompt_unit: number
  completion_price: number
  completion_unit: number
  image_price?: number
  audio_price?: number
  cached_prompt_price?: number
  currency?: string
  enabled?: boolean
}

export interface SyncResult {
  synced: number
  created: string[]
  skipped: string[]
}

export const pricingApi = {
  list: (params?: { channel_id?: string; enabled?: boolean; search?: string }) =>
    alovaInstance.Get<Pricing[]>('/admin/pricing', { params }),
  
  create: (data: CreatePricingRequest) =>
    alovaInstance.Post<Pricing>('/admin/pricing', data),
  
  update: (id: string, data: Partial<CreatePricingRequest>) =>
    alovaInstance.Put<Pricing>(`/admin/pricing/${id}`, data),

  toggle: (id: string, enabled: boolean) =>
    alovaInstance.Post<Pricing>(`/admin/pricing/${id}/toggle`, { enabled }),
  
  delete: (id: string) =>
    alovaInstance.Delete(`/admin/pricing/${id}`),
  
  sync: () =>
    alovaInstance.Post<SyncResult>('/admin/pricing/sync'),
}
