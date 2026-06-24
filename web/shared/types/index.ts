/**
 * Shared TypeScript types for the travel booking platform.
 * These types mirror the backend API models.
 */

/** Unified API response envelope */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  trace_id: string
}

/** Pagination request parameters */
export interface PaginationParams {
  page?: number
  page_size?: number
}

/** Pagination metadata */
export interface PaginationMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

/** Paginated response data */
export interface PaginatedData<T> {
  list: T[]
  pagination: PaginationMeta
}

/** User profile */
export interface UserProfile {
  id: number
  phone: string
  nickname: string
  avatar_url: string
  real_name_status: 'unverified' | 'pending' | 'verified' | 'rejected'
  member_level: number
  status: string
  created_at: string
}

/** Product summary (for list views) */
export interface ProductSummary {
  id: number
  product_no: string
  product_name: string
  category_id: number
  origin_city: string
  destination_cities: string[]
  days: number
  nights: number
  cover_image: string
  summary: string
  status: string
  view_count: number
  order_count: number
  satisfaction_rate: number | null
}

/** Departure date with pricing */
export interface DepartureDate {
  id: number
  product_id: number
  departure_date: string
  return_date: string
  adult_price: number
  child_price: number
  infant_price: number
  single_supplement: number
  total_stock: number
  sold_count: number
  locked_count: number
  status: string
  /** Derived: available_stock = total_stock - sold_count - locked_count */
  available_stock?: number
}

/** Order summary (for list views) */
export interface OrderSummary {
  id: number
  order_no: string
  order_status: string
  payment_status: string
  payable_amount: number
  adult_count: number
  child_count: number
  product_name: string
  departure_date: string
  cover_image: string
  created_at: string
}

/** Frequent traveller */
export interface Traveller {
  id: number
  user_id: number
  phone: string
  birth_date: string | null
  gender: string
  is_default: boolean
  /** Note: real_name and id_card_no are encrypted and not returned in API */
}

/** Payment transaction */
export interface PaymentTransaction {
  id: number
  order_id: number
  payment_no: string
  channel: string
  method: string
  amount: number
  status: string
  paid_at: string | null
  expire_at: string
  created_at: string
}
