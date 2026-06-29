import { useEffect, useState } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { authApi } from '@/api'
import { useAuthStore } from '@/store/auth'

export function Layout() {
  const [loading, setLoading] = useState(true)
  const accessToken = useAuthStore((state) => state.accessToken)
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const clearTokens = useAuthStore((state) => state.clearTokens)
  const navigate = useNavigate()

  useEffect(() => {
    let active = true

    if (!accessToken) {
      setUser(null)
      setLoading(false)
      navigate('/login')
      return () => {
        active = false
      }
    }

    if (user) {
      setLoading(false)
      return () => {
        active = false
      }
    }

    setLoading(true)

    const fetchUserInfo = async () => {
      try {
        const data = await authApi.getMe()
        if (active) {
          setUser(data)
        }
      } catch {
        if (active) {
          clearTokens()
          navigate('/login')
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    fetchUserInfo()

    return () => {
      active = false
    }
  }, [accessToken, user, setUser, clearTokens, navigate])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    )
  }

  if (!user) {
    return null
  }

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar user={user} />
      
      <main className="flex-1 overflow-y-auto">
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
