import request from './request'

export interface Banner {
  id: number
  image_url: string
  title: string
  link: string
  sort_order: number
  status: string
  position?: string
  start_date?: string
  end_date?: string
}

export interface Destination {
  id: number
  name: string
  image_url?: string
  product_count: number
  min_price: number
  sort_order: number
}

// Banner CRUD
export function listBanners(params?: { position?: string; status?: string }) {
  return request.get<{ items: Banner[]; total: number }>('/admin/banners', { params })
}

export function createBanner(data: Partial<Banner>) {
  return request.post<{ id: number }>('/admin/banners', data)
}

export function updateBanner(id: number, data: Partial<Banner>) {
  return request.put(`/admin/banners/${id}`, data)
}

export function deleteBanner(id: number) {
  return request.delete(`/admin/banners/${id}`)
}

// Destinations
export function listDestinations(params?: { category?: string }) {
  return request.get<{ items: Destination[] }>('/destinations/popular', { params })
}
