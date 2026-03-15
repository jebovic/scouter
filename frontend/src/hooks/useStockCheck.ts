import { useQuery } from '@tanstack/react-query'
import { fetchStockStatus } from '../api/stockcheck'
import type { StockStatusDTO } from '../api/stockcheck'

const STALE_MS = 5 * 60 * 1000 // 5 minutes — matches backend cache TTL

export function useStockCheck(product: string, retailer: string) {
  return useQuery<StockStatusDTO>({
    queryKey: ['stock-check', product, retailer],
    queryFn: () => fetchStockStatus(product, retailer),
    enabled: product.trim().length > 0 && retailer.trim().length > 0,
    staleTime: STALE_MS,
  })
}
