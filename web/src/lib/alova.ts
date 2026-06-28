import { createAlova } from 'alova'
import adapterFetch from 'alova/fetch'
import { baseURL } from './constants'
import { authStore } from '@/store/auth'

export const alovaInstance = createAlova({
  baseURL,
  requestAdapter: adapterFetch(),
  timeout: 30000,
  
  beforeRequest(method) {
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
        
        if (data.code === 100201 || data.code === 100202) {
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
