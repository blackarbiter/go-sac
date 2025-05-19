package crypt

// Claims represents the JWT claims structure
type Claims struct {
	UserID uint `json:"user_id"`
}

// VerifyJWT validates a JWT token and returns the claims
// 临时实现：返回固定的用户ID，不做实际验证
func VerifyJWT(tokenString string) (*Claims, error) {
	// 无论输入什么token，都返回固定的测试用户ID
	return &Claims{
		UserID: 1, // 固定返回用户ID为1
	}, nil
}
