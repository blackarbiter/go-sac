package validator

import (
	"regexp"
)

var (
	// 匹配常见SQL注入模式，忽略大小写
	sqlInjectPattern = regexp.MustCompile(`(?i)(\b(union\b.*\bselect|select\b.*\bfrom|insert\b.*\binto|delete\b.*\bfrom|update\b.*\bset|drop\b|truncate\b|exec\b)\b|'\s*or\b|;\s*--|/\*.*\*/)`)
)

// CheckSQLInjection 检测输入是否包含潜在SQL注入攻击
func CheckSQLInjection(input string) bool {
	return sqlInjectPattern.MatchString(input)
}
