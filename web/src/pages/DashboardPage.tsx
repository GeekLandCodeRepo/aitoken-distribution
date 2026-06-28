import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { authApi, type UserInfo } from '@/api'

export function DashboardPage() {
  const { t } = useTranslation()
  const [user, setUser] = useState<UserInfo | null>(null)

  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const data = await authApi.getMe()
        setUser(data)
      } catch {
        console.error('Failed to fetch user info')
      }
    }
    fetchUserInfo()
  }, [])

  if (!user) {
    return <div>{t('common.loading')}</div>
  }

  const balanceUSD = (user.balance / 1000000).toFixed(2)
  const usedQuotaUSD = (user.used_quota / 1000000).toFixed(2)

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">{t('dashboard.title')}</h2>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('dashboard.balance')}</CardTitle>
            <span className="text-2xl">💰</span>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${balanceUSD}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('dashboard.usedQuota')}</CardTitle>
            <span className="text-2xl">📊</span>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${usedQuotaUSD}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('dashboard.requests')}</CardTitle>
            <span className="text-2xl">🔄</span>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{user.request_count}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('dashboard.group')}</CardTitle>
            <span className="text-2xl">👥</span>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold capitalize">{user.group_name}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('dashboard.quickStart')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">{t('dashboard.apiBaseUrl')}</label>
              <div className="mt-1 p-3 bg-muted rounded-md font-mono text-sm">
                {`http://${window.location.hostname}:40680/v1`}
              </div>
            </div>
            <div>
              <label className="text-sm font-medium">{t('dashboard.apiKey')}</label>
              <div className="mt-1 p-3 bg-muted rounded-md font-mono text-sm">
                {t('dashboard.apiKeyHint')}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
