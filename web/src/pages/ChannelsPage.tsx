
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { channelApi, type Channel, type CreateChannelRequest } from '@/api/channel'

const CHANNEL_TYPES = [
  { value: '1', label: 'OpenAI' },
  { value: '2', label: 'Claude' },
  { value: '3', label: 'Gemini' },
  { value: '4', label: 'DeepSeek' },
]

const CHANNEL_TYPE_NAMES = CHANNEL_TYPES.reduce<Record<number, string>>((acc, item) => {
  acc[Number(item.value)] = item.label
  return acc
}, {})

function normalizeModels(value: unknown): string[] {
  if (Array.isArray(value)) return value
  if (typeof value !== 'string' || value === '') return []
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return value.split(',').map(m => m.trim()).filter(Boolean)
  }
}

function formatSuccessRate(channel: Channel) {
  const requestCount = channel.request_count || 0
  const successCount = channel.success_count || 0
  const rate = Number.isFinite(channel.success_rate)
    ? channel.success_rate
    : requestCount > 0
      ? successCount / requestCount
      : 0

  return `${(rate * 100).toFixed(1)}%`
}

export function ChannelsPage() {
  const { t } = useTranslation()
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingChannelId, setEditingChannelId] = useState<string | null>(null)
  const [pendingStatusChange, setPendingStatusChange] = useState<{ channel: Channel; enabled: boolean } | null>(null)
  const [pendingDeleteChannel, setPendingDeleteChannel] = useState<Channel | null>(null)
  const [newChannel, setNewChannel] = useState<CreateChannelRequest>({
    name: '',
    type: 1,
    base_url: '',
    api_key: '',
    models: [''],
  })

  useEffect(() => {
    fetchChannels()
  }, [])

  const fetchChannels = async () => {
    setLoading(true)
    try {
      const data = await channelApi.list()
      setChannels(data.items || [])
    } catch (err) {
      console.error('Failed to fetch channels:', err)
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setEditingChannelId(null)
    setNewChannel({ name: '', type: 1, base_url: '', api_key: '', models: [''] })
  }

  const handleSubmit = async () => {
    const models = newChannel.models.map(m => m.trim()).filter(Boolean)
    if (!newChannel.name || !newChannel.base_url || (!editingChannelId && !newChannel.api_key) || models.length === 0) return

    try {
      if (editingChannelId) {
        const updated = await channelApi.update(editingChannelId, { ...newChannel, models })
        setChannels((items) => items.map((item) => item.id === editingChannelId ? updated : item))
      } else {
        const created = await channelApi.create({ ...newChannel, models })
        setChannels((items) => [created, ...items])
      }
      setDialogOpen(false)
      resetForm()
    } catch (err: any) {
      alert(err.message || t('channels.addFailed'))
    }
  }

  const confirmDelete = async () => {
    if (!pendingDeleteChannel) return
    try {
      await channelApi.delete(pendingDeleteChannel.id)
      setChannels((items) => items.filter((item) => item.id !== pendingDeleteChannel.id))
      setPendingDeleteChannel(null)
    } catch (err: any) {
      alert(err.message || t('channels.deleteFailed'))
    }
  }

  const handleTest = async (id: string) => {
    try {
      const result = await channelApi.test(id)
      alert(result.success ? `${t('channels.testSuccess')}${result.latency_ms}ms` : t('channels.testFailed'))
    } catch (err: any) {
      alert(err.message || t('channels.testFailed'))
    }
  }

  const confirmStatusChange = async () => {
    if (!pendingStatusChange) return
    const { channel, enabled } = pendingStatusChange
    try {
      await channelApi.updateStatus(channel.id, enabled)
      setChannels((items) => items.map((item) => (
        item.id === channel.id ? { ...item, status: enabled ? 1 : 0 } : item
      )))
    } catch (err) {
      console.error('Failed to update channel status:', err)
    } finally {
      setPendingStatusChange(null)
    }
  }

  const handleEdit = (channel: Channel) => {
    setEditingChannelId(channel.id)
    setNewChannel({
      name: channel.name,
      type: channel.type,
      base_url: channel.base_url,
      api_key: '',
      models: normalizeModels(channel.models),
      priority: channel.priority,
      weight: channel.weight,
    })
    setDialogOpen(true)
  }

  const updateModel = (index: number, value: string) => {
    setNewChannel({
      ...newChannel,
      models: newChannel.models.map((model, i) => i === index ? value : model),
    })
  }

  const addModelInput = () => {
    setNewChannel({ ...newChannel, models: [...newChannel.models, ''] })
  }

  const removeModelInput = (index: number) => {
    const models = newChannel.models.filter((_, i) => i !== index)
    setNewChannel({ ...newChannel, models: models.length > 0 ? models : [''] })
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('channels.title')}</h2>
        <Dialog open={dialogOpen} onOpenChange={(open) => {
          setDialogOpen(open)
          if (!open) resetForm()
        }}>
          <DialogTrigger asChild>
            <Button onClick={resetForm}>{t('channels.addChannel')}</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{editingChannelId ? t('channels.editTitle') : t('channels.addTitle')}</DialogTitle>
              <DialogDescription>{editingChannelId ? t('channels.editDesc') : t('channels.addDesc')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>{t('channels.channelName')}</Label>
                <Input
                  placeholder={t('channels.channelNamePlaceholder')}
                  value={newChannel.name}
                  onChange={(e) => setNewChannel({ ...newChannel, name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('common.type')}</Label>
                <Select
                  value={String(newChannel.type)}
                  onValueChange={(v: string) => setNewChannel({ ...newChannel, type: parseInt(v) })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {CHANNEL_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>{t('channels.baseUrl')}</Label>
                <Input
                  placeholder={t('channels.baseUrlPlaceholder')}
                  value={newChannel.base_url}
                  onChange={(e) => setNewChannel({ ...newChannel, base_url: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('channels.apiKey')}</Label>
                <Input
                  type="password"
                  placeholder={editingChannelId ? t('channels.apiKeyEditPlaceholder') : 'sk-...'}
                  value={newChannel.api_key}
                  onChange={(e) => setNewChannel({ ...newChannel, api_key: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label>{t('channels.models')}</Label>
                <div className="space-y-2">
                  {newChannel.models.map((model, index) => (
                    <div key={index} className="flex gap-2">
                      <Input
                        placeholder={t('channels.modelPlaceholder')}
                        value={model}
                        onChange={(e) => updateModel(index, e.target.value)}
                      />
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => removeModelInput(index)}
                        disabled={newChannel.models.length === 1}
                      >
                        {t('common.delete')}
                      </Button>
                    </div>
                  ))}
                  <Button type="button" variant="outline" onClick={addModelInput}>
                    {t('channels.addModel')}
                  </Button>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
                <Button onClick={handleSubmit}>{editingChannelId ? t('common.save') : t('common.create')}</Button>
              </DialogFooter>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('channels.title')}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('common.name')}</TableHead>
                  <TableHead>{t('common.type')}</TableHead>
                  <TableHead>{t('channels.models')}</TableHead>
                  <TableHead>{t('channels.priority')}</TableHead>
                  <TableHead>{t('channels.balance')}</TableHead>
                  <TableHead>{t('channels.successRate')}</TableHead>
                  <TableHead>{t('common.status')}</TableHead>
                  <TableHead>{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {channels.map((channel) => {
                  const models = normalizeModels(channel.models)

                  return (
                  <TableRow key={channel.id}>
                    <TableCell className="font-medium">{channel.name}</TableCell>
                    <TableCell>{channel.type_name || CHANNEL_TYPE_NAMES[channel.type] || channel.type}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {models.slice(0, 3).map((model) => (
                          <Badge key={model} variant="outline" className="text-xs">
                            {model}
                          </Badge>
                        ))}
                        {models.length > 3 && (
                          <Badge variant="outline" className="text-xs">
                            +{models.length - 3}
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{channel.priority}</TableCell>
                    <TableCell>${channel.balance?.toFixed(2) || '0.00'}</TableCell>
                    <TableCell>{formatSuccessRate(channel)}</TableCell>
                    <TableCell>
                      <Switch
                        checked={channel.status === 1}
                        onCheckedChange={(checked) => setPendingStatusChange({ channel, enabled: checked })}
                        aria-label={channel.status === 1 ? t('common.active') : t('common.disabled')}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button size="sm" variant="outline" onClick={() => handleTest(channel.id)}>
                          {t('channels.test')}
                        </Button>
                        <Button size="sm" variant="outline" onClick={() => handleEdit(channel)}>
                          {t('common.edit')}
                        </Button>
                        <Button size="sm" variant="destructive" onClick={() => setPendingDeleteChannel(channel)}>
                          {t('common.delete')}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!pendingStatusChange} onOpenChange={(open) => !open && setPendingStatusChange(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {pendingStatusChange?.enabled ? t('channels.enableTitle') : t('channels.disableTitle')}
            </DialogTitle>
            <DialogDescription>
              {pendingStatusChange?.enabled ? t('channels.enableConfirm') : t('channels.disableConfirm')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPendingStatusChange(null)}>{t('common.cancel')}</Button>
            <Button onClick={confirmStatusChange}>{t('common.confirm')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!pendingDeleteChannel} onOpenChange={(open) => !open && setPendingDeleteChannel(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('channels.deleteTitle')}</DialogTitle>
            <DialogDescription>
              {t('channels.deleteConfirm', { name: pendingDeleteChannel?.name })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPendingDeleteChannel(null)}>{t('common.cancel')}</Button>
            <Button variant="destructive" onClick={confirmDelete}>{t('common.delete')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
