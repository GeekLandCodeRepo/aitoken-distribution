import { alovaInstance } from '../lib/alova'

export interface RedeemCode {
  id: string
  code: string
  quota: number
  used_by: string | null
  used_at: string | null
  expires_at: string | null
  created_by: string
  created_at: string
}

export interface RedeemCodeListResponse {
  total: number
  page: number
  size: number
  items: RedeemCode[]
}

export interface GenerateCodesRequest {
  quota: number
  count: number
  expires_at?: string
}

export interface GenerateCodesResponse {
  count: number
  codes: string[]
}

export interface RedeemResponse {
  quota: number
  balance_before: number
  balance_after: number
}

export const redeemApi = {
  // 用户端
  redeem: (code: string) =>
    alovaInstance.Post<RedeemResponse>('/billing/redeem', { code }),
  
  // 管理端
  list: (params?: { page?: number; size?: number; status?: string }) =>
    alovaInstance.Get<RedeemCodeListResponse>('/admin/redeem-codes', { params }),
  
  generate: (data: GenerateCodesRequest) =>
    alovaInstance.Post<GenerateCodesResponse>('/admin/redeem-codes', data),
  
  delete: (id: string) =>
    alovaInstance.Delete(`/admin/redeem-codes/${id}`),
}
