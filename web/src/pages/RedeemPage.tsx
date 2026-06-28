
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { redeemApi } from '@/api/redeem'

export function RedeemPage() {
  const { t } = useTranslation()
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<{
    quota: number
    balance_before: number
    balance_after: number
  } | null>(null)
  const [error, setError] = useState('')

  const handleRedeem = async () => {
    if (!code.trim()) return

    setLoading(true)
    setError('')
    setResult(null)

    try {
      const data = await redeemApi.redeem(code.trim())
      setResult(data)
      setCode('')
    } catch (err: any) {
      setError(err.message || t('redeem.failed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">{t('redeem.title')}</h2>

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>{t('redeem.enterCode')}</CardTitle>
          <CardDescription>
            {t('redeem.enterCodeDesc')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t('redeem.redeemCode')}</Label>
              <Input
                placeholder="RC-XXXXXXXX"
                value={code}
                onChange={(e) => setCode(e.target.value)}
              />
            </div>

            <Button onClick={handleRedeem} disabled={loading || !code.trim()} className="w-full">
              {loading ? t('redeem.redeeming') : t('redeem.redeemBtn')}
            </Button>

            {error && (
              <div className="p-3 rounded-md text-sm bg-destructive/10 text-destructive">
                {error}
              </div>
            )}

            {result && (
              <div className="p-4 bg-green-50 border border-green-200 rounded-md">
                <div className="text-green-700 font-medium">{t('redeem.redeemSuccess')}</div>
                <div className="mt-2 space-y-1 text-sm text-green-600">
                  <div>{t('redeem.added')}: ${(result.quota / 1000000).toFixed(2)}</div>
                  <div>{t('redeem.balanceBefore')}: ${(result.balance_before / 1000000).toFixed(2)}</div>
                  <div>{t('redeem.balanceAfter')}: ${(result.balance_after / 1000000).toFixed(2)}</div>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
