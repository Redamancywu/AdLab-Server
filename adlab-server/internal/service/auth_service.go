package service

import (
	"errors"
	"log/slog"
	"time"

	"adlab-server/internal/config"
	"adlab-server/internal/model"
	"adlab-server/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ─── JWT Claims ───────────────────────────────────────────────────────────────

// JWTClaims 自定义 JWT 载荷
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	TenantID uint   `json:"tenant_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ─── Auth Service ─────────────────────────────────────────────────────────────

// AuthService 认证服务：登录、JWT 签发与校验
type AuthService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	jwtCfg     config.JWTConfig
}

// NewAuthService 创建 AuthService
func NewAuthService(
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	jwtCfg config.JWTConfig,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		jwtCfg:     jwtCfg,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      *UserInfo  `json:"user"`
}

// UserInfo 返回给前端的用户摘要信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID uint   `json:"tenant_id"`
}

// Login 用户名密码登录，成功返回 JWT Token
func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("查询用户失败")
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}
	if user.Status != "active" {
		return nil, errors.New("账户已被禁用")
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成 Token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	// 异步更新最后登录时间（不阻塞响应）
	go func() {
		if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
			slog.Warn("更新最后登录时间失败", "user_id", user.ID, "error", err)
		}
	}()

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			TenantID: user.TenantID,
		},
	}, nil
}

// ValidateToken 校验并解析 JWT Token
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 Token")
	}
	return claims, nil
}

// RefreshToken 刷新 Token（用户仍在有效期内时续期）
func (s *AuthService) RefreshToken(tokenString string) (*LoginResponse, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, errors.New("Token 无效，请重新登录")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}

	newToken, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("刷新 Token 失败")
	}

	return &LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			TenantID: user.TenantID,
		},
	}, nil
}

// HashPassword 对密码进行 bcrypt 哈希（创建用户时使用）
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// EnsureDefaultAdmin 确保系统中有至少一个管理员账户（首次启动时初始化）
func (s *AuthService) EnsureDefaultAdmin(tenantRepo *repository.TenantRepository) {
	// 检查是否已有租户
	tenants, err := tenantRepo.List()
	if err != nil || len(tenants) > 0 {
		return
	}

	slog.Info("首次启动：初始化默认租户和管理员账号")

	// 创建默认租户
	tenant := &model.Tenant{
		Name:      "Default",
		AppKey:    "default-app-key",
		AppSecret: "default-app-secret-change-me",
		Status:    "active",
	}
	if err := tenantRepo.Create(tenant); err != nil {
		slog.Error("创建默认租户失败", "error", err)
		return
	}

	// 创建默认管理员（密码: admin123，首次登录后请立即修改）
	passwordHash, err := HashPassword("admin123")
	if err != nil {
		slog.Error("生成默认密码哈希失败", "error", err)
		return
	}

	user := &model.User{
		TenantID:     tenant.ID,
		Username:     "admin",
		PasswordHash: passwordHash,
		Role:         "superadmin",
		Status:       "active",
	}
	if err := s.userRepo.Create(user); err != nil {
		slog.Error("创建默认管理员失败", "error", err)
		return
	}

	slog.Info("默认管理员账号已创建", "username", "admin", "password", "admin123", "tip", "请立即修改密码")
}

// generateToken 内部方法：生成 JWT Token
func (s *AuthService) generateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.jwtCfg.ExpireHour) * time.Hour)
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "adlab-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}
