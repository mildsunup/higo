// Package auth 提供认证抽象。
//
// 核心功能：
//   - JWT 令牌生成和验证
//   - Bcrypt 密码哈希
//   - 简单认证器接口
//
// 使用示例：
//
//	jwt := auth.NewJWT(secret, expiry)
//	token, err := jwt.Generate(claims)
//	claims, err := jwt.Verify(token)
package auth
