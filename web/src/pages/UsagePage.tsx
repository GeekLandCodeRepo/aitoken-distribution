import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart'
import { usageApi, type UsageOverview, type RequestLog, type UsageStatsResponse } from '@/api/usage'
import { cn } from '@/lib/utils'

const trendChartConfig = {
  tokens: { label: 'Tokens', color: 'hsl(217 91% 60%)' },
  cost: { label: 'Cost', color: 'hsl(160 84% 39%)' },
} satisfies ChartConfig

const modelChartConfig = {
  tokens: { label: 'Tokens', color: 'hsl(262 83% 58%)' },
} satisfies ChartConfig

export function UsagePage() {
  const { t } = useTranslation()
  const [overview, setOverview] = useState<UsageOverview | null>(null)
  const [stats, setStats] = useState<UsageStatsResponse | null>(null)
  const [logs, setLogs] = useState<RequestLog[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [modelFilter, setModelFilter] = useState('')
  const isKeyFilter = modelFilter.trim().toLowerCase().startsWith('sk')

  useEffect(() => {
    fetchOverview()
    fetchStats()
    fetchLogs()
  }, [page])

  const dailyTrend = useMemo(() => buildDailyTrend(stats), [stats])
  const modelUsage = useMemo(() => buildModelUsage(stats), [stats])
  const successfulRequests = logs.filter((log) => log.status_code >= 200 && log.status_code < 300).length
  const successRate = logs.length > 0 ? Math.round((successfulRequests / logs.length) * 100) : 0
  const cacheHitCount = logs.filter((log) => log.cache_hit || log.cache_tokens > 0).length
  const cacheHitRate = logs.length > 0 ? Math.round((cacheHitCount / logs.length) * 100) : 0
  const averageLatency = logs.length > 0 ? Math.round(logs.reduce((sum, log) => sum + log.latency_ms, 0) / logs.length) : 0
  const monthBudgetUsed = overview && overview.balance + overview.this_month.cost > 0
    ? Math.min(100, Math.round((overview.this_month.cost / (overview.balance + overview.this_month.cost)) * 100))
    : 0

  const fetchOverview = async () => {
    try {
      const data = await usageApi.overview()
      setOverview(data)
    } catch (err) {
      console.error('Failed to fetch overview:', err)
    }
  }

  const fetchStats = async () => {
    try {
      const data = await usageApi.stats({ days: 14 })
      setStats(data)
    } catch (err) {
      console.error('Failed to fetch usage stats:', err)
    }
  }

  const fetchLogs = async (nextPage = page) => {
    setLoading(true)
    try {
      const value = modelFilter.trim()
      const data = await usageApi.logs({
        page: nextPage,
        size: 50,
        model: value && !isKeyFilter ? value : undefined,
        key: value && isKeyFilter ? value : undefined,
      })
      setLogs(data.items || [])
      setTotal(data.total || 0)
    } catch (err) {
      console.error('Failed to fetch logs:', err)
    } finally {
      setLoading(false)
    }
  }

  if (!overview) {
    return <div>{t('common.loading')}</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">{t('usage.title')}</h2>
          <p className="text-sm text-muted-foreground">{t('usage.subtitle')}</p>
        </div>
        <Badge variant="outline" className="w-fit">{t('usage.recentSample', { count: logs.length })}</Badge>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <MetricCard title={t('dashboard.balance')} value={formatMoney(overview.balance)} caption={t('usage.availableBalance')} accent="emerald" />
        <MetricCard title={t('usage.todayRequests')} value={formatNumber(overview.today.requests)} caption={`${formatNumber(overview.today.tokens)} tokens`} accent="blue" />
        <MetricCard title={t('usage.todayCost')} value={formatMoney(overview.today.cost, 4)} caption={t('usage.thisMonthCost', { cost: formatMoney(overview.this_month.cost, 4) })} accent="violet" />
        <MetricCard title={t('usage.cacheHitRate')} value={`${cacheHitRate}%`} caption={t('usage.cacheHitCaption', { count: cacheHitCount })} accent="cyan" />
        <MetricCard title={t('usage.totalRequests')} value={formatNumber(overview.request_count)} caption={t('usage.successRate', { rate: successRate })} accent="orange" />
      </div>

      <div className="grid gap-4 xl:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>{t('usage.trendTitle')}</CardTitle>
            <CardDescription>{t('usage.trendDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartContainer config={trendChartConfig} className="h-[280px] w-full">
              <AreaChart data={dailyTrend} margin={{ left: 8, right: 8, top: 8 }}>
                <defs>
                  <linearGradient id="tokensFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--color-tokens)" stopOpacity={0.35} />
                    <stop offset="95%" stopColor="var(--color-tokens)" stopOpacity={0.03} />
                  </linearGradient>
                </defs>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="date" tickLine={false} axisLine={false} tickMargin={8} />
                <YAxis tickLine={false} axisLine={false} tickMargin={8} width={42} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Area type="monotone" dataKey="tokens" stroke="var(--color-tokens)" fill="url(#tokensFill)" strokeWidth={2} />
              </AreaChart>
            </ChartContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('usage.healthTitle')}</CardTitle>
            <CardDescription>{t('usage.healthDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>{t('usage.monthBudget')}</span>
                <span className="font-medium">{monthBudgetUsed}%</span>
              </div>
              <Progress value={monthBudgetUsed} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <MiniStat label={t('usage.monthRequests')} value={formatNumber(overview.this_month.requests)} />
              <MiniStat label={t('usage.monthTokens')} value={formatNumber(overview.this_month.tokens)} />
              <MiniStat label={t('usage.avgLatency')} value={`${averageLatency}ms`} />
              <MiniStat label={t('usage.success')} value={`${successRate}%`} />
              <MiniStat label={t('usage.cacheHitRate')} value={`${cacheHitRate}%`} />
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 xl:grid-cols-[1fr_1.4fr]">
        <Card>
          <CardHeader>
            <CardTitle>{t('usage.modelBreakdown')}</CardTitle>
            <CardDescription>{t('usage.modelBreakdownDesc')}</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartContainer config={modelChartConfig} className="h-[280px] w-full">
              <BarChart data={modelUsage} layout="vertical" margin={{ left: 8, right: 16 }}>
                <CartesianGrid horizontal={false} />
                <XAxis type="number" hide />
                <YAxis dataKey="model" type="category" tickLine={false} axisLine={false} width={120} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="tokens" fill="var(--color-tokens)" radius={6} />
              </BarChart>
            </ChartContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <CardTitle>{t('usage.requestLogs')}</CardTitle>
              <CardDescription>{t('usage.requestLogsDesc')}</CardDescription>
            </div>
            <div className="flex gap-2">
              <Input
                placeholder={t('usage.filterByModelOrKey')}
                value={modelFilter}
                onChange={(e) => setModelFilter(e.target.value)}
                className="h-9 w-[230px]"
              />
              <Button size="sm" onClick={() => { setPage(1); fetchLogs(1) }}>{t('common.search')}</Button>
            </div>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="py-10 text-center text-sm text-muted-foreground">{t('common.loading')}</div>
            ) : (
              <div className="space-y-3">
                {logs.slice(0, 8).map((log) => (
                  <div key={log.id} className="flex flex-col gap-3 rounded-lg border p-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="min-w-0 space-y-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge variant="outline" className="font-mono">{log.model}</Badge>
                        <Badge className={cn(log.status_code >= 200 && log.status_code < 300 ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-100' : 'bg-red-100 text-red-700 hover:bg-red-100')}>{log.status_code}</Badge>
                        {log.is_stream && <Badge className="bg-violet-100 text-violet-700 hover:bg-violet-100">stream</Badge>}
                      </div>
                      <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
                        <span>{new Date(log.created_at).toLocaleString()}</span>
                        <span>·</span>
                        <span>{t('usage.apiKey')}:</span>
                        <span className="font-medium text-foreground">{log.key_name || t('usage.unlinkedKey')}</span>
                        <span className="font-mono">{formatKey(log)}</span>
                      </div>
                    </div>
                    <div className="grid grid-cols-3 gap-4 text-right text-sm">
                      <div>
                        <div className="font-medium">{formatNumber(log.total_tokens)}</div>
                        <div className="text-xs text-muted-foreground">tokens</div>
                      </div>
                      <div>
                        <div className="font-medium text-emerald-600">{formatMoney(log.cost, 5)}</div>
                        <div className="text-xs text-muted-foreground">{t('usage.cost')}</div>
                      </div>
                      <div>
                        <div className="font-medium">{log.latency_ms}ms</div>
                        <div className="text-xs text-muted-foreground">{t('usage.latency')}</div>
                      </div>
                    </div>
                  </div>
                ))}
                <div className="flex items-center justify-between pt-2 text-sm text-muted-foreground">
                  <span>{t('common.total')}: {total}</span>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>{t('common.previous')}</Button>
                    <Button variant="outline" size="sm" disabled={page * 50 >= total} onClick={() => setPage(page + 1)}>{t('common.next')}</Button>
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function MetricCard({ title, value, caption, accent }: { title: string; value: string; caption: string; accent: 'emerald' | 'blue' | 'violet' | 'orange' | 'cyan' }) {
  const accents = {
    emerald: 'from-emerald-500/15 to-transparent text-emerald-600',
    blue: 'from-blue-500/15 to-transparent text-blue-600',
    violet: 'from-violet-500/15 to-transparent text-violet-600',
    cyan: 'from-cyan-500/15 to-transparent text-cyan-600',
    orange: 'from-orange-500/15 to-transparent text-orange-600',
  }
  return (
    <Card className="overflow-hidden py-0">
      <CardContent className={cn('bg-gradient-to-br p-4', accents[accent])}>
        <div className="text-sm font-medium text-muted-foreground">{title}</div>
        <div className="mt-2 text-2xl font-bold tracking-tight text-foreground">{value}</div>
        <div className="mt-1 text-xs text-muted-foreground">{caption}</div>
      </CardContent>
    </Card>
  )
}

function MiniStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-muted/25 p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-lg font-semibold">{value}</div>
    </div>
  )
}

function buildDailyTrend(stats: UsageStatsResponse | null) {
  const days = Array.from({ length: 14 }, (_, index) => {
    const date = new Date()
    date.setDate(date.getDate() - (13 - index))
    const key = date.toISOString().slice(0, 10)
    return { key, date: key.slice(5), tokens: 0, cost: 0 }
  })
  const byDate = new Map(days.map((item) => [item.key, item]))
  for (const stat of stats?.stats || []) {
    const item = byDate.get(stat.date)
    if (item) {
      item.tokens = stat.tokens
      item.cost = stat.cost / 1000000
    }
  }
  return days
}

function buildModelUsage(stats: UsageStatsResponse | null) {
  return (stats?.by_model || [])
    .map((item) => ({ model: item.model, tokens: item.tokens }))
    .sort((a, b) => b.tokens - a.tokens)
    .slice(0, 8)
}

function formatMoney(value: number, digits = 2) {
  return `$${(value / 1000000).toFixed(digits)}`
}

function formatNumber(value: number) {
  return new Intl.NumberFormat().format(value || 0)
}

function formatKey(log: RequestLog) {
  if (!log.key_prefix) return '-'
  return log.key_suffix ? `${log.key_prefix}......${log.key_suffix}` : `${log.key_prefix}......`
}
