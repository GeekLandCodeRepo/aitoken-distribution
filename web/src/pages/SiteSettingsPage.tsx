import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { userApi, type InviteCode, type InviteSettings, type UserChannelOption } from '@/api'

export function SiteSettingsPage() {
  const { t } = useTranslation()
  const [inviteSettings, setInviteSettings] = useState<InviteSettings>({
    require_invite_register: false,
    user_invite_enabled: false,
    reward_amount: 0,
    new_user_bonus_amount: 0,
  })
  const [inviteCodes, setInviteCodes] = useState<InviteCode[]>([])
  const [generatedInviteCode, setGeneratedInviteCode] = useState<InviteCode | null>(null)
  const [templateOptions, setTemplateOptions] = useState<UserChannelOption[]>([])
  const [templateLoading, setTemplateLoading] = useState(false)
  const [templateSaving, setTemplateSaving] = useState(false)
  const [templateApplying, setTemplateApplying] = useState(false)
  const [applyConfirmOpen, setApplyConfirmOpen] = useState(false)

  useEffect(() => {
    fetchInviteData()
    fetchChannelTemplate()
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

  const fetchChannelTemplate = async () => {
    setTemplateLoading(true)
    try {
      const data = await userApi.getChannelTemplate()
      setTemplateOptions(data || [])
    } catch (err: any) {
      toast.error(err.message || t('siteSettings.channelTemplateFailed'))
    } finally {
      setTemplateLoading(false)
    }
  }

  const saveInviteSettings = async (settings: InviteSettings) => {
    try {
      const updated = await userApi.updateInviteSettings(settings)
      setInviteSettings(updated)
      toast.success(t('siteSettings.inviteSettingsSaved'))
    } catch (err: any) {
      toast.error(err.message || t('siteSettings.inviteSettingsFailed'))
    }
  }

  const createInviteCode = async () => {
    try {
      const code = await userApi.createInviteCode({
        reward_amount: inviteSettings.reward_amount,
        new_user_bonus: inviteSettings.new_user_bonus_amount,
      })
      setGeneratedInviteCode(code)
      toast.success(t('siteSettings.createInviteSuccess'))
      fetchInviteData()
    } catch (err: any) {
      toast.error(err.message || t('siteSettings.createInviteFailed'))
    }
  }

  const copyInviteCode = async () => {
    if (!generatedInviteCode) return
    await navigator.clipboard.writeText(generatedInviteCode.code)
    toast.success(t('siteSettings.inviteCopied'))
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
      toast.success(t('siteSettings.channelTemplateSaved'))
    } catch (err: any) {
      toast.error(err.message || t('siteSettings.channelTemplateFailed'))
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
      toast.success(t('siteSettings.applyToAllSuccess', { detail }))
    } catch (err: any) {
      toast.error(err.message || t('siteSettings.applyToAllFailed'))
    } finally {
      setTemplateApplying(false)
    }
  }

  const formatUSD = (value?: number) => `$${((value || 0) / 1000000).toFixed(2)}`

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('siteSettings.title')}</h2>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('siteSettings.inviteSettings')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid max-w-4xl gap-x-6 gap-y-3 md:grid-cols-[210px_340px_128px]">
            <div className="flex h-10 items-center justify-between gap-4">
              <div className="text-sm font-medium">{t('siteSettings.userInviteEnabled')}</div>
              <Switch
                checked={inviteSettings.user_invite_enabled}
                onCheckedChange={(checked) => saveInviteSettings({ ...inviteSettings, user_invite_enabled: checked })}
              />
            </div>
            <div className="flex h-10 items-center gap-3">
              <Label className="w-40 shrink-0 text-sm">{t('siteSettings.inviteRewardAmount')}</Label>
              <Input
                className="h-9 w-32"
                type="number"
                step="0.01"
                value={inviteSettings.reward_amount / 1000000}
                onChange={(e) => setInviteSettings({ ...inviteSettings, reward_amount: Math.floor((parseFloat(e.target.value) || 0) * 1000000) })}
              />
            </div>
            <Button className="h-9 w-28" onClick={createInviteCode}>{t('siteSettings.createInviteCode')}</Button>

            <div className="flex h-10 items-center justify-between gap-4">
              <div className="text-sm font-medium">{t('siteSettings.requireInviteRegister')}</div>
              <Switch
                checked={inviteSettings.require_invite_register}
                onCheckedChange={(checked) => saveInviteSettings({ ...inviteSettings, require_invite_register: checked })}
              />
            </div>
            <div className="flex h-10 items-center gap-3">
              <Label className="w-40 shrink-0 text-sm">{t('siteSettings.newUserBonusAmount')}</Label>
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
                <Badge variant="outline">{t('siteSettings.inviteRewardShort')}: {formatUSD(code.reward_amount)}</Badge>
                <Badge variant="outline">{t('siteSettings.newUserBonusShort')}: {formatUSD(code.new_user_bonus)}</Badge>
                {code.used_at && <Badge variant="secondary">{t('siteSettings.inviteUsed')}</Badge>}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t('siteSettings.channelTemplate')}</CardTitle>
          <div className="flex gap-2">
            <Button
              variant="destructive"
              onClick={() => setApplyConfirmOpen(true)}
              disabled={templateLoading || templateSaving || templateApplying || templateOptions.length === 0}
            >
              {templateApplying ? t('common.loading') : t('siteSettings.applyToAll')}
            </Button>
            <Button onClick={handleSaveTemplate} disabled={templateSaving || templateLoading}>
              {templateSaving ? t('common.loading') : t('common.save')}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">{t('siteSettings.channelTemplateDesc')}</p>
          {templateLoading ? (
            <div className="py-8 text-center text-muted-foreground">{t('common.loading')}</div>
          ) : templateOptions.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">{t('common.noData')}</div>
          ) : (
            <div className="space-y-2">
              <div className="flex items-center justify-between py-2 border-b mb-2">
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
                  {t('siteSettings.allowedCount', { count: templateOptions.filter((ch) => ch.allowed).length })}
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

              {templateOptions.every((ch) => !ch.allowed) && (
                <div className="flex items-start gap-2 rounded-md border border-amber-500/50 bg-amber-500/10 p-3 text-sm">
                  <span className="text-amber-600 dark:text-amber-400 font-medium shrink-0">⚠</span>
                  <div>
                    <div className="font-medium text-amber-700 dark:text-amber-300">{t('siteSettings.userChannelsEmpty')}</div>
                    <div className="text-amber-600/80 dark:text-amber-400/70 text-xs mt-0.5">{t('siteSettings.userChannelsEmptyHint')}</div>
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={applyConfirmOpen} onOpenChange={setApplyConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive">{t('siteSettings.applyToAll')}</DialogTitle>
            <DialogDescription className="leading-relaxed">
              {t('siteSettings.applyToAllConfirm')}
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
            <DialogTitle>{t('siteSettings.inviteCodeGenerated')}</DialogTitle>
            <DialogDescription>{t('siteSettings.inviteCodeGeneratedDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <Label>{t('auth.inviteCode')}</Label>
            <Input readOnly value={generatedInviteCode?.code || ''} className="font-mono" />
            {generatedInviteCode && (
              <div className="grid gap-2 text-sm text-muted-foreground sm:grid-cols-2">
                <div>{t('siteSettings.inviteRewardAmount')}: {formatUSD(generatedInviteCode.reward_amount)}</div>
                <div>{t('siteSettings.newUserBonusAmount')}: {formatUSD(generatedInviteCode.new_user_bonus)}</div>
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
