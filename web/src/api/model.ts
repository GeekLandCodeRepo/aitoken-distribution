import { alovaInstance } from '../lib/alova'

export interface PlazaModel {
  model_name: string
  prompt_price: number
  prompt_unit: number
  completion_price: number
  completion_unit: number
  cached_prompt_price: number
  currency: string
}

export interface PlazaChannel {
  channel_id: string
  channel_name: string
  channel_type: number
  models: PlazaModel[]
}

export interface ManagedModel {
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

export interface CreateManagedModelRequest {
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

export const modelApi = {
  plaza: () => alovaInstance.Get<PlazaChannel[]>('/models/plaza'),
}

export const adminModelApi = {
  list: (params?: { channel_id?: string; enabled?: boolean; search?: string }) =>
    alovaInstance.Get<ManagedModel[]>('/admin/models', { params }),

  create: (data: CreateManagedModelRequest) =>
    alovaInstance.Post<ManagedModel>('/admin/models', data),

  update: (id: string, data: Partial<CreateManagedModelRequest>) =>
    alovaInstance.Put<ManagedModel>(`/admin/models/${id}`, data),

  toggle: (id: string, enabled: boolean) =>
    alovaInstance.Post<ManagedModel>(`/admin/models/${id}/toggle`, { enabled }),

  delete: (id: string) =>
    alovaInstance.Delete(`/admin/models/${id}`),
}
