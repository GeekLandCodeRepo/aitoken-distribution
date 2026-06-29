import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { billingApi, type BalanceTransaction } from '@/api'

const PAGE_SIZE = 20

interface BalanceTransactionsPageProps {
  admin?: boolean
}

export function BalanceTransactionsPage({ admin = false }: BalanceTransactionsPageProps) {
  const { t } = useTranslation()
  const [transactions, setTransactions] = useState<BalanceTransaction[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [userEmail, setUserEmail] = useState('')
  const [type, setType] = useState('all')

  const fetchTransactions = async (nextPage = page) => {
    setLoading(true)
    try {
      const params = {
        page: nextPage,
        size: PAGE_SIZE,
        type: type === 'all' ? undefined : Number(type),
      }
      const data = admin
        ? await billingApi.adminTransactions({
            ...params,
            user_email: userEmail.trim() || undefined,
          })
        : await billingApi.transactions(params)
      setTransactions(data.items || [])
      setTotal(data.total || 0)
      setPage(data.page || nextPage)
    } catch (err) {
      console.error('Failed to fetch balance transactions:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTransactions(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const formatUSD = (value?: number) => `$${((value || 0) / 1000000).toFixed(6)}`

  const getTypeLabel = (value: number) => {
    if (value === 1) return t('balanceTransactions.typeRecharge')
    if (value === 2) return t('balanceTransactions.typeConsume')
    if (value === 3) return t('balanceTransactions.typeRefund')
    if (value === 4) return t('balanceTransactions.typeGift')
    return String(value)
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('balanceTransactions.title')}</h2>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('balanceTransactions.list')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3 mb-4">
            {admin && (
              <Input
                placeholder={t('balanceTransactions.userIdPlaceholder')}
                value={userEmail}
                onChange={(e) => setUserEmail(e.target.value)}
                className="max-w-xs"
              />
            )}
            <Select value={type} onValueChange={setType}>
              <SelectTrigger className="w-44">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('balanceTransactions.allTypes')}</SelectItem>
                <SelectItem value="1">{t('balanceTransactions.typeRecharge')}</SelectItem>
                <SelectItem value="2">{t('balanceTransactions.typeConsume')}</SelectItem>
                <SelectItem value="3">{t('balanceTransactions.typeRefund')}</SelectItem>
                <SelectItem value="4">{t('balanceTransactions.typeGift')}</SelectItem>
              </SelectContent>
            </Select>
            <Button onClick={() => fetchTransactions(1)}>{t('common.search')}</Button>
          </div>

          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('balanceTransactions.time')}</TableHead>
                    {admin && <TableHead>{t('balanceTransactions.user')}</TableHead>}
                    <TableHead>{t('common.type')}</TableHead>
                    <TableHead>{t('balanceTransactions.amount')}</TableHead>
                    <TableHead>{t('balanceTransactions.balanceAfter')}</TableHead>
                    <TableHead>{t('balanceTransactions.description')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transactions.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={admin ? 6 : 5} className="text-center text-muted-foreground">
                        {t('common.noData')}
                      </TableCell>
                    </TableRow>
                  ) : (
                    transactions.map((tx) => (
                      <TableRow key={tx.id}>
                        <TableCell>{new Date(tx.created_at).toLocaleString()}</TableCell>
                        {admin && (
                          <TableCell>
                            <div className="font-medium">{tx.username || tx.user_id}</div>
                            <div className="text-xs text-muted-foreground">{tx.email || tx.user_id}</div>
                          </TableCell>
                        )}
                        <TableCell>{getTypeLabel(tx.type)}</TableCell>
                        <TableCell className={tx.amount >= 0 ? 'text-green-600' : 'text-red-600'}>
                          {tx.amount >= 0 ? '+' : ''}{formatUSD(tx.amount)}
                        </TableCell>
                        <TableCell>{formatUSD(tx.balance_after)}</TableCell>
                        <TableCell className="max-w-sm truncate">{tx.description || '-'}</TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>

              <div className="flex items-center justify-between mt-4 text-sm text-muted-foreground">
                <div>{t('common.total')}: {total}</div>
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => fetchTransactions(page - 1)}>
                    {t('common.previous')}
                  </Button>
                  <span>{page} / {totalPages}</span>
                  <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => fetchTransactions(page + 1)}>
                    {t('common.next')}
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
