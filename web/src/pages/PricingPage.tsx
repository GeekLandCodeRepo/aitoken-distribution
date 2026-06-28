
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
import { pricingApi, type Pricing, type CreatePricingRequest } from '@/api/pricing'
import { channelApi, type Channel } from '@/api/channel'

const UNIT_OPTIONS = [
  { value: '1000', label: '1K' },
  { value: '1000000', label: '1M' },
  { value: '10000000', label: '10M' },
  { value: '100000000', label: '100M' },
  { value: '200000000', label: '200M' },
]

export function PricingPage() {
  const { t } = useTranslation()
  const [pricings, setPricings] = useState<Pricing[]>([])
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingPricingId, setEditingPricingId] = useState<string | null>(null)
  const [updatingPricingIds, setUpdatingPricingIds] = useState<Set<string>>(new Set())
  const [newPricing, setNewPricing] = useState<CreatePricingRequest>({
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
    fetchPricings()
    fetchChannels()
  }, [])

  const fetchPricings = async () => {
    setLoading(true)
    try {
      const data = await pricingApi.list()
      setPricings(data || [])
    } catch (err) {
      console.error('Failed to fetch pricings:', err)
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
    setEditingPricingId(null)
    setNewPricing({
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
    if (!newPricing.model_name || !newPricing.channel_id) return

    try {
      if (editingPricingId) {
        const updated = await pricingApi.update(editingPricingId, newPricing)
        setPricings((items) => items.map((item) => item.id === editingPricingId ? updated : item))
      } else {
        const created = await pricingApi.create(newPricing)
        setPricings((items) => [created, ...items])
      }
      setDialogOpen(false)
      resetForm()
      await fetchPricings()
    } catch (err: any) {
      alert(err.message || t('pricing.addFailed'))
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm(t('pricing.deleteConfirm'))) return

    try {
      await pricingApi.delete(id)
      setPricings((items) => items.filter((item) => item.id !== id))
      await fetchPricings()
    } catch (err: any) {
      alert(err.message || t('pricing.deleteFailed'))
    }
  }

  const handleToggleEnabled = async (pricing: Pricing, enabled: boolean) => {
    if (updatingPricingIds.has(pricing.id)) return

    setUpdatingPricingIds((ids) => new Set(ids).add(pricing.id))
    setPricings((items) => items.map((item) => item.id === pricing.id ? { ...item, enabled } : item))

    try {
      const updated = await pricingApi.toggle(pricing.id, enabled)
      setPricings((items) => items.map((item) => item.id === pricing.id ? updated : item))
    } catch (err: any) {
      setPricings((items) => items.map((item) => item.id === pricing.id ? { ...item, enabled: pricing.enabled } : item))
      alert(err.message || t('pricing.updateFailed'))
    } finally {
      setUpdatingPricingIds((ids) => {
        const next = new Set(ids)
        next.delete(pricing.id)
        return next
      })
    }
  }

  const handleSync = async () => {
    try {
      const result = await pricingApi.sync()
      alert(t('pricing.syncResult', { created: result.created.length, skipped: result.skipped.length }))
      await fetchPricings()
    } catch (err: any) {
      alert(err.message || t('pricing.syncFailed'))
    }
  }

  const handleEdit = (pricing: Pricing) => {
    setEditingPricingId(pricing.id)
    setNewPricing({
      channel_id: pricing.channel_id,
      model_name: pricing.model_name,
      prompt_price: pricing.prompt_price,
      prompt_unit: pricing.prompt_unit,
      completion_price: pricing.completion_price,
      completion_unit: pricing.completion_unit,
      image_price: pricing.image_price ?? undefined,
      audio_price: pricing.audio_price ?? undefined,
      cached_prompt_price: pricing.cached_prompt_price,
      currency: pricing.currency,
      enabled: pricing.enabled,
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
        <h2 className="text-2xl font-bold">{t('pricing.title')}</h2>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleSync}>{t('pricing.syncModels')}</Button>
          <Dialog open={dialogOpen} onOpenChange={(open) => {
            setDialogOpen(open)
            if (!open) resetForm()
          }}>
            <DialogTrigger asChild>
              <Button onClick={resetForm}>{t('pricing.addPricing')}</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{editingPricingId ? t('pricing.editTitle') : t('pricing.addTitle')}</DialogTitle>
                <DialogDescription>{editingPricingId ? t('pricing.editDesc') : t('pricing.addDesc')}</DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t('pricing.channel')}</Label>
                  <Select
                    value={newPricing.channel_id}
                    onValueChange={(v: string) => setNewPricing({ ...newPricing, channel_id: v })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t('pricing.channelPlaceholder')} />
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
                  <Label>{t('pricing.modelName')}</Label>
                  <Input
                    placeholder="gpt-4o"
                    value={newPricing.model_name}
                    onChange={(e) => setNewPricing({ ...newPricing, model_name: e.target.value })}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>{t('pricing.promptPrice')}</Label>
                    <Input
                      type="number"
                      step="0.000001"
                      value={newPricing.prompt_price}
                      onChange={(e) => setNewPricing({ ...newPricing, prompt_price: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('pricing.promptUnit')}</Label>
                    <Select
                      value={String(newPricing.prompt_unit)}
                      onValueChange={(v: string) => setNewPricing({ ...newPricing, prompt_unit: parseInt(v) })}
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
                    <Label>{t('pricing.completionPrice')}</Label>
                    <Input
                      type="number"
                      step="0.000001"
                      value={newPricing.completion_price}
                      onChange={(e) => setNewPricing({ ...newPricing, completion_price: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('pricing.completionUnit')}</Label>
                    <Select
                      value={String(newPricing.completion_unit)}
                      onValueChange={(v: string) => setNewPricing({ ...newPricing, completion_unit: parseInt(v) })}
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
                    <Label>{t('pricing.cachedPromptPrice')}</Label>
                    <Input
                      type="number"
                      step="0.000001"
                      value={newPricing.cached_prompt_price}
                      onChange={(e) => setNewPricing({ ...newPricing, cached_prompt_price: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                </div>
                <div className="flex items-center justify-between rounded-lg border p-3">
                  <div className="space-y-0.5">
                    <Label>{t('pricing.enabled')}</Label>
                    <p className="text-sm text-muted-foreground">{t('pricing.enabledDesc')}</p>
                  </div>
                  <Switch
                    checked={newPricing.enabled ?? false}
                    onCheckedChange={(checked) => setNewPricing({ ...newPricing, enabled: checked })}
                  />
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
                  <Button onClick={handleSubmit}>{editingPricingId ? t('common.save') : t('common.create')}</Button>
                </DialogFooter>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('pricing.title')}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div>{t('common.loading')}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('pricing.channel')}</TableHead>
                  <TableHead>{t('pricing.modelName')}</TableHead>
                  <TableHead>{t('pricing.promptPrice')}</TableHead>
                  <TableHead>{t('pricing.completionPrice')}</TableHead>
                  <TableHead>{t('pricing.cachedPromptPrice')}</TableHead>
                  <TableHead>{t('pricing.enabled')}</TableHead>
                  <TableHead>{t('common.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pricings.map((pricing) => (
                  <TableRow key={pricing.id}>
                    <TableCell>
                      <Badge variant="outline">
                        {getChannelName(pricing.channel_id)}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono">{pricing.model_name}</TableCell>
                    <TableCell>${pricing.prompt_price}/{formatUnit(pricing.prompt_unit)}</TableCell>
                    <TableCell>${pricing.completion_price}/{formatUnit(pricing.completion_unit)}</TableCell>
                    <TableCell>${pricing.cached_prompt_price}/{formatUnit(pricing.prompt_unit)}</TableCell>
                    <TableCell>
                      <Switch
                        checked={pricing.enabled}
                        disabled={updatingPricingIds.has(pricing.id)}
                        onCheckedChange={(checked) => handleToggleEnabled(pricing, checked)}
                        aria-label={t('pricing.enabled')}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button size="sm" variant="outline" onClick={() => handleEdit(pricing)}>
                          {t('common.edit')}
                        </Button>
                        <Button size="sm" variant="destructive" onClick={() => handleDelete(pricing.id)}>
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
