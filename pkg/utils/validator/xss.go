package validator

import (
	"github.com/microcosm-cc/bluemonday"
)

// SanitizeXSS 使用严格策略过滤XSS攻击
func SanitizeXSS(input string) string {
	p := bluemonday.StrictPolicy() // 禁止所有HTML标签
	return p.Sanitize(input)
}

// SanitizeXSSRelaxed 允许安全HTML标签（需在业务明确需要时使用）
func SanitizeXSSRelaxed(input string) string {
	p := bluemonday.UGCPolicy() // 允许安全标签
	return p.Sanitize(input)
}
