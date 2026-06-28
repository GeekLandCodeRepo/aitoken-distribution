import { alovaInstance } from '@/lib/alova'

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
  invite_code: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface UserInfo {
  id: string
  email: string
  username: string
  role: number
  balance: number
  used_quota: number
  request_count: number
  group_name: string
}

export const authApi = {
  login: (data: LoginRequest) =>
    alovaInstance.Post<TokenResponse>('/auth/login', data),
  
  register: (data: RegisterRequest) =>
    alovaInstance.Post<UserInfo>('/auth/register', data),
  
  refreshToken: (refreshToken: string) =>
    alovaInstance.Post<TokenResponse>('/auth/refresh', { refresh_token: refreshToken }),
  
  getMe: () =>
    alovaInstance.Get<UserInfo>('/auth/me', { params: { _: Date.now() } }),
  
  changePassword: (oldPassword: string, newPassword: string) =>
    alovaInstance.Put('/auth/password', { old_password: oldPassword, new_password: newPassword }),
}
