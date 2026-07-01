import { createAlova } from 'alova'
import adapterFetch from 'alova/fetch'
import { baseURL } from './constants'
import { authStore } from '@/store/auth'

type TokenResponse = {
  access_token: string
  refresh_token: string
  expires_in: number
}

let refreshPromise: Promise<string | null> | null = null

function getTokenExp(token: string | null) {
  if (!token) return 0
  try {
    const payload = token.split('.')[1]
    if (!payload) return 0
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
    const json = JSON.parse(window.atob(padded))
    return typeof json.exp === 'number' ? json.exp * 1000 : 0
  } catch {
    return 0
  }
}

function isAuthRequest(url: string) {
  return url.includes('/auth/login') || url.includes('/auth/register') || url.includes('/auth/refresh')
}

async function refreshAccessToken() {
  const refreshToken = authStore.getState().refreshToken
  if (!refreshToken) return null

  const response = await fetch(`${baseURL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }

  const data = await response.json()
  if (data.code !== 0) {
    throw new Error(data.message || 'refresh token failed')
  }

  const tokens = data.data as TokenResponse
  authStore.getState().setTokens(tokens.access_token, tokens.refresh_token)
  return tokens.access_token
}

async function ensureFreshAccessToken() {
  const { accessToken, refreshToken } = authStore.getState()
  if (!accessToken || !refreshToken) return

  // Refresh one minute before expiry to avoid a request racing against expiration.
  const refreshBefore = Date.now() + 60 * 1000
  if (getTokenExp(accessToken) > refreshBefore) return

  if (!refreshPromise) {
    refreshPromise = refreshAccessToken()
      .catch(() => {
        authStore.getState().clearTokens()
        window.location.href = '/login'
        return null
      })
      .finally(() => {
        refreshPromise = null
      })
  }
  await refreshPromise
}

export const alovaInstance = createAlova({
  baseURL,
  requestAdapter: adapterFetch(),
  timeout: 30000,
  
  async beforeRequest(method) {
    const url = String((method as any).url || '')
    if (!isAuthRequest(url)) {
      await ensureFreshAccessToken()
    }

    const token = authStore.getState().accessToken
    if (token) {
      method.config.headers = {
        ...method.config.headers,
        Authorization: `Bearer ${token}`,
      }
    }
  },
  
  responded: {
    onSuccess: async (response) => {
      // 检查响应状态
      if (!response.ok) {
        // 尝试解析错误响应
        try {
          const errorData = await response.json()
          throw new Error(errorData.message || `HTTP ${response.status}`)
        } catch (e) {
          if (e instanceof Error && e.message.startsWith('HTTP')) {
            throw e
          }
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
      }

      // 检查 Content-Type
      const contentType = response.headers.get('content-type')
      if (!contentType || !contentType.includes('application/json')) {
        // 如果不是 JSON，返回空对象
        return {}
      }

      // 解析 JSON
      const text = await response.text()
      if (!text) {
        return {}
      }

      try {
        const data = JSON.parse(text)
        
        if (data.code === 0) {
          return data.data
        }
        
        if (data.code === 100201 || data.code === 100202 || data.code === 100203) {
          authStore.getState().clearTokens()
          window.location.href = '/login'
        }
        
        throw new Error(data.message || 'Request failed')
      } catch (e) {
        if (e instanceof Error && (e.message === 'Request failed' || e.message.includes('token'))) {
          throw e
        }
        throw new Error('Invalid JSON response')
      }
    },
    
    onError: (error) => {
      throw error
    },
  },
})
