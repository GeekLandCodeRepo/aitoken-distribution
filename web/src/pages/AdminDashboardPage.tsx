import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { enUS, zhCN } from 'date-fns/locale'
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'

import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { usageApi, type ModelUsageStats, type UserUsageStats } from '@/api/usage'

interface ChartRow {
  label: string
  fullLabel: string
  percentage: number
}

function truncateLabel(value: string) {
  return value.length > 24 ? `${value.slice(0, 24)}...` : value
}

export function AdminDashboardPage() {
  const { t, i18n } = useTranslation()
  const [topModels, setTopModels] = useState<ModelUsageStats[]>([])
  const [topUsers, setTopUsers] = useState<UserUsageStats[]>([])
  const [selectedDate, setSelectedDate] = useState(new Date())
  const [dailyStats, setDailyStats] = useState({ requests: 0, tokens: 0, cost: 0 })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDashboard()
  }, [])

  useEffect(() => {
    fetchDailyStats(selectedDate)
  }, [selectedDate])

  const fetchDashboard = async () => {
    setLoading(true)
    try {
      const [modelsData, usersData] = await Promise.all([
        usageApi.adminTopModels({ limit: 10 }),
        usageApi.adminTopUsers({ limit: 10 }),
      ])
      setTopModels(modelsData || [])
      setTopUsers(usersData || [])
    } catch (err) {
      console.error('Failed to fetch admin dashboard:', err)
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (date: Date) => {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }

  const formatDisplayDate = (date: Date) => {
    return date.toLocaleDateString(i18n.language.startsWith('zh') ? 'zh-CN' : 'en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const formatTokenCount = (value: number) => {
    const abs = Math.abs(value)
    if (abs >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
    if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
    if (abs >= 1_000) return `${(value / 1_000).toFixed(2)}K`
    return value.toLocaleString()
  }

  const calendarLocale = i18n.language.startsWith('zh') ? zhCN : enUS

  const fetchDailyStats = async (date: Date) => {
    try {
      const stats = await usageApi.adminDaily({ date: formatDate(date) })
      setDailyStats(stats || { requests: 0, tokens: 0, cost: 0 })
    } catch (err) {
      console.error('Failed to fetch daily stats:', err)
    }
  }

  const chartConfig = {
    percentage: {
      label: t('adminDashboard.usageRate'),
      color: 'var(--chart-1)',
    },
  } satisfies ChartConfig

  const modelChartData: ChartRow[] = topModels.map((item) => ({
    label: truncateLabel(item.model),
    fullLabel: item.model,
    percentage: item.percentage,
  }))

  const userChartData: ChartRow[] = topUsers.map((item) => {
    const label = item.username || item.email || item.user_id
    return {
      label: truncateLabel(label),
      fullLabel: label,
      percentage: item.percentage,
    }
  })

  const renderChart = (data: ChartRow[]) => {
    if (loading) {
      return <div>{t('common.loading')}</div>
    }

    if (data.length === 0) {
      return <div className="text-sm text-muted-foreground">{t('adminDashboard.noData')}</div>
    }

    return (
      <ChartContainer config={chartConfig} className="h-[420px] w-full">
        <BarChart
          accessibilityLayer
          data={data}
          layout="vertical"
          margin={{ left: 8, right: 32 }}
        >
          <CartesianGrid horizontal={false} />
          <XAxis
            type="number"
            domain={[0, 100]}
            tickFormatter={(value) => `${value}%`}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            type="category"
            dataKey="label"
            width={170}
            tickLine={false}
            axisLine={false}
            tickMargin={8}
          />
          <ChartTooltip
            content={
              <ChartTooltipContent
                hideLabel={false}
                labelFormatter={(_, payload) => payload?.[0]?.payload?.fullLabel ?? ''}
                formatter={(value) => `${Number(value).toFixed(1)}%`}
              />
            }
          />
          <Bar dataKey="percentage" fill="var(--color-percentage)" radius={6} />
        </BarChart>
      </ChartContainer>
    )
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">{t('adminDashboard.title')}</h2>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t('adminDashboard.dailyStats')}</CardTitle>
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline">{formatDisplayDate(selectedDate)}</Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <Calendar
                mode="single"
                locale={calendarLocale}
                selected={selectedDate}
                onSelect={(date) => date && setSelectedDate(date)}
              />
            </PopoverContent>
          </Popover>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-lg border p-4">
              <div className="text-sm text-muted-foreground">{t('adminDashboard.dailyRequests')}</div>
              <div className="mt-2 text-3xl font-bold">{dailyStats.requests}</div>
            </div>
            <div className="rounded-lg border p-4">
              <div className="text-sm text-muted-foreground">{t('adminDashboard.dailyTokens')}</div>
              <div className="mt-2 text-3xl font-bold">{formatTokenCount(dailyStats.tokens)}</div>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t('adminDashboard.topModels')}</CardTitle>
          </CardHeader>
          <CardContent>
            {renderChart(modelChartData)}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('adminDashboard.topUsers')}</CardTitle>
          </CardHeader>
          <CardContent>
            {renderChart(userChartData)}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
