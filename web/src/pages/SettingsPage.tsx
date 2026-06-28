import { useTranslation } from 'react-i18next'
import type { MouseEvent as ReactMouseEvent, ReactNode } from 'react'
import { useState } from 'react'
import { Check, ChevronRight, Globe, Lock, Moon, Palette, Sun } from 'lucide-react'
import { toast } from 'sonner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { authApi } from '@/api'
import { COLOR_THEMES, type ColorThemeDef, type Theme, useTheme } from '@/hooks/useTheme'
import { cn } from '@/lib/utils'

type ThemeModeOption = {
  id: Theme
  icon: ReactNode
  labelKey: string
  descriptionKey: string
}

const modeOptions: ThemeModeOption[] = [
  { id: 'light', icon: <Sun className="size-4" />, labelKey: 'settings.modeLight', descriptionKey: 'settings.modeLightDesc' },
  { id: 'dark', icon: <Moon className="size-4" />, labelKey: 'settings.modeDark', descriptionKey: 'settings.modeDarkDesc' },
]

/* ─── Reusable entry card for launching a dialog ─── */
function SettingEntry({
  icon,
  title,
  description,
  trailing,
  onClick,
}: {
  icon: ReactNode
  title: string
  description: string
  trailing?: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-3.5 rounded-lg border border-border bg-background p-4 text-left outline-none transition-all hover:border-primary/40 hover:bg-muted/35 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/35"
    >
      <span className="inline-flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
        {icon}
      </span>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-semibold text-foreground">{title}</div>
        <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
      </div>
      {trailing && (
        <span className="hidden shrink-0 text-xs text-muted-foreground sm:block">{trailing}</span>
      )}
      <ChevronRight className="size-4 shrink-0 text-muted-foreground/60" />
    </button>
  )
}

/* ─── Password dialog ─── */
function PasswordDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { t } = useTranslation()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const resetForm = () => {
    setOldPassword('')
    setNewPassword('')
    setConfirmPassword('')
    setError('')
  }

  const handleOpenChange = (next: boolean) => {
    onOpenChange(next)
    if (!next) resetForm()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (newPassword !== confirmPassword) {
      setError(t('settings.passwordMismatch'))
      return
    }
    if (newPassword.length < 8) {
      setError(t('settings.passwordTooShort'))
      return
    }

    setLoading(true)
    try {
      await authApi.changePassword(oldPassword, newPassword)
      toast.success(t('settings.passwordChangeSuccess'))
      handleOpenChange(false)
    } catch (err: any) {
      setError(err.message || t('settings.passwordChangeFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('settings.changePassword')}</DialogTitle>
          <DialogDescription>{t('settings.changePasswordDesc')}</DialogDescription>
        </DialogHeader>
        <form id="password-form" onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}
          <div className="space-y-1.5">
            <Label htmlFor="dlg-oldPassword">{t('settings.currentPassword')}</Label>
            <Input
              id="dlg-oldPassword"
              type="password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              required
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="dlg-newPassword">{t('settings.newPassword')}</Label>
            <Input
              id="dlg-newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="dlg-confirmPassword">{t('settings.confirmNewPassword')}</Label>
            <Input
              id="dlg-confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
        </form>
        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button type="submit" form="password-form" disabled={loading}>
            {loading ? t('settings.changing') : t('settings.changeBtn')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

/* ─── Language dialog ─── */
const LANGUAGES = [
  { code: 'zh', label: '中文', flag: '🇨🇳' },
  { code: 'en', label: 'English', flag: '🇬🇧' },
] as const

function LanguageDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { t, i18n } = useTranslation()

  const handleSelect = (code: string) => {
    i18n.changeLanguage(code)
    localStorage.setItem('language', code)
    toast.success(t('settings.languageChanged'))
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('settings.language')}</DialogTitle>
          <DialogDescription>{t('settings.languageDesc')}</DialogDescription>
        </DialogHeader>
        <div className="grid gap-2">
          {LANGUAGES.map((lang) => {
            const active = i18n.language === lang.code
            return (
              <button
                key={lang.code}
                type="button"
                onClick={() => handleSelect(lang.code)}
                className={cn(
                  'flex items-center gap-3 rounded-lg border p-3 text-left outline-none transition-all hover:border-primary/40 hover:bg-muted/35 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/35',
                  active ? 'border-primary/55 bg-primary/10' : 'border-border',
                )}
              >
                <span className="text-xl leading-none">{lang.flag}</span>
                <span className="flex-1 text-sm font-medium text-foreground">{lang.label}</span>
                {active && <Check className="size-4 text-primary" />}
              </button>
            )
          })}
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ─── Preview mini-card ─── */
function ThemePreviewCard() {
  const { t } = useTranslation()

  return (
    <div className="min-w-0">
      <h4 className="mb-1 text-sm font-semibold text-foreground">{t('settings.previewTitle')}</h4>
      <p className="mb-3 text-xs leading-relaxed text-muted-foreground">{t('settings.previewDesc')}</p>
      <div className="overflow-hidden rounded-lg border border-border bg-background shadow-sm">
        <div className="flex min-h-[220px] min-w-0">
          <div className="w-[132px] shrink-0 border-r border-border bg-sidebar p-3 max-sm:w-[96px]">
            <div className="mb-4 flex items-center gap-2">
              <span className="size-7 rounded-lg bg-primary/15 ring-1 ring-primary/20" />
              <span className="h-3 w-14 rounded-full bg-foreground/16 max-sm:hidden" />
            </div>
            <div className="space-y-2">
              <span className="block h-7 rounded-md bg-primary/12 ring-1 ring-primary/20" />
              <span className="block h-7 rounded-md bg-muted/70" />
              <span className="block h-7 rounded-md bg-muted/50" />
            </div>
          </div>
          <div className="min-w-0 flex-1 p-4">
            <div className="mb-4 flex items-start justify-between gap-3">
              <div className="min-w-0 flex-1">
                <span className="mb-2 block h-4 w-36 max-w-full rounded-full bg-foreground/18" />
                <span className="block h-3 w-56 max-w-full rounded-full bg-muted" />
              </div>
              <Button size="sm">
                <Palette className="size-3.5" />
                {t('settings.previewAction')}
              </Button>
            </div>
            <div className="grid gap-3 sm:grid-cols-3">
              {[0, 1, 2].map((item) => (
                <div key={item} className="min-h-[74px] rounded-lg border border-border bg-card p-3">
                  <span className="mb-3 block h-3 w-16 rounded-full bg-muted" />
                  <span className={cn('block h-5 rounded-full', item === 0 ? 'w-20 bg-primary/18' : item === 1 ? 'w-16 bg-emerald-500/18' : 'w-24 bg-amber-500/18')} />
                </div>
              ))}
            </div>
            <div className="mt-3 overflow-hidden rounded-lg border border-border bg-card">
              {[0, 1, 2].map((row) => (
                <div key={row} className="grid grid-cols-[1fr_72px_56px] items-center gap-3 border-b border-border px-3 py-2.5 last:border-b-0">
                  <span className="h-3 rounded-full bg-muted" />
                  <span className="h-3 rounded-full bg-muted/80" />
                  <span className={cn('h-5 rounded-full', row === 0 ? 'bg-primary/18' : 'bg-muted/70')} />
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Theme style card ─── */
function ThemeStyleCard({ item, active, onSelect }: { item: ColorThemeDef; active: boolean; onSelect: () => void }) {
  const { t } = useTranslation()

  return (
    <button
      type="button"
      aria-pressed={active}
      onClick={onSelect}
      className={cn(
        'group min-w-0 rounded-lg border bg-card p-2.5 text-left shadow-sm outline-none transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/35',
        active ? 'border-primary/55 ring-2 ring-primary/20' : 'border-border',
      )}
    >
      <div className="relative aspect-[16/10] overflow-hidden rounded-md border shadow-inner" style={{ backgroundColor: item.previewBg, borderColor: item.previewMuted }}>
        <div className="absolute inset-y-0 left-0 w-[28%] border-r" style={{ backgroundColor: item.previewSurface, borderColor: item.previewMuted }}>
          <span className="mx-2 mt-2 block h-4 rounded-md" style={{ backgroundColor: item.previewPrimary }} />
          <span className="mx-2 mt-2 block h-2 rounded-full" style={{ backgroundColor: item.previewMuted }} />
          <span className="mx-2 mt-1.5 block h-2 rounded-full" style={{ backgroundColor: item.previewMuted }} />
        </div>
        <div className="ml-[28%] p-2.5">
          <span className="mb-2 block h-3 w-20 rounded-full" style={{ backgroundColor: item.previewPrimary }} />
          <div className="grid grid-cols-2 gap-1.5">
            <span className="h-8 rounded-md" style={{ backgroundColor: item.previewSurface }} />
            <span className="h-8 rounded-md" style={{ backgroundColor: item.previewMuted }} />
          </div>
          <span className="mt-2 block h-2 rounded-full" style={{ backgroundColor: item.previewMuted }} />
          <span className="mt-1.5 block h-2 w-2/3 rounded-full" style={{ backgroundColor: item.previewMuted }} />
        </div>
        {active && (
          <span className="absolute right-2 top-2 inline-flex size-6 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-sm">
            <Check className="size-3.5" />
          </span>
        )}
      </div>
      <div className="mt-2.5 flex min-w-0 items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold text-foreground">{t(item.nameKey)}</div>
          <p className="mt-0.5 line-clamp-2 text-xs leading-relaxed text-muted-foreground">{t(item.descriptionKey)}</p>
        </div>
        <div className="flex shrink-0 items-center pt-0.5">
          <span className="size-3 rounded-full border border-black/5 shadow-inner" style={{ backgroundColor: item.previewPrimary }} />
          <span className="-ml-1 size-3 rounded-full border border-black/5 shadow-inner" style={{ backgroundColor: item.previewBg }} />
        </div>
      </div>
    </button>
  )
}

/* ─── Page ─── */
export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme, colorTheme, setColorTheme } = useTheme()
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [languageOpen, setLanguageOpen] = useState(false)

  const currentLangLabel = i18n.language === 'zh' ? '中文' : 'English'

  const handleModeChange = (nextTheme: Theme, event: ReactMouseEvent<HTMLButtonElement>) => {
    setTheme(nextTheme, event)
  }

  return (
    <div className="mx-auto w-full max-w-5xl space-y-8">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground">{t('settings.title')}</h1>
        <p className="mt-1.5 text-sm text-muted-foreground">{t('settings.appearanceDesc')}</p>
      </div>

      {/* Account & Language — entry buttons */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid gap-3 sm:grid-cols-2">
            <SettingEntry
              icon={<Lock className="size-4" />}
              title={t('settings.changePassword')}
              description={t('settings.changePasswordDesc')}
              onClick={() => setPasswordOpen(true)}
            />
            <SettingEntry
              icon={<Globe className="size-4" />}
              title={t('settings.language')}
              description={t('settings.languageDesc')}
              trailing={currentLangLabel}
              onClick={() => setLanguageOpen(true)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Appearance — main visual section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Palette className="size-5 text-primary" />
            {t('settings.appearance')}
          </CardTitle>
          <CardDescription>{t('settings.appearanceDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-8">
          {/* Mode selector + Preview */}
          <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,0.75fr)_minmax(0,1.25fr)]">
            <div>
              <h4 className="mb-1 text-sm font-semibold text-foreground">{t('settings.modeTitle')}</h4>
              <p className="mb-3 text-xs leading-relaxed text-muted-foreground">{t('settings.modeDesc')}</p>
              <div className="grid min-w-0 gap-2 sm:grid-cols-2 xl:grid-cols-1">
                {modeOptions.map((item) => {
                  const active = theme === item.id
                  return (
                    <button
                      key={item.id}
                      type="button"
                      aria-pressed={active}
                      onClick={(event) => handleModeChange(item.id, event)}
                      className={cn(
                        'flex min-h-[72px] min-w-0 items-start gap-3 rounded-lg border p-3 text-left outline-none transition-all hover:border-primary/40 hover:bg-muted/35 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/35',
                        active ? 'border-primary/55 bg-primary/10 text-primary' : 'border-border bg-background text-foreground',
                      )}
                    >
                      <span className={cn('mt-0.5 inline-flex size-8 shrink-0 items-center justify-center rounded-lg', active ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground')}>
                        {item.icon}
                      </span>
                      <span className="min-w-0">
                        <span className="flex items-center gap-1.5 text-sm font-semibold">
                          {t(item.labelKey)}
                          {active ? <Check className="size-3.5" /> : null}
                        </span>
                        <span className="mt-0.5 block text-xs leading-relaxed text-muted-foreground">{t(item.descriptionKey)}</span>
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>

            <ThemePreviewCard />
          </div>

          <div className="h-px bg-border" />

          {/* Color themes */}
          <section>
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-foreground">{t('settings.stylesTitle')}</h4>
              <p className="mt-0.5 text-xs leading-relaxed text-muted-foreground">{t('settings.stylesDesc')}</p>
            </div>
            <div className="grid gap-3 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
              {COLOR_THEMES.map((item) => (
                <ThemeStyleCard key={item.id} item={item} active={item.id === colorTheme} onSelect={() => setColorTheme(item.id)} />
              ))}
            </div>
          </section>
        </CardContent>
      </Card>

      {/* Dialogs */}
      <PasswordDialog open={passwordOpen} onOpenChange={setPasswordOpen} />
      <LanguageDialog open={languageOpen} onOpenChange={setLanguageOpen} />
    </div>
  )
}
