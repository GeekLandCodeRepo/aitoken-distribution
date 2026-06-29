import { alovaInstance } from '@/lib/alova'

export interface BalanceTransaction {
  id: string
  user_id: string
  username: string
  email: string
  type: number
  amount: number
  balance_after: number
  reference_type: string
  reference_id: string
  description: string
  created_at: string
}

export interface BalanceTransactionListResponse {
  total: number
  page: number
  size: number
  items: BalanceTransaction[]
}

export const billingApi = {
  transactions: (params?: { page?: number; size?: number; type?: number }) =>
    alovaInstance.Get<BalanceTransactionListResponse>('/billing/transactions', { params }),

  adminTransactions: (params?: { page?: number; size?: number; user_email?: string; type?: number }) =>
    alovaInstance.Get<BalanceTransactionListResponse>('/admin/transactions', { params }),
}
