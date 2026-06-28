import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Search, SlidersHorizontal, Zap } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { usageApi, type RequestLog } from '@/api/usage'
import { cn } from '@/lib/utils'

const timeRanges = ['today', '1h', '6h', '24h', '7d', '30d', 'custom'] as const
const pageSizeOptions = [20, 50, 100]

export function LogsPage() {
  const { t } = useTranslation()
  const [logs, setLogs] = useState<RequestLog[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [search, setSearch] = useState('')
  const [modelFilter, setModelFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [typeFilter, setTypeFilter] = useState('all')
  const [timeRange, setTimeRange] = useState<(typeof timeRanges)[number]>('today')

  const isAdminLogs = window.location.pathname.startsWith('/admin/')
  const isKeySearch = search.trim().toLowerCase().startsWith('sk')

  useEffect(() => {
    fetchLogs()
  }, [page, pageSize])

  const visibleLogs = useMemo(() => {
    return logs.filter((log) => {
      const keyword = search.trim().toLowerCase()
      const matchesSearch = isKeySearch || !keyword || [log.email, log.username, log.ip_address, log.key_name, log.channel]
        .filter(Boolean)
        .some((value) => value!.toLowerCase().includes(keyword))
      const matchesStatus = statusFilter === 'all' || String(log.status_code) === statusFilter
      const matchesType = typeFilter === 'all' || (typeFilter === 'stream' ? log.is_stream : !log.is_stream)
      return matchesSearch && matchesStatus && matchesType
    })
  }, [logs, search, isKeySearch, statusFilter, typeFilter])

  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const pageNumbers = getPageNumbers(page, totalPages)

  const fetchLogs = async (nextPage = page) => {
    setLoading(true)
    try {
      const params = {
        page: nextPage,
        size: pageSize,
        model: modelFilter || undefined,
        key: isKeySearch ? search.trim() : undefined,
      }
      const data = isAdminLogs ? await usageApi.adminLogs(params) : await usageApi.logs(params)
      setLogs(data.items || [])
      setTotal(data.total || 0)
    } catch (err) {
      console.error('Failed to fetch logs:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = () => {
    setPage(1)
    fetchLogs(1)
  }

  return (
    <div className="space-y-4">
      <Card className="gap-3 py-4">
        <CardContent className="space-y-3 px-4">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-lg font-semibold">{t('logs.title')}</h2>
              <div className="flex rounded-lg border bg-muted/30 p-0.5">
                {timeRanges.map((range) => (
                  <Button
                    key={range}
                    variant={timeRange === range ? 'outline' : 'ghost'}
                    size="sm"
                    className="h-7 rounded-md px-2.5 text-xs shadow-none"
                    onClick={() => setTimeRange(range)}
                  >
                    {t(`logs.timeRanges.${range}`)}
                  </Button>
                ))}
              </div>
            </div>
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              <span>{total} {t('logs.records')}</span>
              <Button variant="destructive" size="sm" className="h-8" disabled>
                {t('logs.clearLogs')}
              </Button>
            </div>
          </div>

          <div className="flex flex-col gap-2 rounded-lg border bg-background p-2 lg:flex-row lg:items-center">
            <div className="relative min-w-[260px] flex-1">
              <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder={t('logs.searchPlaceholder')}
                className="h-9 pl-9"
              />
            </div>
            <Input
              value={modelFilter}
              onChange={(event) => setModelFilter(event.target.value)}
              placeholder={t('logs.modelFilter')}
              className="h-9 lg:w-[150px]"
            />
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="h-9 lg:w-[150px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('logs.allStatus')}</SelectItem>
                <SelectItem value="200">200</SelectItem>
                <SelectItem value="400">400</SelectItem>
                <SelectItem value="401">401</SelectItem>
                <SelectItem value="500">500</SelectItem>
              </SelectContent>
            </Select>
            <Select value={typeFilter} onValueChange={setTypeFilter}>
              <SelectTrigger className="h-9 lg:w-[150px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('logs.allTypes')}</SelectItem>
                <SelectItem value="stream">stream</SelectItem>
                <SelectItem value="standard">standard</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm" className="h-9 gap-1" onClick={handleSearch}>
              <Zap className="size-3.5" /> Fast
            </Button>
            <Button variant="outline" size="sm" className="ml-auto h-9 gap-1">
              <SlidersHorizontal className="size-3.5" /> {t('logs.columns')}
            </Button>
          </div>

          <div className="overflow-hidden rounded-lg border">
            <div className="max-h-[620px] overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/35 hover:bg-muted/35">
                    <TableHead className="sticky top-0 z-20 min-w-[70px] bg-muted">{t('logs.status')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[150px] bg-muted">{t('logs.model')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[190px] bg-muted">{t('logs.sourceAccount')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[170px] bg-muted">{t('logs.apiKey')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[120px] bg-muted">IP</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[170px] bg-muted">{t('logs.endpoint')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[130px] bg-muted">{t('logs.type')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[210px] bg-muted text-right">TOKEN</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[120px] bg-muted text-right">{t('logs.cost')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[110px] bg-muted text-right">{t('logs.cacheRead')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[95px] bg-muted text-right">{t('logs.firstByte')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[85px] bg-muted text-right">{t('logs.totalTime')}</TableHead>
                    <TableHead className="sticky top-0 z-20 min-w-[120px] bg-muted text-right">{t('logs.time')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loading ? (
                    <TableRow>
                      <TableCell colSpan={13} className="h-24 text-center text-muted-foreground">
                        {t('common.loading')}
                      </TableCell>
                    </TableRow>
                  ) : visibleLogs.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={13} className="h-24 text-center text-muted-foreground">
                        {t('common.noData')}
                      </TableCell>
                    </TableRow>
                  ) : (
                    visibleLogs.map((log) => <LogRow key={log.id} log={log} />)
                  )}
                </TableBody>
              </Table>
            </div>
          </div>

          <div className="flex flex-col gap-3 text-sm text-muted-foreground lg:flex-row lg:items-center lg:justify-between">
            <span>{t('logs.showing', { start: total === 0 ? 0 : (page - 1) * pageSize + 1, end: Math.min(page * pageSize, total), total })}</span>
            <div className="flex flex-wrap items-center gap-2 lg:justify-end">
              <span>{t('logs.pageSize')}</span>
              <Select value={String(pageSize)} onValueChange={(value) => { setPageSize(Number(value)); setPage(1) }}>
                <SelectTrigger className="h-8 w-[100px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {pageSizeOptions.map((size) => (
                    <SelectItem key={size} value={String(size)}>{size} {t('logs.perPage')}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                {t('common.previous')}
              </Button>
              {pageNumbers.map((item, index) => item === 'ellipsis' ? (
                <span key={`${item}-${index}`} className="px-1">...</span>
              ) : (
                <Button
                  key={item}
                  variant={page === item ? 'default' : 'outline'}
                  size="sm"
                  className="h-8 w-8 px-0"
                  onClick={() => setPage(item)}
                >
                  {item}
                </Button>
              ))}
              <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
                {t('common.next')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function LogRow({ log }: { log: RequestLog }) {
  const firstByte = log.first_byte_ms || 0
  const totalTime = log.latency_ms || 0

  return (
    <TableRow className="text-xs">
      <TableCell>
        <Badge className={cn('font-mono', log.status_code >= 200 && log.status_code < 300 ? 'bg-emerald-100 text-emerald-700 hover:bg-emerald-100' : 'bg-red-100 text-red-700 hover:bg-red-100')}>
          {log.status_code}
        </Badge>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="font-mono text-xs">{log.model}</Badge>
          <PriorityBadge tokens={log.total_tokens} />
        </div>
      </TableCell>
      <TableCell className="text-muted-foreground">{log.email || log.username || '-'}</TableCell>
      <TableCell>
        <div className="space-y-0.5">
          <div className="font-medium text-foreground">{log.key_name || '-'}</div>
          <div className="font-mono text-[11px] text-muted-foreground">{formatKey(log)}</div>
        </div>
      </TableCell>
      <TableCell className="font-mono text-muted-foreground">{log.ip_address || '-'}</TableCell>
      <TableCell className="font-mono text-muted-foreground">{log.endpoint || '/v1/chat/completions'}</TableCell>
      <TableCell>
        <div className="flex items-center gap-1.5">
          <Badge className="bg-violet-100 text-violet-700 hover:bg-violet-100">{log.is_stream ? 'stream' : 'standard'}</Badge>
          {log.cache_hit && <Badge className="bg-cyan-100 text-cyan-700 hover:bg-cyan-100">Compact</Badge>}
        </div>
      </TableCell>
      <TableCell className="text-right font-mono">
        <span className="text-blue-600">↓{formatNumber(log.prompt_tokens)}</span>
        <span className="mx-1 text-muted-foreground">|</span>
        <span className="text-emerald-600">↑{formatNumber(log.completion_tokens)}</span>
        {log.cache_tokens > 0 && <span className="ml-2 text-orange-500">♺{formatNumber(log.cache_tokens)}</span>}
      </TableCell>
      <TableCell className="text-right font-mono font-semibold text-emerald-600">${(log.cost / 1000000).toFixed(6)}</TableCell>
      <TableCell className="text-right">
        <Badge className="bg-indigo-100 font-mono text-indigo-700 hover:bg-indigo-100">{formatNumber(log.cache_tokens)}</Badge>
      </TableCell>
      <TableCell className={cn('text-right font-mono', durationClass(firstByte))}>{firstByte > 0 ? `${formatSeconds(firstByte)}s` : '-'}</TableCell>
      <TableCell className={cn('text-right font-mono', durationClass(totalTime))}>{formatSeconds(totalTime)}s</TableCell>
      <TableCell className="text-right font-mono text-[11px] text-muted-foreground">
        <div>{formatDate(log.created_at)}</div>
        <div>{formatClock(log.created_at)}</div>
      </TableCell>
    </TableRow>
  )
}

function PriorityBadge({ tokens }: { tokens: number }) {
  if (tokens > 100000) {
    return <Badge className="bg-rose-100 text-rose-600 hover:bg-rose-100">xhigh</Badge>
  }
  if (tokens > 50000) {
    return <Badge className="bg-red-100 text-red-600 hover:bg-red-100">high</Badge>
  }
  if (tokens > 10000) {
    return <Badge className="bg-orange-100 text-orange-600 hover:bg-orange-100">medium</Badge>
  }
  return <Badge className="bg-slate-100 text-slate-600 hover:bg-slate-100">low</Badge>
}

function getPageNumbers(page: number, totalPages: number) {
  const pages: Array<number | 'ellipsis'> = []
  const maxVisible = Math.min(totalPages, 5)
  for (let i = 1; i <= maxVisible; i++) {
    pages.push(i)
  }
  if (totalPages > 6) {
    pages.push('ellipsis')
    pages.push(totalPages)
  }
  if (page > 5 && page < totalPages) {
    return [1, 'ellipsis' as const, page, totalPages]
  }
  return pages
}

function formatNumber(value: number) {
  return new Intl.NumberFormat().format(value || 0)
}

function formatSeconds(ms: number) {
  return (ms / 1000).toFixed(1)
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('sv-SE')
}

function formatClock(value: string) {
  return new Date(value).toLocaleTimeString('sv-SE', { hour12: false })
}

function formatKey(log: RequestLog) {
  if (!log.key_prefix) return '-'
  return log.key_suffix ? `${log.key_prefix}......${log.key_suffix}` : `${log.key_prefix}......`
}

function durationClass(ms: number) {
  if (ms >= 10000) return 'text-red-500'
  if (ms >= 3000) return 'text-orange-500'
  if (ms > 0 && ms <= 2000) return 'text-emerald-600'
  return 'text-muted-foreground'
}
