import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { authApi } from '@/api'

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState({ type: '', text: '' })

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage({ type: '', text: '' })

    if (newPassword !== confirmPassword) {
      setMessage({ type: 'error', text: t('settings.passwordMismatch') })
      return
    }

    if (newPassword.length < 8) {
      setMessage({ type: 'error', text: t('settings.passwordTooShort') })
      return
    }

    setLoading(true)
    try {
      await authApi.changePassword(oldPassword, newPassword)
      setMessage({ type: 'success', text: t('settings.passwordChangeSuccess') })
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err: any) {
      setMessage({ type: 'error', text: err.message || t('settings.passwordChangeFailed') })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">{t('settings.title')}</h2>

      <Card>
        <CardHeader>
          <CardTitle>{t('settings.changePassword')}</CardTitle>
          <CardDescription>{t('settings.changePasswordDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleChangePassword} className="space-y-4">
            {message.text && (
              <div
                className={`p-3 rounded-md text-sm ${
                  message.type === 'success'
                    ? 'bg-green-50 text-green-700 border border-green-200'
                    : 'bg-destructive/10 text-destructive'
                }`}
              >
                {message.text}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="oldPassword">{t('settings.currentPassword')}</Label>
              <Input
                id="oldPassword"
                type="password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="newPassword">{t('settings.newPassword')}</Label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">{t('settings.confirmNewPassword')}</Label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" disabled={loading}>
              {loading ? t('settings.changing') : t('settings.changeBtn')}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>语言 / Language</CardTitle>
          <CardDescription>切换界面语言</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Button
              variant={i18n.language === 'zh' ? 'default' : 'outline'}
              onClick={() => {
                i18n.changeLanguage('zh')
                localStorage.setItem('language', 'zh')
              }}
            >
              中文
            </Button>
            <Button
              variant={i18n.language === 'en' ? 'default' : 'outline'}
              onClick={() => {
                i18n.changeLanguage('en')
                localStorage.setItem('language', 'en')
              }}
            >
              English
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
