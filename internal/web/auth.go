package web

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"moebot-next/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// JWT secret 持久化在 app_meta(key) 下；避免把随机密钥写回 yaml。
	metaKeyJWTSecret = "auth.jwt_secret"
	jwtTokenTTL      = 7 * 24 * time.Hour
)

var (
	usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

	// jwtSecret 在 ensureJWTSecret() 中懒加载（首次调用），后续读写并发安全。
	jwtSecretMu sync.RWMutex
	jwtSecret   []byte
)

// ensureJWTSecret 确保进程内有可用的 JWT 签名密钥。
//
// 优先读取 config.Web.Auth.JWTSecret（兼容老配置），其次落库的 app_meta，
// 都缺失时生成随机 32 字节密钥并写入 app_meta（不回写 yaml，避免破坏注释）。
func (s *Server) ensureJWTSecret() ([]byte, error) {
	jwtSecretMu.RLock()
	if len(jwtSecret) > 0 {
		secret := jwtSecret
		jwtSecretMu.RUnlock()
		return secret, nil
	}
	jwtSecretMu.RUnlock()

	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()
	if len(jwtSecret) > 0 {
		return jwtSecret, nil
	}

	if cfgSecret := strings.TrimSpace(s.Config.Web.Auth.JWTSecret); cfgSecret != "" {
		jwtSecret = []byte(cfgSecret)
		return jwtSecret, nil
	}

	stored, err := s.DB.GetAppMeta(metaKeyJWTSecret)
	if err != nil {
		return nil, err
	}
	if stored != "" {
		jwtSecret = []byte(stored)
		return jwtSecret, nil
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	hexSecret := hex.EncodeToString(buf)
	if err := s.DB.SetAppMeta(metaKeyJWTSecret, hexSecret); err != nil {
		return nil, err
	}
	jwtSecret = []byte(hexSecret)
	log.Info().Msg("Generated new JWT secret and persisted to app_meta")
	return jwtSecret, nil
}

func (s *Server) signToken(user *models.AdminUser) (string, error) {
	secret, err := s.ensureJWTSecret()
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"sub":      user.Username,
		"uid":      user.ID,
		"nickname": user.Nickname,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(jwtTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *Server) parseToken(raw string) (*models.AdminUser, error) {
	secret, err := s.ensureJWTSecret()
	if err != nil {
		return nil, err
	}
	token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	username, _ := claims["sub"].(string)
	if username == "" {
		return nil, errors.New("token missing subject")
	}
	user, err := s.DB.GetAdminUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// publicAPIPaths 是不需要登录即可访问的 /api 路径白名单。
//
// 这些接口要么用于鉴权流程本身（status/login/setup），要么用于公开元信息
// （deployer），要么用于探活（health）。其余 /api 路径一律需要 Bearer token。
var publicAPIPaths = map[string]struct{}{
	"/api/auth/status": {},
	"/api/auth/login":  {},
	"/api/setup":       {},
	"/api/deployer":    {},
	"/api/health":      {},
}

// authMiddleware Fiber 中间件：除白名单外，要求请求带 Authorization: Bearer <jwt>。
//
// 若数据库尚无任何管理员（首启状态），所有受保护接口一律返回 409，
// 前端守卫会引导去 /setup。挂在 `s.App` 路径前缀 `/api` 上，因此插件通过
// `webServer.App.Group("/api")` 注册的子路由也会被覆盖。
func (s *Server) authMiddleware(c *fiber.Ctx) error {
	if _, ok := publicAPIPaths[c.Path()]; ok {
		return c.Next()
	}
	// 放行 CORS preflight，由 cors 中间件处理。
	if c.Method() == fiber.MethodOptions {
		return c.Next()
	}
	count, err := s.DB.CountAdminUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query admin users")
	}
	if count == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   true,
			"code":    "setup_required",
			"message": "请先完成控制台初始化",
		})
	}

	header := c.Get(fiber.HeaderAuthorization)
	const prefix = "Bearer "
	var token string
	if strings.HasPrefix(header, prefix) {
		token = strings.TrimSpace(strings.TrimPrefix(header, prefix))
	} else {
		// 兼容浏览器 <img>/<a download> 等无法自定义请求头的场景：
		// 允许通过 ?token=<jwt> 查询参数传入。仅用于 GET 静态资源类
		// 端点（如 /api/plugins/gallery/pics/:pid/image）。
		token = strings.TrimSpace(c.Query("token"))
	}
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Missing bearer token")
	}
	user, err := s.parseToken(token)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
	}
	c.Locals("admin_user", user)
	return c.Next()
}

func currentAdmin(c *fiber.Ctx) *models.AdminUser {
	if v := c.Locals("admin_user"); v != nil {
		if u, ok := v.(*models.AdminUser); ok {
			return u
		}
	}
	return nil
}

// --- 校验工具 ---

func validateUsername(s string) error {
	if !usernamePattern.MatchString(s) {
		return errors.New("用户名需为 3–32 位英文/数字/下划线")
	}
	return nil
}

func validateNickname(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("昵称不能为空")
	}
	if utf8.RuneCountInString(s) > 32 {
		return errors.New("昵称长度不能超过 32 字符")
	}
	return nil
}

func validatePassword(s string) error {
	if len(s) < 8 {
		return errors.New("密码至少 8 位")
	}
	if len(s) > 128 {
		return errors.New("密码长度不能超过 128 位")
	}
	return nil
}

// --- HTTP Handlers ---

// handleAuthStatus 公开接口：返回控制台是否已初始化以及（如已初始化）昵称。
//
// 前端守卫据此决定跳 /setup 还是 /login，无需鉴权。
func (s *Server) handleAuthStatus(c *fiber.Ctx) error {
	count, err := s.DB.CountAdminUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query admin users")
	}
	out := fiber.Map{"initialized": count > 0}
	if count > 0 {
		if u, err := s.DB.GetAdminUser(); err == nil {
			out["nickname"] = u.Nickname
			out["username"] = u.Username
		}
	}
	return c.JSON(out)
}

// handleSetup 首启创建管理员账号。仅当 admin_users 为空时允许。
type setupRequest struct {
	Username        string `json:"username"`
	Nickname        string `json:"nickname"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

func (s *Server) handleSetup(c *fiber.Ctx) error {
	count, err := s.DB.CountAdminUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query admin users")
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "控制台已初始化")
	}
	var req setupRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON body")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Nickname = strings.TrimSpace(req.Nickname)
	if err := validateUsername(req.Username); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := validateNickname(req.Nickname); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := validatePassword(req.Password); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if req.Password != req.PasswordConfirm {
		return fiber.NewError(fiber.StatusBadRequest, "两次输入的密码不一致")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}
	user := &models.AdminUser{
		Username:     req.Username,
		Nickname:     req.Nickname,
		PasswordHash: string(hash),
	}
	if err := s.DB.CreateAdminUser(user); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create admin user")
	}

	// 同步给渲染服务，立刻让所有渲染 footer 显示 deployer。
	if s.Renderer != nil {
		if err := s.Renderer.SetDeployer(user.Nickname); err != nil {
			log.Warn().Err(err).Msg("Failed to push deployer nickname to renderer")
		}
	}

	token, err := s.signToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to sign token")
	}
	return c.JSON(fiber.Map{
		"token":    token,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON body")
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "账号与密码不能为空")
	}
	user, err := s.DB.GetAdminUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "账号或密码错误")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query admin user")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "账号或密码错误")
	}
	token, err := s.signToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to sign token")
	}
	return c.JSON(fiber.Map{
		"token":    token,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

func (s *Server) handleAuthMe(c *fiber.Ctx) error {
	u := currentAdmin(c)
	if u == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Not authenticated")
	}
	return c.JSON(fiber.Map{
		"username": u.Username,
		"nickname": u.Nickname,
	})
}

type changePasswordRequest struct {
	OldPassword        string `json:"old_password"`
	NewPassword        string `json:"new_password"`
	NewPasswordConfirm string `json:"new_password_confirm"`
}

func (s *Server) handleChangePassword(c *fiber.Ctx) error {
	u := currentAdmin(c)
	if u == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Not authenticated")
	}
	var req changePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON body")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "旧密码不正确")
	}
	if err := validatePassword(req.NewPassword); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if req.NewPassword != req.NewPasswordConfirm {
		return fiber.NewError(fiber.StatusBadRequest, "两次输入的新密码不一致")
	}
	if req.NewPassword == req.OldPassword {
		return fiber.NewError(fiber.StatusBadRequest, "新密码不能与旧密码相同")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to hash password")
	}
	if err := s.DB.UpdateAdminPassword(u.ID, string(hash)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update password")
	}
	return c.JSON(fiber.Map{"message": "ok"})
}

// handleDeployer 公开接口：返回 deployer 昵称（用于前端控制台底部以及兜底渲染补偿）。
func (s *Server) handleDeployer(c *fiber.Ctx) error {
	count, err := s.DB.CountAdminUsers()
	if err != nil || count == 0 {
		return c.JSON(fiber.Map{"nickname": ""})
	}
	u, err := s.DB.GetAdminUser()
	if err != nil {
		return c.JSON(fiber.Map{"nickname": ""})
	}
	return c.JSON(fiber.Map{"nickname": u.Nickname})
}
