
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
import { adminModelApi, type ManagedModel, type CreateManagedModelRequest } from '@/api/model'
import { channelApi, type Channel } from '@/api/channel'

const UNIT_OPTIONS = [
  { value: '1000', label: '1K' },
  { value: '1000000', label: '1M' },
  { value: '10000000', label: '10M' },
  { value: '100000000', label: '100M' },
  { value: '200000000', label: '200M' },
]

export function ModelManagementPage() {
  const { t } = useTranslation()
  const [models, setModels] = useState<ManagedModel[]>([])
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingModelId, setEditingModelId] = useState<string | null>(null)
  const [updatingModelIds, setUpdatingModelIds] = useState<Set<string>>(new Set())
  const [newModel, setNewModel] = useState<CreateManagedModelRequest>({
    channel_id: '',
    model_name: '',
    prompt_price: 0,
    prompt_unit: 1000000,
    completion_price: 0,
    completion_unit: 1000000,
    cached_prompt_price: 0,
    currency: 'USD',
    enabled: true,
  })

  useEffect(() => {
    fetchModels()
    fetchChannels()
  }, [])

  const fetchModels = async () => {
    setLoading(true)
    try {
      const data = await adminModelApi.list()
      setModels(data || [])
    } catch (err) {
      console.error('Failed to fetch models:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchChannels = async () => {
    try {
      const data = await channelApi.list({ page: 1, size: 100 })
      setChannels(data.items || [])
    } catch (err) {
      console.error('Failed to fetch channels:', err)
    }
  }

  const getChannelName = (channelId: string) => {
    return channels.find(c => c.id === channelId)?.name || channelId
  }

  const resetForm = () => {
    setEditingModelId(null)
    setNewModel({
      channel_id: '',
      model_name: '',
      prompt_price: 0,
      prompt_unit: 1000000,
      completion_price: 0,
      completion_unit: 1000000,
      cached_prompt_price: 0,
      currency: 'USD',
      enabled: true,
    })
  }

  const handleSubmit = async () => {
    if (!newModel.model_name || !newModel.channel_id) return

    try {
      if (editingModelId) {
        const updated = await adminModelApi.update(editingModelId, newModel)
        setModels((items) => items.map((item) => item.id === editingModelId ? updated : item))
      } else {
        const created = await adminModelApi.create(newModel)
        setModels((items) => [created, ...items])
      }
      setDialogOpen(false)
      resetForm()
      await fetchModels()
    } catch (err: any) {
      alert(err.message || t('modelManagement.addFailed'))
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(t('modelManagement.deleteConfirm'))) return

    try {
      await adminModelApi.delete(id)
      setModels((items) => items.filter((item) => item.id !== id))
      await fetchModels()
    } catch (err: any) {
      alert(err.message || t('modelManagement.deleteFailed'))
    }
  }

  const handleToggleEnabled = async (model: ManagedModel, enabled: boolean) => {
    if (updatingModelIds.has(model.id)) return

    setUpdatingModelIds((ids) => new Set(ids).add(model.id))
    setModels((items) => items.map((item) => item.id === model.id ? { ...item, enabled } : item))

    try {
      const updated = await adminModelApi.toggle(model.id, enabled)
      setModels((items) => items.map((item) => item.id === model.id ? updated : item))
    } catch (err: any) {
      setModels((items) => items.map((item) => item.id === model.id ? { ...item, enabled: model.enabled } : item))
      alert(err.message || t('modelManagement.updateFailed'))
    } finally {
      setUpdatingModelIds((ids) => {
        const next = new Set(ids)
        next.delete(model.id)
        return next
      })
    }
  }

  const handleEdit = (model: ManagedModel) => {
    setEditingModelId(model.id)
    setNewModel({
      channel_id: model.channel_id,
      model_name: model.model_name,
      prompt_price: model.prompt_price,
      prompt_unit: model.prompt_unit,
      completion_price: model.completion_price,
      completion_unit: model.completion_unit,
      image_price: model.image_price ?? undefined,
      audio_price: model.audio_price ?? undefined,
      cached_prompt_price: model.cached_prompt_price,
      currency: model.currency,
      enabled: model.enabled,
    })
    setDialogOpen(true)
  }

  const formatUnit = (unit: number) => {
    if (unit >= 1000000) return `${unit / 1000000}M`
    if (unit >= 1000) return `${unit / 1000}K`
    return String(unit)
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">{t('modelManagement.title')}</h2>
        <div className="flex gap-2">
          <Dialog open={dialogOpen} onOpenChange={(open) => {
            setDialogOpen(open)
            if (!open) resetForm()
          }}>
            <DialogTrigger asChild>
              <Button onClick={resetForm}>{t('modelManagement.addModel')}</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{editingModelId ? t('modelManagement.editTitle') : t('modelManagement.addTitle')}</DialogTitle>
                <DialogDescription>{editingModelId ? t('modelManagement.editDesc') : t('modelManagement.addDesc')}</DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t('modelManagement.channel')}</Label>
                  <Select
                    value={newModel.channel_id}
                    onValueChange={(v: string) => setNewModel({ ...newModel, channel_id: v })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t('modelManagement.channelPlaceholder')} />
                    </SelectTrigger>
                    <SelectContent>
                      {channels.map((ch) => (
                        <SelectItem key={ch.id} value={ch.id}>
                          {ch.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>{t('modelManagement.modelName')}</Label>
                  <Input
                    placeholder="gpt-4o"
                    value={newModel.model_name}
                    onChange={(e) => setNewModel({ ...newModel, model_name: e.target.value })}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>{t('modelManagement.promptPrice')}</Label>
                    <Input
                      type="number"
                      step="0.000001"
                      value={newModel.prompt_price}
                      onChange={(e) => setNewModel({ ...newModel, prompt_price: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('modelManagement.promptUnit')}</Label>
                    <Select
                      value={String(newModel.prompt_unit)}
                      onValueChange={(v: string) => setNewModel({ ...newModel, prompt_unit: parseInt(v) })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {UNIT_OPTIONS.map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>{t('modelManagement.completionPrice')}</Label>
                    <Input
                      type="number"
                      step="0.000001"
                      value={newModel.completion_price}
                      onChange={(e) => setNewModel({ ...newModel, completion_price: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('modelManagement.completionUnit')}</Label>
                    <Select
                      value={String(newModel.completion_unit)}
                      onValueChange={(v: string) => setNewModel({ ...newModel, completion_unit: parseInt(v) })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {UNIT_OPTIONS.map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>{t('modelManagement.cachedPromptPrice')}</Label>
                    <Input
                      type="number"
                      step="0.000001"
                      value={newModel.cached_prompt_price}
                      onChange={(e) => setNewModel({ ...newModel, cached_prompt_price: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                </div>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div className="space-y-0.5">
                    <Label>{t('modelManagement.enabled')}</Label>
                    <p className="text-sm text-muted-foreground">{t('modelManagement.enabledDesc')}</p>
                  </div>
                  <Switch
                    checked={newModel.enabled ?? false}
                    onCheckedChange={(checked) => setNewModel({ ...newModel, enabled: checked })}
                  />
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
                  <Button onClick={handleSubmit}>{editingModelId ? t('common.save') : t('common.create')}</Button>
                </DialogFooter>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('modelManagement.title')}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('modelManagement.channel')}</TableHead>
                  <TableHead>{t('modelManagement.modelName')}</TableHead>
                  <TableHead>{t('modelManagement.promptPrice')}</TableHead>
                  <TableHead>{t('modelManagement.completionPrice')}</TableHead>
                  <TableHead>{t('modelManagement.cachedPromptPrice')}</TableHead>
                  <TableHead>{t('modelManagement.enabled')}</TableHead>
                  <TableHead>{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {models.map((model) => (
                  <TableRow key={model.id}>
                    <TableCell>
                      <Badge variant="outline">
                        {getChannelName(model.channel_id)}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono">{model.model_name}</TableCell>
                    <TableCell>${model.prompt_price}/{formatUnit(model.prompt_unit)}</TableCell>
                    <TableCell>${model.completion_price}/{formatUnit(model.completion_unit)}</TableCell>
                    <TableCell>${model.cached_prompt_price}/{formatUnit(model.prompt_unit)}</TableCell>
                    <TableCell>
                      <Switch
                        checked={model.enabled}
                        disabled={updatingModelIds.has(model.id)}
                        onCheckedChange={(checked) => handleToggleEnabled(model, checked)}
                        aria-label={t('modelManagement.enabled')}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button size="sm" variant="outline" onClick={() => handleEdit(model)}>
                          {t('common.edit')}
                        </Button>
                        <Button size="sm" variant="destructive" onClick={() => handleDelete(model.id)}>
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
