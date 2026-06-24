/**
 * Validation utilities for form fields.
 */

/**
 * Validate Chinese phone number (11 digits starting with 1).
 */
export function validatePhone(phone: string): boolean {
  return /^1[3-9]\d{9}$/.test(phone)
}

/**
 * Validate Chinese ID card number (18 digits with ISO 7064:1983.MOD 11-2 checksum).
 */
export function validateIDCard(idCard: string): boolean {
  if (!/^\d{17}[\dXx]$/.test(idCard)) return false

  // Weights for each position (ISO 7064:1983.MOD 11-2)
  const weights = [7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2]
  // Check digit mapping
  const checkDigits = ['1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2']

  let sum = 0
  for (let i = 0; i < 17; i++) {
    sum += parseInt(idCard[i]) * weights[i]
  }

  const expectedCheck = checkDigits[sum % 11]
  return idCard[17].toUpperCase() === expectedCheck
}

/**
 * Validate email format.
 */
export function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

/**
 * Validate password strength (8+ chars, at least one letter and one number).
 */
export function validatePassword(password: string): boolean {
  return password.length >= 8 && /[a-zA-Z]/.test(password) && /\d/.test(password)
}
