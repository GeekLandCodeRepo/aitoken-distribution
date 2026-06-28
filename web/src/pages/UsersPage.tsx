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
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { userApi, type InviteCode, type InviteSettings, type User, type UserChannelOption } from '@/api'
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
  const [inviteSettings, setInviteSettings] = useState<InviteSettings>({
    require_invite_register: false,
    user_invite_enabled: false,
    reward_amount: 0,
    new_user_bonus_amount: 0,
  })
  const [inviteCodes, setInviteCodes] = useState<InviteCode[]>([])
  const [generatedInviteCode, setGeneratedInviteCode] = useState<InviteCode | null>(null)
  const [channelsOpen, setChannelsOpen] = useState(false)
  const [channelsUser, setChannelsUser] = useState<User | null>(null)
  const [channelOptions, setChannelOptions] = useState<UserChannelOption[]>([])
  const [channelsLoading, setChannelsLoading] = useState(false)
  const [channelsSaving, setChannelsSaving] = useState(false)
  const [templateOpen, setTemplateOpen] = useState(false)
  const [templateOptions, setTemplateOptions] = useState<UserChannelOption[]>([])
  const [templateLoading, setTemplateLoading] = useState(false)
  const [templateSaving, setTemplateSaving] = useState(false)
  const [templateApplying, setTemplateApplying] = useState(false)
  const [applyConfirmOpen, setApplyConfirmOpen] = useState(false)

  useEffect(() => {
    fetchUsers()
    fetchInviteData()
  }, [])

  const fetchInviteData = async () => {
    try {
      const [settings, codes] = await Promise.all([
        userApi.inviteSettings(),
        userApi.listInviteCodes({ page: 1, size: 10 }),
      ])
      setInviteSettings({
        require_invite_register: settings.require_invite_register ?? false,
        user_invite_enabled: settings.user_invite_enabled ?? false,
        reward_amount: settings.reward_amount ?? 0,
        new_user_bonus_amount: settings.new_user_bonus_amount ?? 0,
      })
      setInviteCodes(codes.items || [])
    } catch (err) {
      console.error('Failed to fetch invite data:', err)
    }
  }

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

  const saveInviteSettings = async (settings: InviteSettings) => {
    try {
      const updated = await userApi.updateInviteSettings(settings)
      setInviteSettings(updated)
      toast.success(t('users.inviteSettingsSaved'))
    } catch (err: any) {
      toast.error(err.message || t('users.inviteSettingsFailed'))
    }
  }

  const createInviteCode = async () => {
    try {
      const code = await userApi.createInviteCode({
        reward_amount: inviteSettings.reward_amount,
        new_user_bonus: inviteSettings.new_user_bonus_amount,
      })
      setGeneratedInviteCode(code)
      toast.success(t('users.createInviteSuccess'))
      fetchInviteData()
    } catch (err: any) {
      toast.error(err.message || t('users.createInviteFailed'))
    }
  }

  const copyInviteCode = async () => {
    if (!generatedInviteCode) return

    await navigator.clipboard.writeText(generatedInviteCode.code)
    toast.success(t('users.inviteCopied'))
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

  const openTemplateDialog = async () => {
    setTemplateOpen(true)
    setTemplateLoading(true)
    try {
      const data = await userApi.getChannelTemplate()
      setTemplateOptions(data || [])
    } catch (err: any) {
      toast.error(err.message || t('users.channelTemplateFailed'))
      setTemplateOpen(false)
    } finally {
      setTemplateLoading(false)
    }
  }

  const toggleTemplateChannel = (channelId: string) => {
    setTemplateOptions((prev) =>
      prev.map((ch) => (ch.id === channelId ? { ...ch, allowed: !ch.allowed } : ch))
    )
  }

  const handleSaveTemplate = async () => {
    setTemplateSaving(true)
    try {
      const allowedIds = templateOptions.filter((ch) => ch.allowed).map((ch) => ch.id)
      const data = await userApi.updateChannelTemplate(allowedIds)
      setTemplateOptions(data || [])
      toast.success(t('users.channelTemplateSaved'))
    } catch (err: any) {
      toast.error(err.message || t('users.channelTemplateFailed'))
    } finally {
      setTemplateSaving(false)
    }
  }

  const handleApplyTemplateToAll = async () => {
    setApplyConfirmOpen(false)
    setTemplateApplying(true)
    try {
      const result = await userApi.applyChannelTemplateToAllUsers()
      const detail = result?.affected != null ? ` (${result.affected})` : ''
      toast.success(t('users.applyToAllSuccess', { detail }))
    } catch (err: any) {
      toast.error(err.message || t('users.applyToAllFailed'))
    } finally {
      setTemplateApplying(false)
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
        <CardHeader>
          <CardTitle>{t('users.inviteSettings')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid max-w-4xl gap-x-6 gap-y-3 md:grid-cols-[210px_340px_128px]">
            <div className="flex h-10 items-center justify-between gap-4">
              <div className="text-sm font-medium">{t('users.userInviteEnabled')}</div>
              <Switch
                checked={inviteSettings.user_invite_enabled}
                onCheckedChange={(checked) => saveInviteSettings({ ...inviteSettings, user_invite_enabled: checked })}
              />
            </div>
            <div className="flex h-10 items-center gap-3">
              <Label className="w-40 shrink-0 text-sm">{t('users.inviteRewardAmount')}</Label>
              <Input
                className="h-9 w-32"
                type="number"
                step="0.01"
                value={inviteSettings.reward_amount / 1000000}
                onChange={(e) => setInviteSettings({ ...inviteSettings, reward_amount: Math.floor((parseFloat(e.target.value) || 0) * 1000000) })}
              />
            </div>
            <Button className="h-9 w-28" onClick={createInviteCode}>{t('users.createInviteCode')}</Button>

            <div className="flex h-10 items-center justify-between gap-4">
              <div className="text-sm font-medium">{t('users.requireInviteRegister')}</div>
              <Switch
                checked={inviteSettings.require_invite_register}
                onCheckedChange={(checked) => saveInviteSettings({ ...inviteSettings, require_invite_register: checked })}
              />
            </div>
            <div className="flex h-10 items-center gap-3">
              <Label className="w-40 shrink-0 text-sm">{t('users.newUserBonusAmount')}</Label>
              <Input
                className="h-9 w-32"
                type="number"
                step="0.01"
                value={inviteSettings.new_user_bonus_amount / 1000000}
                onChange={(e) => setInviteSettings({ ...inviteSettings, new_user_bonus_amount: Math.floor((parseFloat(e.target.value) || 0) * 1000000) })}
              />
            </div>
            <Button className="h-9 w-28" variant="outline" onClick={() => saveInviteSettings(inviteSettings)}>{t('common.save')}</Button>
          </div>
          <div className="flex flex-wrap gap-2">
            {inviteCodes.map((code) => (
              <div key={code.id} className="flex items-center gap-2 rounded-md border px-2 py-1 text-xs">
                <span className="font-mono">{code.code}</span>
                <Badge variant="outline">{t('users.inviteRewardShort')}: {formatUSD(code.reward_amount)}</Badge>
                <Badge variant="outline">{t('users.newUserBonusShort')}: {formatUSD(code.new_user_bonus)}</Badge>
                {code.used_at && <Badge variant="secondary">{t('users.inviteUsed')}</Badge>}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t('users.list')}</CardTitle>
          <div className="flex gap-2">
            <Button variant="outline" onClick={openTemplateDialog}>{t('users.channelTemplate')}</Button>
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

      <Dialog open={templateOpen} onOpenChange={(open) => { if (!open) { setTemplateOpen(false); setTemplateOptions([]) } }}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{t('users.channelTemplate')}</DialogTitle>
            <DialogDescription>{t('users.channelTemplateDesc')}</DialogDescription>
          </DialogHeader>

          {templateLoading ? (
            <div className="py-8 text-center text-muted-foreground">{t('common.loading')}</div>
          ) : (
            <div className="flex-1 overflow-y-auto -mx-6 px-6 space-y-2">
              {templateOptions.length === 0 ? (
                <div className="py-8 text-center text-muted-foreground">{t('common.noData')}</div>
              ) : (
                <>
                  <div className="flex items-center justify-between py-2 sticky top-0 bg-background z-10 border-b mb-2">
                    <div className="flex items-center gap-2">
                      <Checkbox
                        checked={templateOptions.length > 0 && templateOptions.every((ch) => ch.allowed)}
                        onCheckedChange={(checked) => {
                          setTemplateOptions((prev) => prev.map((ch) => ({ ...ch, allowed: !!checked })))
                        }}
                      />
                      <span className="text-sm font-medium">
                        {templateOptions.filter((ch) => ch.allowed).length} / {templateOptions.length}
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {t('users.allowedCount', { count: templateOptions.filter((ch) => ch.allowed).length })}
                    </span>
                  </div>

                  {templateOptions.map((ch) => (
                    <label
                      key={ch.id}
                      className="flex items-start gap-3 rounded-md border p-3 cursor-pointer hover:bg-accent/50 transition-colors"
                    >
                      <Checkbox
                        checked={ch.allowed}
                        onCheckedChange={() => toggleTemplateChannel(ch.id)}
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

              {templateOptions.length > 0 && templateOptions.every((ch) => !ch.allowed) && (
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
            <div className="flex items-center gap-2">
              <Button
                variant="destructive"
                onClick={() => setApplyConfirmOpen(true)}
                disabled={templateLoading || templateSaving || templateApplying || templateOptions.length === 0}
              >
                {templateApplying ? t('common.loading') : t('users.applyToAll')}
              </Button>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => { setTemplateOpen(false); setTemplateOptions([]) }}>
                {t('common.cancel')}
              </Button>
              <Button onClick={handleSaveTemplate} disabled={templateSaving || templateLoading}>
                {templateSaving ? t('common.loading') : t('common.save')}
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={applyConfirmOpen} onOpenChange={setApplyConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive">{t('users.applyToAll')}</DialogTitle>
            <DialogDescription className="leading-relaxed">
              {t('users.applyToAllConfirm')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setApplyConfirmOpen(false)}>{t('common.cancel')}</Button>
            <Button variant="destructive" onClick={handleApplyTemplateToAll}>
              {t('common.confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!generatedInviteCode} onOpenChange={(open) => !open && setGeneratedInviteCode(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('users.inviteCodeGenerated')}</DialogTitle>
            <DialogDescription>{t('users.inviteCodeGeneratedDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <Label>{t('auth.inviteCode')}</Label>
            <Input readOnly value={generatedInviteCode?.code || ''} className="font-mono" />
            {generatedInviteCode && (
              <div className="grid gap-2 text-sm text-muted-foreground sm:grid-cols-2">
                <div>{t('users.inviteRewardAmount')}: {formatUSD(generatedInviteCode.reward_amount)}</div>
                <div>{t('users.newUserBonusAmount')}: {formatUSD(generatedInviteCode.new_user_bonus)}</div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGeneratedInviteCode(null)}>{t('common.close')}</Button>
            <Button onClick={copyInviteCode}>{t('common.copy')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
