import { alovaInstance } from '@/lib/alova'

export interface User {
  id: string
  username: string
  email: string
  role: number
  balance: number
  used_quota: number
  request_count: number
  status: number
  group_name: string
  created_at: string
}

export interface UserListResponse {
  total: number
  page: number
  size: number
  items: User[]
}

export interface TopUpRequest {
  amount: number
  description?: string
}

export interface TopUpResponse {
  user_id: string
  balance_before: number
  balance_after: number
  amount: number
}

export interface CreateUserRequest {
  email: string
  username: string
  password: string
  role: number
  balance: number
}

export interface InviteSettings {
  require_invite_register: boolean
  user_invite_enabled: boolean
  reward_amount: number
  new_user_bonus_amount: number
}

export interface InviteCode {
  id: string
  code: string
  inviter_user_id: string
  used_by_user_id: string
  used_at: string | null
  reward_amount: number
  new_user_bonus: number
  created_at: string
}

export interface InviteCodeListResponse {
  total: number
  page: number
  size: number
  items: InviteCode[]
}

export interface UserChannelOption {
  id: string
  name: string
  type: number
  status: number
  base_url: string
  models: string[]
  allowed: boolean
}

export const userApi = {
  list: (params?: { page?: number; size?: number; search?: string; status?: number; role?: number }) =>
    alovaInstance.Get<UserListResponse>('/admin/users', { params }),

  create: (data: CreateUserRequest) =>
    alovaInstance.Post<User>('/admin/users', data),
  
  getById: (id: string) =>
    alovaInstance.Get<User>(`/admin/users/${id}`),
  
  update: (id: string, data: Partial<User>) =>
    alovaInstance.Put<User>(`/admin/users/${id}`, data),
  
  delete: (id: string) =>
    alovaInstance.Delete(`/admin/users/${id}`),
  
  topUp: (id: string, data: TopUpRequest) =>
    alovaInstance.Post<TopUpResponse>(`/admin/users/${id}/topup`, data),

  resetPassword: (id: string) =>
    alovaInstance.Post<{ default_password: string }>(`/admin/users/${id}/reset-password`),

  inviteSettings: () =>
    alovaInstance.Get<InviteSettings>('/admin/invites/settings'),

  updateInviteSettings: (data: InviteSettings) =>
    alovaInstance.Put<InviteSettings>('/admin/invites/settings', data),

  createInviteCode: (data: { inviter_user_id?: string; reward_amount?: number; new_user_bonus?: number }) =>
    alovaInstance.Post<InviteCode>('/admin/invites', data),

  listInviteCodes: (params?: { page?: number; size?: number }) =>
    alovaInstance.Get<InviteCodeListResponse>('/admin/invites', { params }),

  listChannels: (userId: string) =>
    alovaInstance.Get<UserChannelOption[]>(`/admin/users/${userId}/channels`),

  updateChannels: (userId: string, channel_ids: string[]) =>
    alovaInstance.Put<UserChannelOption[]>(`/admin/users/${userId}/channels`, { channel_ids }),

  getChannelTemplate: () =>
    alovaInstance.Get<UserChannelOption[]>('/admin/users/channel-template'),

  updateChannelTemplate: (channel_ids: string[]) =>
    alovaInstance.Put<UserChannelOption[]>('/admin/users/channel-template', { channel_ids }),

  applyChannelTemplateToAllUsers: () =>
    alovaInstance.Post<{ affected?: number; message?: string }>('/admin/users/channel-template/apply-all'),
}
