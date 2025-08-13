package helper

import "strconv"

// IsValid проверяет, является ли строка с номером действительной в соответствии с алгоритмом Луна.
func IsValid(number string) bool {
	sum := 0
	parity := len(number) % 2

	for i, d := range number {
		digit, _ := strconv.Atoi(string(d))

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
