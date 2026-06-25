import { useQuery } from '@tanstack/vue-query'
import { computed, ref } from 'vue'
import { useApi } from './useApi'
import type { Ref } from 'vue'

/**
 * Product summary for list/card display.
 */
export interface ProductSummary {
  id: number
  product_no: string
  product_name: string
  cover_image?: string
  origin_city: string
  destination_cities: string[]
  days: number
  nights: number
  min_price: number
  product_grade?: string
  satisfaction_rate?: number
  order_count: number
  tags?: string[]
}

/**
 * Paginated product list response.
 */
export interface PaginatedProducts {
  items: ProductSummary[]
  total: number
  page: number
  page_size: number
}

/**
 * Itinerary day entry.
 */
export interface ItineraryDay {
  day_no: number
  title: string
  description?: string
  meals?: { breakfast?: boolean; lunch?: boolean; dinner?: boolean }
  hotel?: string
  transport?: string
  spots?: Array<{ name: string; description?: string; duration?: string; image?: string }>
}

/**
 * Cancellation rule.
 */
export interface CancellationRule {
  id: number
  rule_name: string
  days_before_min: number
  days_before_max?: number
  refund_percentage: number
  description?: string
}

/**
 * Departure calendar entry.
 */
export interface DepartureDate {
  id: number
  departure_date: string
  return_date: string
  adult_price: number
  child_price: number
  infant_price: number
  single_supplement: number
  available_stock: number
  stock_status: 'sufficient' | 'tight' | 'sold_out'
  cutoff_days: number
}

/**
 * Review summary.
 */
export interface ReviewSummary {
  total_count: number
  average_rating: number
  rating_distribution: Record<string, number>
}

/**
 * Product review.
 */
export interface ProductReview {
  id: number
  user_id: number
  rating: number
  content?: string
  images?: string[]
  is_anonymous: boolean
  created_at: string
}

/**
 * Full product detail.
 */
export interface ProductDetail extends ProductSummary {
  summary?: string
  description?: string
  transport_mode?: string
  min_group_size: number
  max_group_size: number
  fee_included?: string
  fee_excluded?: string
  booking_notes?: string
  itinerary?: ItineraryDay[]
  cancellation_rules?: CancellationRule[]
  images?: string[]
  review_summary?: ReviewSummary
}

/**
 * Review list response.
 */
export interface ReviewListResponse {
  items: ProductReview[]
  total: number
  summary?: ReviewSummary
}

/**
 * Product list query parameters.
 */
export interface ProductListParams {
  destination?: string
  origin?: string
  days_min?: number
  days_max?: number
  price_min?: number
  price_max?: number
  category_id?: number
  product_grade?: string
  keyword?: string
  sort?: string
  page?: number
  page_size?: number
}

/**
 * Composable for product-related data fetching with caching.
 */
export function useProduct() {
  const api = useApi()

  /**
   * Fetch product list with filters.
   */
  function useProductList(params: Ref<ProductListParams>) {
    return useQuery({
      queryKey: ['products', params],
      queryFn: () => api.get<PaginatedProducts>('/products', { params: params.value }),
      staleTime: 60 * 1000, // 1 minute
    })
  }

  /**
   * Fetch product detail by ID.
   */
  function useProductDetail(id: Ref<number | string>) {
    return useQuery({
      queryKey: ['product', id],
      queryFn: () => api.get<ProductDetail>(`/products/${id.value}`),
      enabled: computed(() => !!id.value),
      staleTime: 5 * 60 * 1000, // 5 minutes
    })
  }

  /**
   * Fetch departure calendar for a product.
   */
  function useDepartureCalendar(productId: Ref<number | string>, month: Ref<string>, months: Ref<number> = ref(3)) {
    return useQuery({
      queryKey: ['departures', productId, month, months],
      queryFn: () => api.get<DepartureDate[]>(`/products/${productId.value}/departures`, {
        params: { month: month.value, months: months.value },
      }),
      enabled: computed(() => !!productId.value && !!month.value),
      staleTime: 2 * 60 * 1000,
    })
  }

  /**
   * Fetch product reviews.
   */
  function useProductReviews(productId: Ref<number | string>, rating?: Ref<number | undefined>, page: Ref<number> = ref(1)) {
    return useQuery({
      queryKey: ['reviews', productId, rating, page],
      queryFn: () => api.get<ReviewListResponse>(`/products/${productId.value}/reviews`, {
        params: {
          rating: rating?.value,
          page: page.value,
          page_size: 20,
        },
      }),
      enabled: computed(() => !!productId.value),
      staleTime: 2 * 60 * 1000,
    })
  }

  /**
   * Fetch product itinerary.
   */
  function useProductItinerary(productId: Ref<number | string>) {
    return useQuery({
      queryKey: ['itinerary', productId],
      queryFn: () => api.get<ItineraryDay[]>(`/products/${productId.value}/itinerary`),
      enabled: computed(() => !!productId.value),
      staleTime: 10 * 60 * 1000,
    })
  }

  return {
    useProductList,
    useProductDetail,
    useDepartureCalendar,
    useProductReviews,
    useProductItinerary,
  }
}
