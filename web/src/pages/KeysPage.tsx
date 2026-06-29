
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
import { Progress } from '@/components/ui/progress'
import { Switch } from '@/components/ui/switch'
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
  const [editingKey, setEditingKey] = useState<ApiKey | null>(null)
  const [editForm, setEditForm] = useState<CreateKeyRequest>({ name: '' })
  const [pendingStatusChange, setPendingStatusChange] = useState<{ key: ApiKey; enabled: boolean } | null>(null)
  const [statusUpdatingId, setStatusUpdatingId] = useState<string | null>(null)

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

  const requestStatusChange = (key: ApiKey, enabled: boolean) => {
    if ((key.status === 1) === enabled) return
    setPendingStatusChange({ key, enabled })
  }

  const confirmStatusChange = async () => {
    if (!pendingStatusChange) return

    const { key, enabled } = pendingStatusChange
    setStatusUpdatingId(key.id)
    try {
      await apiKeyApi.updateStatus(key.id, enabled)
      setKeys((prev) => prev.map((item) => (
        item.id === key.id ? { ...item, status: enabled ? 1 : 0 } : item
      )))
      setPendingStatusChange(null)
    } catch (err: any) {
      console.error('Failed to update key status:', err)
      alert(err.message || t('keys.updateFailed'))
    } finally {
      setStatusUpdatingId(null)
    }
  }

  const openEditDialog = (key: ApiKey) => {
    setEditingKey(key)
    setEditForm({
      name: key.name,
      quota_limit: key.quota_limit,
      rate_limit: key.rate_limit,
    })
  }

  const handleUpdate = async () => {
    if (!editingKey || !editForm.name) return
    try {
      const updated = await apiKeyApi.update(editingKey.id, editForm)
      setKeys((prev) => prev.map((key) => (
        key.id === editingKey.id ? { ...key, ...updated } : key
      )))
      setEditingKey(null)
      setEditForm({ name: '' })
    } catch (err: any) {
      alert(err.message || t('keys.updateFailed'))
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

  const formatUSD = (value?: number) => `$${((value || 0) / 1000000).toFixed(6)}`

  const parseUSDToUnits = (value: string) => {
    const amount = parseFloat(value)
    if (!Number.isFinite(amount) || amount < 0) return -1
    return Math.floor(amount * 1000000)
  }

  const quotaPercent = (key: ApiKey) => {
    if (key.quota_limit <= 0) return 0
    return Math.min(100, Math.max(0, (key.used_quota / key.quota_limit) * 100))
  }

  const formatLastUsed = (value: string) => {
    const date = new Date(value)
    return {
      date: date.toLocaleDateString(),
      time: date.toLocaleTimeString(),
    }
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
                    step="0.000001"
                    value={newKey.quota_limit && newKey.quota_limit > 0 ? newKey.quota_limit / 1000000 : ''}
                    onChange={(e) => setNewKey({ ...newKey, quota_limit: parseUSDToUnits(e.target.value) })}
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
                  <TableHead>{t('keys.keyInfo')}</TableHead>
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
                    <TableCell>
                      <div className="space-y-0.5">
                        <div className="font-medium">{key.name}</div>
                        <div className="font-mono text-xs text-muted-foreground">{formatKeyDisplay(key)}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="min-w-[220px] space-y-1.5">
                        <div className="flex items-center justify-between gap-2 text-xs">
                          <span className="font-medium">
                            {formatUSD(key.used_quota)} / {key.quota_limit === -1 ? t('common.unlimited') : formatUSD(key.quota_limit)}
                          </span>
                          <span className="text-muted-foreground">
                            {quotaPercent(key).toFixed(0)}%
                          </span>
                        </div>
                        <Progress value={quotaPercent(key)} className="h-1.5" />
                      </div>
                    </TableCell>
                    <TableCell>
                      {key.rate_limit === -1 ? t('common.unlimited') : `${key.rate_limit}/min`}
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={key.status === 1}
                        disabled={statusUpdatingId === key.id}
                        onCheckedChange={(checked) => requestStatusChange(key, checked)}
                      />
                    </TableCell>
                    <TableCell>
                      {key.last_used_at ? (
                        <div className="space-y-0.5 text-sm leading-tight">
                          <div>{formatLastUsed(key.last_used_at).date}</div>
                          <div className="text-muted-foreground">{formatLastUsed(key.last_used_at).time}</div>
                        </div>
                      ) : (
                        t('keys.never')
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => openEditDialog(key)}
                        >
                          {t('common.edit')}
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

      <Dialog open={!!editingKey} onOpenChange={(open) => !open && setEditingKey(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('keys.editTitle')}</DialogTitle>
            <DialogDescription>{t('keys.editDesc')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t('common.name')}</Label>
              <Input
                value={editForm.name}
                onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('keys.quotaLimit')}</Label>
              <Input
                type="number"
                step="0.000001"
                placeholder="-1"
                value={editForm.quota_limit && editForm.quota_limit > 0 ? editForm.quota_limit / 1000000 : ''}
                onChange={(e) => setEditForm({ ...editForm, quota_limit: parseUSDToUnits(e.target.value) })}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('keys.rateLimit')}</Label>
              <Input
                type="number"
                placeholder="-1"
                value={editForm.rate_limit && editForm.rate_limit > 0 ? editForm.rate_limit : ''}
                onChange={(e) => setEditForm({ ...editForm, rate_limit: parseInt(e.target.value) || -1 })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditingKey(null)}>{t('common.cancel')}</Button>
            <Button onClick={handleUpdate}>{t('common.save')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!pendingStatusChange} onOpenChange={(open) => !open && setPendingStatusChange(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {pendingStatusChange?.enabled ? t('keys.enableConfirmTitle') : t('keys.disableConfirmTitle')}
            </DialogTitle>
            <DialogDescription>
              {pendingStatusChange?.enabled
                ? t('keys.enableConfirmDesc', { name: pendingStatusChange.key.name })
                : t('keys.disableConfirmDesc', { name: pendingStatusChange?.key.name })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPendingStatusChange(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant={pendingStatusChange?.enabled ? 'default' : 'destructive'}
              onClick={confirmStatusChange}
              disabled={!!statusUpdatingId}
            >
              {pendingStatusChange?.enabled ? t('common.active') : t('common.disabled')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
