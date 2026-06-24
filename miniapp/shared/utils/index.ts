/**
 * Shared utilities for mini-program.
 * Business logic shared with web platform.
 */

/**
 * Format cents to display string (e.g., 19999 → "199.99").
 */
export function formatAmount(cents: number | undefined | null): string {
  if (cents === undefined || cents === null) return '0.00'
  return (cents / 100).toFixed(2)
}

/**
 * Format cents to display with yuan symbol (e.g., 19999 → "¥199.99").
 */
export function formatPrice(cents: number | undefined | null): string {
  if (cents === undefined || cents === null) return '¥0.00'
  return `¥${(cents / 100).toFixed(2)}`
}

/**
 * Parse a yuan display string to cents (e.g., "199.99" → 19999).
 */
export function parseAmount(yuan: string): number {
  const num = parseFloat(yuan)
  if (isNaN(num)) return 0
  return Math.round(num * 100)
}

/**
 * Format a date string to YYYY-MM-DD.
 */
export function formatDate(date: string | Date | undefined | null): string {
  if (!date) return ''
  const d = typeof date === 'string' ? new Date(date) : date
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Format a date string to YYYY-MM-DD HH:mm:ss.
 */
export function formatDateTime(date: string | Date | undefined | null): string {
  if (!date) return ''
  const d = typeof date === 'string' ? new Date(date) : date
  const dateStr = formatDate(d)
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  return `${dateStr} ${hours}:${minutes}:${seconds}`
}

/**
 * Validate Chinese phone number.
 */
export function validatePhone(phone: string): boolean {
  return /^1[3-9]\d{9}$/.test(phone)
}

/**
 * Validate Chinese ID card number (18 digits with ISO 7064:1983.MOD 11-2 checksum).
 */
export function validateIDCard(idCard: string): boolean {
  if (!/^\d{17}[\dXx]$/.test(idCard)) return false

  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  const checkDigits = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']

  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(idCard[i]) * weights[i]
  }

  const expectedCheck = checkDigits[sum % 11]
  return idCard[17].toUpperCase() === expectedCheck
}

/**
 * Mask phone number for display (e.g., "138****8000").
 */
export function maskPhone(phone: string): string {
  if (phone.length < 7) return phone
  return phone.slice(0, 3) + '****' + phone.slice(-4)
}

/**
 * Mask ID card for display (e.g., "110101********1234").
 */
export function maskIDCard(idCard: string): string {
  if (idCard.length < 10) return idCard
  return idCard.slice(0, 6) + '*'.repeat(idCard.length - 10) + idCard.slice(-4)
}
