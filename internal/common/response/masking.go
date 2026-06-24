package response

import (
	"strings"
)

// MaskPhone masks a phone number showing first 3 and last 4 digits.
// Example: "13800138000" -> "138****8000"
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + strings.Repeat("*", len(phone)-7) + phone[len(phone)-4:]
}

// MaskIDCard masks an ID card number showing first 6 and last 4 digits.
// Example: "110101199001011234" -> "110101********1234"
func MaskIDCard(idCard string) string {
	if len(idCard) < 10 {
		return idCard
	}
	return idCard[:6] + strings.Repeat("*", len(idCard)-10) + idCard[len(idCard)-4:]
}

// MaskName masks a name showing only the surname (first character).
// Example: "张三" -> "张*", "欧阳明" -> "欧**"
func MaskName(name string) string {
	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-1)
}

// MaskBankCard masks a bank card number showing first 6 and last 4 digits.
// Example: "6225880112345678" -> "622588******5678"
func MaskBankCard(cardNo string) string {
	if len(cardNo) < 10 {
		return cardNo
	}
	return cardNo[:6] + strings.Repeat("*", len(cardNo)-10) + cardNo[len(cardNo)-4:]
}

// MaskEmail masks an email address showing first 3 characters of local part.
// Example: "user@example.com" -> "use***@example.com"
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) < 3 {
		return email
	}
	return parts[0][:3] + "***@" + parts[1]
}
