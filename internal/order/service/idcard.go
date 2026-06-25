package service

import (
	"time"
)

// ValidateIDCard validates an 18-digit Chinese ID card number using ISO 7064:1983.MOD 11-2.
func ValidateIDCard(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}

	// Check that all characters are digits (except possibly the last one which can be 'X')
	for i := 0; i < 17; i++ {
		if idCard[i] < '0' || idCard[i] > '9' {
			return false
		}
	}
	last := idCard[17]
	if last != 'X' && last != 'x' && (last < '0' || last > '9') {
		return false
	}

	// Weights for each position
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// Check codes corresponding to remainder 0-10
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(idCard[i]-'0') * weights[i]
	}
	remainder := sum % 11
	expectedCheck := checkCodes[remainder]

	// Compare (case-insensitive for X)
	actualCheck := idCard[17]
	if actualCheck >= 'a' && actualCheck <= 'z' {
		actualCheck = actualCheck - 32 // to uppercase
	}

	return actualCheck == expectedCheck
}

// ParseBirthDateFromIDCard extracts the birth date from an 18-digit ID card number.
func ParseBirthDateFromIDCard(idCard string) (time.Time, error) {
	if len(idCard) != 18 {
		return time.Time{}, ErrInvalidIDCard
	}
	return time.Parse("20060102", idCard[6:14])
}

// IsChildByBirthDate checks if a person is a child (2-12 years old, not including 12).
func IsChildByBirthDate(birthDate time.Time, referenceDate time.Time) bool {
	age := referenceDate.Year() - birthDate.Year()
	if referenceDate.YearDay() < birthDate.YearDay() {
		age--
	}
	return age >= 2 && age < 12
}

// IsInfantByBirthDate checks if a person is an infant (<2 years old).
func IsInfantByBirthDate(birthDate time.Time, referenceDate time.Time) bool {
	age := referenceDate.Year() - birthDate.Year()
	if referenceDate.YearDay() < birthDate.YearDay() {
		age--
	}
	return age < 2
}
