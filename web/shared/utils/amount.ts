/**
 * Amount utilities — all monetary values are stored as integer cents
 * to avoid floating-point precision issues.
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
 * Add two cent amounts safely.
 */
export function addCents(...amounts: number[]): number {
  return amounts.reduce((sum, a) => sum + a, 0)
}
