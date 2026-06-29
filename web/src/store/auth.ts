import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { UserInfo } from '@/api/auth'

type AuthState = {
  accessToken: string | null
  refreshToken: string | null
  user: UserInfo | null
  setTokens: (accessToken: string, refreshToken: string) => void
  setUser: (user: UserInfo | null) => void
  clearTokens: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: localStorage.getItem('access_token'),
      refreshToken: localStorage.getItem('refresh_token'),
      user: null,
      setTokens: (accessToken, refreshToken) => {
        localStorage.setItem('access_token', accessToken)
        localStorage.setItem('refresh_token', refreshToken)
        set({ accessToken, refreshToken, user: null })
      },
      setUser: (user) => set({ user }),
      clearTokens: () => {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        set({ accessToken: null, refreshToken: null, user: null })
      },
    }),
    {
      name: 'auth',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    },
  ),
)

export const authStore = useAuthStore
