import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { authApi } from '@/api'
import { useAuthStore } from '@/store/auth'

export default function LoginPage() {
  const { t, i18n } = useTranslation()
  const [email, setEmail] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [inviteCode, setInviteCode] = useState('')
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [loading, setLoading] = useState(false)
  const setTokens = useAuthStore((state) => state.setTokens)
  const clearTokens = useAuthStore((state) => state.clearTokens)
  const navigate = useNavigate()

  useEffect(() => {
    clearTokens()
  }, [clearTokens])

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      clearTokens()
      if (mode === 'register') {
        await authApi.register({ email, username, password, invite_code: inviteCode })
        toast.success(t('auth.registerSuccess'))
        setMode('login')
      } else {
        const data = await authApi.login({ email, password })
        setTokens(data.access_token, data.refresh_token)
        toast.success(t('auth.loginSuccess'))
        navigate('/dashboard')
      }
    } catch (err: any) {
      toast.error(err.message || t('common.networkError'))
    } finally {
      setLoading(false)
    }
  }

  const toggleLanguage = () => {
    const newLang = i18n.language === 'zh' ? 'en' : 'zh'
    i18n.changeLanguage(newLang)
    localStorage.setItem('language', newLang)
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-[400px]">
        <CardHeader className="space-y-1">
          <div className="flex justify-end">
            <Button variant="ghost" size="sm" onClick={toggleLanguage}>
              {i18n.language === 'zh' ? 'English' : '中文'}
            </Button>
          </div>
          <CardTitle className="text-2xl text-center">{mode === 'register' ? t('auth.registerTitle') : t('auth.loginTitle')}</CardTitle>
          <CardDescription className="text-center">
            {mode === 'register' ? t('auth.registerDescription') : t('auth.loginDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">{t('auth.email')}</Label>
              <Input
                id="email"
                type="email"
                placeholder="admin@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            {mode === 'register' && (
              <div className="space-y-2">
                <Label htmlFor="username">{t('auth.username')}</Label>
                <Input
                  id="username"
                  placeholder="username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                />
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="password">{t('auth.password')}</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            {mode === 'register' && (
              <div className="space-y-2">
                <Label htmlFor="inviteCode">{t('auth.inviteCode')}</Label>
                <Input
                  id="inviteCode"
                  value={inviteCode}
                  onChange={(e) => setInviteCode(e.target.value)}
                />
              </div>
            )}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? t('common.loading') : mode === 'register' ? t('auth.register') : t('auth.login')}
            </Button>
            <Button
              type="button"
              variant="ghost"
              className="w-full"
              onClick={() => setMode(mode === 'register' ? 'login' : 'register')}
            >
              {mode === 'register' ? t('auth.hasAccount') : t('auth.needInviteRegister')}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
