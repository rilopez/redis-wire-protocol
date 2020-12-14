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
			sb.WriteString(BulkString(&str))
		case int:
			sb.WriteString(Integer(item.(int)))
		}

	}
	return sb.String()
}

func Integer(v int) string {
	return fmt.Sprintf(":%d\r\n", v)
}

func BulkString(str *string) string {
	if str == nil {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(*str), *str)
}

func SimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func Error(err error) string {
	return fmt.Sprintf("-%s\r\n", err)
}
