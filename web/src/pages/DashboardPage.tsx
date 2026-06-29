import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from '@/components/ui/chart'
import { usageApi, type APIKeyUsageStats, type TokenTrendPoint } from '@/api/usage'
import { useAuthStore } from '@/store/auth'

const formatDate = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatTokenCount = (value: number) => {
  const abs = Math.abs(value)
  if (abs >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (abs >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

export function DashboardPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const [trendGranularity, setTrendGranularity] = useState<'hour' | 'day'>('hour')
  const [tokenTrend, setTokenTrend] = useState<TokenTrendPoint[]>([])
  const [topAPIKeys, setTopAPIKeys] = useState<APIKeyUsageStats[]>([])
  const [trendLoading, setTrendLoading] = useState(false)
  const [keysLoading, setKeysLoading] = useState(false)

  useEffect(() => {
    const fetchDashboard = async () => {
      try {
        const keys = await usageApi.topAPIKeys({ limit: 10 })
        setTopAPIKeys(keys || [])
      } catch {
        console.error('Failed to fetch dashboard')
      }
    }
    fetchDashboard()
  }, [])

  useEffect(() => {
    fetchTokenTrend()
  }, [trendGranularity])

  const fetchTokenTrend = async () => {
    setTrendLoading(true)
    try {
      const data = await usageApi.tokenTrend(
        trendGranularity === 'hour'
          ? { granularity: 'hour', date: formatDate(new Date()) }
          : { granularity: 'day', days: 14 },
      )
      setTokenTrend(data || [])
    } catch (err) {
      console.error('Failed to fetch token trend:', err)
      setTokenTrend([])
    } finally {
      setTrendLoading(false)
    }
  }

  const refreshTopAPIKeys = async () => {
    setKeysLoading(true)
    try {
      const keys = await usageApi.topAPIKeys({ limit: 10 })
      setTopAPIKeys(keys || [])
    } catch (err) {
      console.error('Failed to fetch top api keys:', err)
      setTopAPIKeys([])
    } finally {
      setKeysLoading(false)
    }
  }

  if (!user) {
    return <div>{t('common.loading')}</div>
  }

  const balanceUSD = (user.balance / 1000000).toFixed(2)
  const usedQuotaUSD = (user.used_quota / 1000000).toFixed(2)

  const tokenTrendConfig = {
    tokens: { label: t('dashboard.totalTokens'), color: 'var(--chart-1)' },
    prompt_tokens: { label: t('dashboard.promptTokens'), color: 'var(--chart-2)' },
    completion_tokens: { label: t('dashboard.completionTokens'), color: 'var(--chart-3)' },
    reasoning_tokens: { label: t('dashboard.reasoningTokens'), color: 'var(--chart-4)' },
    cache_tokens: { label: t('dashboard.cacheTokens'), color: 'var(--chart-5)' },
  } satisfies ChartConfig

  const renderTokenTrendChart = () => {
    if (trendLoading) return <div>{t('common.loading')}</div>
    if (tokenTrend.length === 0) return <div className="text-sm text-muted-foreground">{t('dashboard.noTrendData')}</div>

    return (
      <ChartContainer config={tokenTrendConfig} className="h-[340px] w-full">
        <AreaChart accessibilityLayer data={tokenTrend} margin={{ left: 8, right: 24 }}>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="label" tickLine={false} axisLine={false} tickMargin={8} minTickGap={24} />
          <YAxis tickLine={false} axisLine={false} tickMargin={8} tickFormatter={(value) => formatTokenCount(Number(value))} />
          <ChartTooltip
            content={
              <ChartTooltipContent
                formatter={(value, name) => {
                  const key = String(name) as keyof typeof tokenTrendConfig
                  return `${tokenTrendConfig[key]?.label || name}: ${formatTokenCount(Number(value))}`
                }}
              />
            }
          />
          <Area type="monotone" dataKey="tokens" stroke="var(--color-tokens)" fill="var(--color-tokens)" fillOpacity={0.18} strokeWidth={2} />
          <Area type="monotone" dataKey="prompt_tokens" stroke="var(--color-prompt_tokens)" fill="var(--color-prompt_tokens)" fillOpacity={0.08} strokeWidth={1.5} />
          <Area type="monotone" dataKey="completion_tokens" stroke="var(--color-completion_tokens)" fill="var(--color-completion_tokens)" fillOpacity={0.08} strokeWidth={1.5} />
          <Area type="monotone" dataKey="reasoning_tokens" stroke="var(--color-reasoning_tokens)" fill="var(--color-reasoning_tokens)" fillOpacity={0.05} strokeWidth={1.2} />
          <Area type="monotone" dataKey="cache_tokens" stroke="var(--color-cache_tokens)" fill="var(--color-cache_tokens)" fillOpacity={0.05} strokeWidth={1.2} />
        </AreaChart>
      </ChartContainer>
    )
  }

  const renderKeyLabel = (key: APIKeyUsageStats) => {
    const name = key.key_name || t('dashboard.unknownApiKey')
    if (!key.key_prefix && !key.key_suffix) return name
    return `${name} (${key.key_prefix}...${key.key_suffix})`
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">{t('dashboard.title')}</h2>
        <p className="text-sm text-muted-foreground">{t('dashboard.subtitle')}</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card className="overflow-hidden border-primary/15 bg-gradient-to-br from-primary/10 via-card to-card">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('dashboard.balance')}</CardTitle>
            <span className="text-xl">💰</span>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold tracking-tight">${balanceUSD}</div>
            <p className="mt-1 text-xs text-muted-foreground">{t('dashboard.availableBalance')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('dashboard.usedQuota')}</CardTitle>
            <span className="text-xl">📊</span>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold tracking-tight">${usedQuotaUSD}</div>
            <p className="mt-1 text-xs text-muted-foreground">{t('dashboard.totalSpent')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('dashboard.requests')}</CardTitle>
            <span className="text-xl">🔄</span>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold tracking-tight">{user.request_count.toLocaleString()}</div>
            <p className="mt-1 text-xs text-muted-foreground">{t('dashboard.totalRequests')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">{t('dashboard.group')}</CardTitle>
            <span className="text-xl">👥</span>
          </CardHeader>
          <CardContent>
            <div className="truncate text-3xl font-bold capitalize tracking-tight">{user.group_name}</div>
            <p className="mt-1 text-xs text-muted-foreground">{t('dashboard.accountGroup')}</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.7fr)_minmax(340px,0.9fr)]">
        <Card className="min-w-0">
          <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <CardTitle>{t('dashboard.tokenTrend')}</CardTitle>
              <p className="mt-1 text-sm text-muted-foreground">{t('dashboard.tokenTrendDesc')}</p>
            </div>
            <div className="flex w-full gap-2 sm:w-auto">
              <Button
                size="sm"
                className="flex-1 sm:flex-none"
                variant={trendGranularity === 'hour' ? 'default' : 'outline'}
                onClick={() => setTrendGranularity('hour')}
              >
                {t('dashboard.byHour')}
              </Button>
              <Button
                size="sm"
                className="flex-1 sm:flex-none"
                variant={trendGranularity === 'day' ? 'default' : 'outline'}
                onClick={() => setTrendGranularity('day')}
              >
                {t('dashboard.byDay')}
              </Button>
            </div>
          </CardHeader>
          <CardContent>{renderTokenTrendChart()}</CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between gap-3">
              <div>
                <CardTitle>{t('dashboard.topApiKeys')}</CardTitle>
                <p className="mt-1 text-sm text-muted-foreground">{t('dashboard.topApiKeysDesc')}</p>
              </div>
              <Button size="sm" variant="outline" onClick={refreshTopAPIKeys}>{t('common.refresh')}</Button>
            </CardHeader>
            <CardContent>
              {keysLoading ? (
                <div>{t('common.loading')}</div>
              ) : topAPIKeys.length === 0 ? (
                <div className="text-sm text-muted-foreground">{t('common.noData')}</div>
              ) : (
                <div className="space-y-3">
                  {topAPIKeys.map((key, index) => (
                    <div key={`${key.key_id || 'unknown'}-${index}`} className="rounded-lg border bg-muted/20 p-3">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-medium">#{index + 1} {renderKeyLabel(key)}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {t('dashboard.requests')}: {key.requests.toLocaleString()} · {key.percentage.toFixed(1)}%
                          </div>
                        </div>
                        <div className="shrink-0 text-right font-mono font-semibold">{formatTokenCount(key.tokens)}</div>
                      </div>
                      <div className="mt-3 h-2 overflow-hidden rounded-full bg-muted">
                        <div className="h-full rounded-full bg-primary" style={{ width: `${Math.min(100, Math.max(0, key.percentage))}%` }} />
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t('dashboard.quickStart')}</CardTitle>
              <p className="text-sm text-muted-foreground">{t('dashboard.quickStartDesc')}</p>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium">{t('dashboard.apiBaseUrl')}</label>
                  <div className="mt-1 overflow-x-auto rounded-md bg-muted p-3 font-mono text-sm">
                    {`http://${window.location.hostname}:40680/v1`}
                  </div>
                </div>
                <div>
                  <label className="text-sm font-medium">{t('dashboard.apiKey')}</label>
                  <div className="mt-1 rounded-md bg-muted p-3 text-sm text-muted-foreground">
                    {t('dashboard.apiKeyHint')}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
