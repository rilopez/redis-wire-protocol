package resp

import (
	"fmt"
	"strings"
)

func Array(elements []interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(elements)))
	for _, item := range elements {
		switch item.(type) {
		case string:
			str := item.(string)
			sb.WriteString(BulkString(str))
		case int:
			sb.WriteString(Integer(item.(int)))
		}

	}
	return sb.String()
}

func Integer(v int) string {
	return fmt.Sprintf(":%d\r\n", v)
}

func BulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}
