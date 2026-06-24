import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

/**
 * Format a date to YYYY-MM-DD.
 */
export function formatDate(date: string | Date | undefined | null): string {
  if (!date) return ''
  return dayjs(date).format('YYYY-MM-DD')
}

/**
 * Format a date to YYYY-MM-DD HH:mm:ss.
 */
export function formatDateTime(date: string | Date | undefined | null): string {
  if (!date) return ''
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

/**
 * Format a date as relative time (e.g., "3天前").
 */
export function formatRelativeTime(date: string | Date): string {
  return dayjs(date).fromNow()
}

/**
 * Calculate days between two dates.
 */
export function daysBetween(start: string | Date, end: string | Date): number {
  return dayjs(end).diff(dayjs(start), 'day')
}

/**
 * Check if a date is in the past.
 */
export function isPast(date: string | Date): boolean {
  return dayjs(date).isBefore(dayjs(), 'day')
}

/**
 * Get today's date as YYYY-MM-DD.
 */
export function today(): string {
  return dayjs().format('YYYY-MM-DD')
}
