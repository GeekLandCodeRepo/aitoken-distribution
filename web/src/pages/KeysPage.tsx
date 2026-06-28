
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
import { Badge } from '@/components/ui/badge'
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
import { apiKeyApi, type ApiKey, type CreateKeyRequest } from '@/api/key'

export function KeysPage() {
  const { t } = useTranslation()
  const [keys, setKeys] = useState<ApiKey[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [newKey, setNewKey] = useState<CreateKeyRequest>({ name: '' })
  const [createdKey, setCreatedKey] = useState<string | null>(null)

  useEffect(() => {
    fetchKeys()
  }, [])

  const fetchKeys = async () => {
    setLoading(true)
    try {
      const data = await apiKeyApi.list()
      setKeys(data || [])
    } catch (err) {
      console.error('Failed to fetch keys:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async () => {
    if (!newKey.name) return

    try {
      const data = await apiKeyApi.create(newKey)
      setCreatedKey(data.key)
      fetchKeys()
    } catch (err: any) {
      alert(err.message || t('keys.createFailed'))
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(t('keys.deleteConfirm'))) return

    try {
      await apiKeyApi.delete(id)
      fetchKeys()
    } catch (err: any) {
      alert(err.message || t('keys.deleteFailed'))
    }
  }

  const handleToggle = async (id: string) => {
    try {
      await apiKeyApi.toggle(id)
      fetchKeys()
    } catch (err) {
      console.error('Failed to toggle key:', err)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    alert(t('keys.copied'))
  }

  const formatKeyDisplay = (key: ApiKey) => {
    if (key.key_suffix) {
      return `${key.key_prefix}......${key.key_suffix}`
    }
    return `${key.key_prefix}......`
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('keys.title')}</h2>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>{t('keys.createKey')}</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t('keys.createTitle')}</DialogTitle>
              <DialogDescription>{t('keys.createDesc')}</DialogDescription>
            </DialogHeader>
            {createdKey ? (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-md">
                  <Label>{t('keys.yourKey')}</Label>
                  <div className="mt-2 font-mono text-sm break-all">{createdKey}</div>
                </div>
                <div className="flex gap-2">
                  <Button onClick={() => copyToClipboard(createdKey)}>
                    {t('common.copy')}
                  </Button>
                  <Button variant="outline" onClick={() => {
                    setCreatedKey(null)
                    setDialogOpen(false)
                    setNewKey({ name: '' })
                  }}>
                    {t('common.close')}
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t('common.name')}</Label>
                  <Input
                    placeholder={t('keys.keyNamePlaceholder')}
                    value={newKey.name}
                    onChange={(e) => setNewKey({ ...newKey, name: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('keys.quotaLimit')}</Label>
                  <Input
                    type="number"
                    placeholder="-1"
                    value={newKey.quota_limit || ''}
                    onChange={(e) => setNewKey({ ...newKey, quota_limit: parseInt(e.target.value) || -1 })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t('keys.rateLimit')}</Label>
                  <Input
                    type="number"
                    placeholder="-1"
                    value={newKey.rate_limit || ''}
                    onChange={(e) => setNewKey({ ...newKey, rate_limit: parseInt(e.target.value) || -1 })}
                  />
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setDialogOpen(false)}>
                    {t('common.cancel')}
                  </Button>
                  <Button onClick={handleCreate}>{t('common.create')}</Button>
                </DialogFooter>
              </div>
            )}
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('keys.yourKeys')}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('common.name')}</TableHead>
                  <TableHead>{t('keys.keyPrefix')}</TableHead>
                  <TableHead>{t('keys.quota')}</TableHead>
                  <TableHead>{t('keys.rateLimitHeader')}</TableHead>
                  <TableHead>{t('common.status')}</TableHead>
                  <TableHead>{t('keys.lastUsed')}</TableHead>
                  <TableHead>{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((key) => (
                  <TableRow key={key.id}>
                    <TableCell className="font-medium">{key.name}</TableCell>
                    <TableCell className="font-mono">{formatKeyDisplay(key)}</TableCell>
                    <TableCell>
                      {key.quota_limit === -1 ? t('common.unlimited') : `$${(key.used_quota / 1000000).toFixed(2)} / ${(key.quota_limit / 1000000).toFixed(2)}`}
                    </TableCell>
                    <TableCell>
                      {key.rate_limit === -1 ? t('common.unlimited') : `${key.rate_limit}/min`}
                    </TableCell>
                    <TableCell>
                      <Badge variant={key.status === 1 ? 'default' : 'destructive'}>
                        {key.status === 1 ? t('common.active') : t('common.disabled')}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {key.last_used_at ? new Date(key.last_used_at).toLocaleString() : t('keys.never')}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleToggle(key.id)}
                        >
                          {key.status === 1 ? t('common.disabled') : t('common.active')}
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => handleDelete(key.id)}
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
