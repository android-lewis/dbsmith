package utils

import (
	"fmt"
)

func FormatNumber(n int64) string {
	if n < 1000 && n > -1000 {
		return fmt.Sprintf("%d", n)
	}

	negative := n < 0
	if negative {
		n = -n
	}

	str := fmt.Sprintf("%d", n)
	result := ""
	count := 0

	for i := len(str) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result = "," + result
		}
		result = string(str[i]) + result
		count++
	}

	if negative {
		result = "-" + result
	}

	return result
}
