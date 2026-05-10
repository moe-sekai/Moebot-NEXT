package web

import (
	"net/http"
	"testing"

	"moebot-next/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// mustSeedAdmin 在测试用 DB 中创建一个固定的管理员账号，避免 authMiddleware
// 在 admin_users 表为空时返回 409。后续登录或带 Bearer token 即可放行。
func mustSeedAdmin(t *testing.T, server *Server) *models.AdminUser {
	t.Helper()
	if existing, err := server.DB.GetAdminUser(); err == nil && existing != nil {
		return existing
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("test-pass-1234"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := &models.AdminUser{
		Username:     "tester",
		Nickname:     "Tester",
		PasswordHash: string(hash),
	}
	if err := server.DB.CreateAdminUser(user); err != nil {
		t.Fatal(err)
	}
	return user
}

// mustAuthorizeRequest 给请求附上一个 valid Bearer token，
// 调用前必须确保 server.DB 中至少存在一个 AdminUser（mustSeedAdmin）。
func mustAuthorizeRequest(t *testing.T, server *Server, req *http.Request) {
	t.Helper()
	user := mustSeedAdmin(t, server)
	tok, err := server.signToken(user)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
}
