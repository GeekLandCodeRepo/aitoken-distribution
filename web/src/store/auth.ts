import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type AuthState = {
  accessToken: string | null
  refreshToken: string | null
  setTokens: (accessToken: string, refreshToken: string) => void
  clearTokens: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: localStorage.getItem('access_token'),
      refreshToken: localStorage.getItem('refresh_token'),
      setTokens: (accessToken, refreshToken) => {
        localStorage.setItem('access_token', accessToken)
        localStorage.setItem('refresh_token', refreshToken)
        set({ accessToken, refreshToken })
      },
      clearTokens: () => {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        set({ accessToken: null, refreshToken: null })
      },
    }),
    {
      name: 'auth',
    },
  ),
)

export const authStore = useAuthStore
