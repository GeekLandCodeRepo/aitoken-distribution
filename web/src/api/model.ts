import { alovaInstance } from '../lib/alova'

export interface PlazaModel {
  model_name: string
  prompt_price: number
  prompt_unit: number
  completion_price: number
  completion_unit: number
  cached_prompt_price: number
  currency: string
}

export interface PlazaChannel {
  channel_id: string
  channel_name: string
  channel_type: number
  models: PlazaModel[]
}

export const modelApi = {
  plaza: () => alovaInstance.Get<PlazaChannel[]>('/models/plaza'),
}
