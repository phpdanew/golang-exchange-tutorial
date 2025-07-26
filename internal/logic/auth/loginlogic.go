package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// 1. 检查登录失败次数限制
	if err := l.checkLoginAttempts(req.Email); err != nil {
		return nil, err
	}

	// 2. 验证邮箱格式
	if err := l.validateEmail(req.Email); err != nil {
		l.recordFailedLogin(req.Email)
		return nil, err
	}

	// 3. 查找用户
	user, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.recordFailedLogin(req.Email)
			return nil, model.ErrUserNotFound
		}
		l.Errorf("Failed to find user by email: %v", err)
		return nil, model.ErrInternalServer
	}

	// 4. 检查用户状态
	if user.Status != 1 {
		l.recordFailedLogin(req.Email)
		return nil, model.ErrUserDisabled
	}

	// 5. 验证密码
	if err := l.verifyPassword(req.Password, user.Password); err != nil {
		l.recordFailedLogin(req.Email)
		return nil, model.ErrInvalidPassword
	}

	// 6. 生成JWT token
	token, err := l.generateJWTToken(user)
	if err != nil {
		l.Errorf("Failed to generate JWT token: %v", err)
		return nil, model.ErrInternalServer
	}

	// 7. 清除失败登录记录
	l.clearFailedLogin(req.Email)

	// 8. 返回登录响应
	resp = &types.LoginResponse{
		Token: token,
		User: types.User{
			ID:       user.ID,
			Email:    user.Email,
			Nickname: user.Nickname,
			Status:   user.Status,
		},
	}

	l.Infof("User logged in successfully: %s", req.Email)
	return resp, nil
}

// validateEmail 验证邮箱格式（复用注册逻辑中的验证）
func (l *LoginLogic) validateEmail(email string) error {
	if email == "" {
		return model.ErrInvalidEmail
	}
	// 这里可以添加更详细的邮箱格式验证，但为了简化，只检查非空
	return nil
}

// verifyPassword 验证密码
func (l *LoginLogic) verifyPassword(plainPassword, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return model.ErrInvalidPassword
	}
	return nil
}

// generateJWTToken 生成JWT token
func (l *LoginLogic) generateJWTToken(user *model.User) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second)

	claims := jwt.MapClaims{
		"iat":    now.Unix(),
		"exp":    expire.Unix(),
		"userId": user.ID,
		"email":  user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// checkLoginAttempts 检查登录失败次数限制
func (l *LoginLogic) checkLoginAttempts(email string) error {
	key := fmt.Sprintf("login_attempts:%s", email)
	
	// 获取失败次数
	attemptsStr, err := l.svcCtx.RedisClient.Get(key)
	if err != nil {
		// Redis错误不应该阻止登录，只记录日志
		l.Errorf("Failed to get login attempts from Redis: %v", err)
		return nil
	}

	if attemptsStr == "" {
		return nil
	}

	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		l.Errorf("Failed to parse login attempts: %v", err)
		return nil
	}

	// 如果失败次数超过5次，则限制登录
	if attempts >= 5 {
		return errors.New("too many failed login attempts, please try again later")
	}

	return nil
}

// recordFailedLogin 记录失败登录
func (l *LoginLogic) recordFailedLogin(email string) {
	key := fmt.Sprintf("login_attempts:%s", email)
	
	// 增加失败次数
	_, err := l.svcCtx.RedisClient.Incr(key)
	if err != nil {
		l.Errorf("Failed to increment login attempts: %v", err)
		return
	}

	// 设置过期时间为15分钟
	err = l.svcCtx.RedisClient.Expire(key, 15*60)
	if err != nil {
		l.Errorf("Failed to set expiration for login attempts: %v", err)
	}
}

// clearFailedLogin 清除失败登录记录
func (l *LoginLogic) clearFailedLogin(email string) {
	key := fmt.Sprintf("login_attempts:%s", email)
	
	_, err := l.svcCtx.RedisClient.Del(key)
	if err != nil {
		l.Errorf("Failed to clear login attempts: %v", err)
	}
}
