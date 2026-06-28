import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { userApi, type User, type UserChannelOption } from '@/api'
import { Checkbox } from '@/components/ui/checkbox'

export function UsersPage() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [topUpOpen, setTopUpOpen] = useState(false)
  const [topUpUser, setTopUpUser] = useState<User | null>(null)
  const [topUpAmount, setTopUpAmount] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [resetPasswordUser, setResetPasswordUser] = useState<User | null>(null)
  const [newUser, setNewUser] = useState({ email: '', username: '', password: '', role: 1, balance: '' })
  const [channelsOpen, setChannelsOpen] = useState(false)
  const [channelsUser, setChannelsUser] = useState<User | null>(null)
  const [channelOptions, setChannelOptions] = useState<UserChannelOption[]>([])
  const [channelsLoading, setChannelsLoading] = useState(false)
  const [channelsSaving, setChannelsSaving] = useState(false)

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    setLoading(true)
    try {
      const data = await userApi.list({ search })
      setUsers(data.items || [])
    } catch (err) {
      console.error('Failed to fetch users:', err)
    } finally {
      setLoading(false)
    }
  }

  const openTopUpDialog = (user: User) => {
    setTopUpUser(user)
    setTopUpAmount('')
    setTopUpOpen(true)
  }

  const handleTopUp = async () => {
    if (!topUpUser || !topUpAmount) return

    const amount = parseFloat(topUpAmount)
    if (!Number.isFinite(amount) || amount <= 0) return

    try {
      await userApi.topUp(topUpUser.id, {
        amount: Math.floor(amount * 1000000),
        description: 'Admin top up',
      })
      toast.success(t('users.topUpSuccess'))
      setTopUpOpen(false)
      setTopUpUser(null)
      setTopUpAmount('')
      fetchUsers()
    } catch (err: any) {
      toast.error(err.message || t('users.topUpFailed'))
    }
  }

  const handleToggleStatus = async (userId: string, currentStatus: number) => {
    try {
      await userApi.update(userId, { status: currentStatus === 1 ? 0 : 1 })
      toast.success(t('users.updateSuccess'))
      fetchUsers()
    } catch {
      toast.error(t('users.updateFailed'))
    }
  }

  const handleResetPassword = async (user: User) => {
    try {
      const result = await userApi.resetPassword(user.id)
      toast.success(t('users.resetPasswordSuccess', { password: result.default_password }))
      setResetPasswordUser(null)
    } catch (err: any) {
      toast.error(err.message || t('users.resetPasswordFailed'))
    }
  }

  const handleCreateUser = async () => {
    try {
      await userApi.create({
        email: newUser.email,
        username: newUser.username,
        password: newUser.password,
        role: newUser.role,
        balance: Math.floor((parseFloat(newUser.balance) || 0) * 1000000),
      })
      setCreateOpen(false)
      setNewUser({ email: '', username: '', password: '', role: 1, balance: '' })
      toast.success(t('users.createSuccess'))
      fetchUsers()
    } catch (err: any) {
      toast.error(err.message || t('users.createFailed'))
    }
  }

  const openChannelsDialog = async (user: User) => {
    setChannelsUser(user)
    setChannelsOpen(true)
    setChannelsLoading(true)
    try {
      const data = await userApi.listChannels(user.id)
      setChannelOptions(data || [])
    } catch (err: any) {
      toast.error(err.message || t('users.userChannelsFailed'))
      setChannelsOpen(false)
    } finally {
      setChannelsLoading(false)
    }
  }

  const toggleChannel = (channelId: string) => {
    setChannelOptions((prev) =>
      prev.map((ch) => (ch.id === channelId ? { ...ch, allowed: !ch.allowed } : ch))
    )
  }

  const handleSaveChannels = async () => {
    if (!channelsUser) return
    setChannelsSaving(true)
    try {
      const allowedIds = channelOptions.filter((ch) => ch.allowed).map((ch) => ch.id)
      const data = await userApi.updateChannels(channelsUser.id, allowedIds)
      setChannelOptions(data || [])
      toast.success(t('users.userChannelsSaved'))
      setChannelsOpen(false)
    } catch (err: any) {
      toast.error(err.message || t('users.userChannelsFailed'))
    } finally {
      setChannelsSaving(false)
    }
  }

  const formatUSD = (value?: number) => `$${((value || 0) / 1000000).toFixed(2)}`

  const getTotalQuota = (user: User) => (user.balance || 0) + (user.used_quota || 0)

  const getQuotaPercent = (user: User) => {
    const total = getTotalQuota(user)
    if (total <= 0) return 0
    return Math.min(100, Math.max(0, ((user.used_quota || 0) / total) * 100))
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('users.title')}</h2>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t('users.list')}</CardTitle>
          <div className="flex gap-2">
            <Button onClick={() => setCreateOpen(true)}>{t('users.createUser')}</Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 mb-4">
            <Input
              placeholder={t('users.searchPlaceholder')}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="max-w-sm"
            />
            <Button onClick={fetchUsers}>{t('common.search')}</Button>
          </div>

          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('users.username')}</TableHead>
                  <TableHead>{t('users.email')}</TableHead>
                  <TableHead>{t('users.role')}</TableHead>
                  <TableHead>{t('users.quota')}</TableHead>
                  <TableHead>{t('users.requests')}</TableHead>
                  <TableHead>{t('common.status')}</TableHead>
                  <TableHead>{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.username}</TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>
                      <Badge variant={user.role >= 10 ? 'default' : 'secondary'}>
                        {user.role >= 10 ? t('common.admin') : t('common.user')}
                      </Badge>
                    </TableCell>
                    <TableCell className="w-56 min-w-56">
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between gap-2 text-xs">
                          <span className="font-medium">
                            {formatUSD(user.used_quota)} / {formatUSD(getTotalQuota(user))}
                          </span>
                          <span className="text-muted-foreground">{getQuotaPercent(user).toFixed(0)}%</span>
                        </div>
                        <Progress value={getQuotaPercent(user)} />
                        <div className="text-xs text-muted-foreground">
                          {t('users.availableQuota')}: {formatUSD(user.balance)}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{user.request_count}</TableCell>
                    <TableCell>
                      <Badge variant={user.status === 1 ? 'default' : 'destructive'}>
                        {user.status === 1 ? t('common.active') : t('common.disabled')}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => openTopUpDialog(user)}
                        >
                          {t('users.topUp')}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => openChannelsDialog(user)}
                        >
                          {t('users.userChannels')}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setResetPasswordUser(user)}
                        >
                          {t('users.resetPassword')}
                        </Button>
                        <Button
                          size="sm"
                          variant={user.status === 1 ? 'destructive' : 'default'}
                          onClick={() => handleToggleStatus(user.id, user.status)}
                        >
                          {user.status === 1 ? t('users.disable') : t('users.enable')}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={topUpOpen} onOpenChange={setTopUpOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('users.topUp')}</DialogTitle>
            <DialogDescription>
              {topUpUser ? `${topUpUser.username} (${topUpUser.email})` : ''}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label>{t('users.topUpAmount')}</Label>
            <Input
              type="number"
              min="0"
              step="0.01"
              value={topUpAmount}
              onChange={(e) => setTopUpAmount(e.target.value)}
              placeholder="10"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTopUpOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleTopUp}>{t('users.topUp')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('users.createUser')}</DialogTitle>
            <DialogDescription>{t('users.createUserDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t('users.email')}</Label>
              <Input value={newUser.email} onChange={(e) => setNewUser({ ...newUser, email: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>{t('users.username')}</Label>
              <Input value={newUser.username} onChange={(e) => setNewUser({ ...newUser, username: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>{t('auth.password')}</Label>
              <Input type="password" value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>{t('users.topUpAmount')}</Label>
              <Input type="number" value={newUser.balance} onChange={(e) => setNewUser({ ...newUser, balance: e.target.value })} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleCreateUser}>{t('common.create')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!resetPasswordUser} onOpenChange={(open) => !open && setResetPasswordUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('users.resetPassword')}</DialogTitle>
            <DialogDescription>
              {resetPasswordUser ? t('users.resetPasswordConfirm', { username: resetPasswordUser.username }) : ''}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setResetPasswordUser(null)}>{t('common.cancel')}</Button>
            <Button onClick={() => resetPasswordUser && handleResetPassword(resetPasswordUser)}>
              {t('users.resetPassword')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={channelsOpen} onOpenChange={(open) => { if (!open) { setChannelsOpen(false); setChannelsUser(null); setChannelOptions([]) } }}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{t('users.userChannels')}</DialogTitle>
            <DialogDescription>
              {channelsUser ? t('users.userChannelsDesc', { username: channelsUser.username }) : ''}
            </DialogDescription>
          </DialogHeader>

          {channelsLoading ? (
            <div className="py-8 text-center text-muted-foreground">{t('common.loading')}</div>
          ) : (
            <div className="flex-1 overflow-y-auto -mx-6 px-6 space-y-2">
              {channelOptions.length === 0 ? (
                <div className="py-8 text-center text-muted-foreground">{t('common.noData')}</div>
              ) : (
                <>
                  <div className="flex items-center justify-between py-2 sticky top-0 bg-background z-10 border-b mb-2">
                    <div className="flex items-center gap-2">
                      <Checkbox
                        checked={channelOptions.length > 0 && channelOptions.every((ch) => ch.allowed)}
                        onCheckedChange={(checked) => {
                          setChannelOptions((prev) => prev.map((ch) => ({ ...ch, allowed: !!checked })))
                        }}
                      />
                      <span className="text-sm font-medium">
                        {channelOptions.filter((ch) => ch.allowed).length} / {channelOptions.length}
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {t('users.allowedCount', { count: channelOptions.filter((ch) => ch.allowed).length })}
                    </span>
                  </div>

                  {channelOptions.map((ch) => (
                    <label
                      key={ch.id}
                      className="flex items-start gap-3 rounded-md border p-3 cursor-pointer hover:bg-accent/50 transition-colors"
                    >
                      <Checkbox
                        checked={ch.allowed}
                        onCheckedChange={() => toggleChannel(ch.id)}
                        className="mt-0.5"
                      />
                      <div className="flex-1 min-w-0 space-y-1">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-sm">{ch.name}</span>
                          <Badge variant={ch.status === 1 ? 'default' : 'destructive'} className="text-xs">
                            {ch.status === 1 ? t('common.active') : t('common.disabled')}
                          </Badge>
                        </div>
                        <div className="text-xs text-muted-foreground truncate">{ch.base_url}</div>
                        {ch.models.length > 0 && (
                          <div className="flex flex-wrap gap-1 pt-1">
                            {ch.models.slice(0, 8).map((model) => (
                              <Badge key={model} variant="secondary" className="text-xs font-mono">
                                {model}
                              </Badge>
                            ))}
                            {ch.models.length > 8 && (
                              <Badge variant="secondary" className="text-xs">
                                +{ch.models.length - 8}
                              </Badge>
                            )}
                          </div>
                        )}
                      </div>
                    </label>
                  ))}
                </>
              )}

              {channelOptions.length > 0 && channelOptions.every((ch) => !ch.allowed) && (
                <div className="flex items-start gap-2 rounded-md border border-amber-500/50 bg-amber-500/10 p-3 text-sm">
                  <span className="text-amber-600 dark:text-amber-400 font-medium shrink-0">⚠</span>
                  <div>
                    <div className="font-medium text-amber-700 dark:text-amber-300">{t('users.userChannelsEmpty')}</div>
                    <div className="text-amber-600/80 dark:text-amber-400/70 text-xs mt-0.5">{t('users.userChannelsEmptyHint')}</div>
                  </div>
                </div>
              )}
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => { setChannelsOpen(false); setChannelsUser(null); setChannelOptions([]) }}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSaveChannels} disabled={channelsSaving || channelsLoading}>
              {channelsSaving ? t('common.loading') : t('common.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

    </div>
  )
}
