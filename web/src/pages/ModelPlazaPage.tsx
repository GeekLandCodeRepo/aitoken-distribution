import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { modelApi, type PlazaChannel } from '@/api/model'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

const channelTypeNames: Record<number, string> = {
  1: 'OpenAI',
  2: 'Claude',
  3: 'Gemini',
  4: 'DeepSeek',
  5: 'Xiaomi',
}

export function ModelPlazaPage() {
  const { t } = useTranslation()
  const [channels, setChannels] = useState<PlazaChannel[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchModels()
  }, [])

  const fetchModels = async () => {
    setLoading(true)
    try {
      const data = await modelApi.plaza()
      setChannels(data || [])
    } catch (err) {
      console.error('Failed to fetch model plaza:', err)
    } finally {
      setLoading(false)
    }
  }

  const formatPrice = (price: number, unit: number, currency: string) => {
    if (price === 0) {
      return t('modelPlaza.free')
    }
    return `${currency} ${price.toFixed(4)} / ${formatTokenUnit(unit)}`
  }

  const formatTokenUnit = (unit: number) => {
    if (unit >= 1000000) {
      return t('modelPlaza.tokenUnitMillion', { count: unit / 1000000 })
    }
    if (unit >= 1000) {
      return t('modelPlaza.tokenUnitThousand', { count: unit / 1000 })
    }
    return t('modelPlaza.tokenUnit', { count: unit })
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">{t('modelPlaza.title')}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t('modelPlaza.description')}</p>
      </div>

      {loading ? (
        <div>{t('common.loading')}</div>
      ) : channels.length === 0 ? (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            {t('modelPlaza.noModels')}
          </CardContent>
        </Card>
      ) : (
        channels.map((channel) => (
          <section key={channel.channel_id} className="space-y-3">
            <div className="flex flex-wrap items-center gap-3">
              <h3 className="text-lg font-semibold">{channel.channel_name}</h3>
              <Badge variant="secondary">
                {channelTypeNames[channel.channel_type] || t('modelPlaza.unknownChannel')}
              </Badge>
              <span className="text-sm text-muted-foreground">
                {t('modelPlaza.modelCount', { count: channel.models.length })}
              </span>
            </div>

            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {channel.models.map((model) => (
                <Card key={`${channel.channel_id}-${model.model_name}`} className="overflow-hidden">
                  <CardHeader>
                    <CardTitle className="break-all font-mono text-base">{model.model_name}</CardTitle>
                    <CardDescription>{t('modelPlaza.modelCardDesc')}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3 text-sm">
                    <div className="flex items-center justify-between gap-4">
                      <span className="text-muted-foreground">{t('modelPlaza.promptPrice')}</span>
                      <span className="font-medium">
                        {formatPrice(model.prompt_price, model.prompt_unit, model.currency)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <span className="text-muted-foreground">{t('modelPlaza.completionPrice')}</span>
                      <span className="font-medium">
                        {formatPrice(model.completion_price, model.completion_unit, model.currency)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <span className="text-muted-foreground">{t('modelPlaza.cachedPromptPrice')}</span>
                      <span className="font-medium">
                        {formatPrice(model.cached_prompt_price, model.prompt_unit, model.currency)}
                      </span>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>
        ))
      )}
    </div>
  )
}
