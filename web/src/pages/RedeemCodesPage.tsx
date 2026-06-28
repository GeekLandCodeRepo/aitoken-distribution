
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { redeemApi, type RedeemCode, type GenerateCodesRequest } from '@/api/redeem'

export function RedeemCodesPage() {
  const { t } = useTranslation()
  const [codes, setCodes] = useState<RedeemCode[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [generatedCodes, setGeneratedCodes] = useState<string[] | null>(null)
  const [formData, setFormData] = useState<GenerateCodesRequest>({
    quota: 10000000,
    count: 10,
  })

  useEffect(() => {
    fetchCodes()
  }, [])

  const fetchCodes = async () => {
    setLoading(true)
    try {
      const data = await redeemApi.list()
      setCodes(data.items || [])
    } catch (err) {
      console.error('Failed to fetch codes:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleGenerate = async () => {
    try {
      const data = await redeemApi.generate(formData)
      setGeneratedCodes(data.codes)
      fetchCodes()
    } catch (err: any) {
      alert(err.message || t('redeemCodes.generateFailed'))
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(t('redeemCodes.deleteConfirm'))) return

    try {
      await redeemApi.delete(id)
      fetchCodes()
    } catch (err: any) {
      alert(err.message || t('redeemCodes.deleteFailed'))
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    alert(t('redeemCodes.copied'))
  }

  const copyAllCodes = () => {
    if (generatedCodes) {
      navigator.clipboard.writeText(generatedCodes.join('\n'))
      alert(t('redeemCodes.allCodesCopied'))
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('redeemCodes.title')}</h2>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>{t('redeemCodes.generate')}</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t('redeemCodes.generateTitle')}</DialogTitle>
              <DialogDescription>{t('redeemCodes.generateDesc')}</DialogDescription>
            </DialogHeader>
            {generatedCodes ? (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-md max-h-60 overflow-y-auto">
                  <div className="space-y-1">
                    {generatedCodes.map((code, index) => (
                      <div key={index} className="font-mono text-sm">{code}</div>
                    ))}
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button onClick={copyAllCodes}>{t('redeemCodes.copyAll')}</Button>
                  <Button variant="outline" onClick={() => {
                    setGeneratedCodes(null)
                    setDialogOpen(false)
                  }}>
                    {t('common.close')}
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t('redeemCodes.quotaPerCode')}</Label>
                  <Input
                    type="number"
                    value={formData.quota}
                    onChange={(e) => setFormData({ ...formData, quota: parseInt(e.target.value) || 0 })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('redeemCodes.numberOfCodes')}</Label>
                  <Input
                    type="number"
                    value={formData.count}
                    onChange={(e) => setFormData({ ...formData, count: parseInt(e.target.value) || 1 })}
                  />
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
                  <Button onClick={handleGenerate}>{t('redeemCodes.generate')}</Button>
                </DialogFooter>
              </div>
            )}
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('redeemCodes.title')}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('redeemCodes.code')}</TableHead>
                  <TableHead>{t('redeemCodes.quota')}</TableHead>
                  <TableHead>{t('common.status')}</TableHead>
                  <TableHead>{t('redeemCodes.usedBy')}</TableHead>
                  <TableHead>{t('redeemCodes.expiresAt')}</TableHead>
                  <TableHead>{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {codes.map((code) => (
                  <TableRow key={code.id}>
                    <TableCell className="font-mono">{code.code}</TableCell>
                    <TableCell>${(code.quota / 1000000).toFixed(2)}</TableCell>
                    <TableCell>
                      <span className={code.used_by ? 'text-muted-foreground' : 'text-green-600'}>
                        {code.used_by ? t('redeemCodes.used') : t('redeemCodes.available')}
                      </span>
                    </TableCell>
                    <TableCell>{code.used_by || '-'}</TableCell>
                    <TableCell>
                      {code.expires_at ? new Date(code.expires_at).toLocaleDateString() : t('redeemCodes.never')}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => copyToClipboard(code.code)}
                        >
                          {t('common.copy')}
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => handleDelete(code.id)}
                        >
                          {t('common.delete')}
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
    </div>
  )
}
