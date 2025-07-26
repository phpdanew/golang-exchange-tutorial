package auth

import (
	"context"
	"errors"
	"regexp"
	"time"

	"crypto-exchange/internal/svc"
	"crypto-exchange/internal/types"
	"crypto-exchange/model"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.User, err error) {
	// 1. 验证邮箱格式
	if err := l.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// 2. 验证密码强度
	if err := l.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// 3. 检查用户是否已存在
	existingUser, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Errorf("Failed to check user existence: %v", err)
		return nil, model.ErrInternalServer
	}
	if existingUser != nil {
		return nil, model.ErrUserExists
	}

	// 4. 加密密码
	hashedPassword, err := l.hashPassword(req.Password)
	if err != nil {
		l.Errorf("Failed to hash password: %v", err)
		return nil, model.ErrInternalServer
	}

	// 5. 创建用户记录
	now := time.Now()
	user := &model.User{
		Email:     req.Email,
		Password:  hashedPassword,
		Nickname:  req.Nickname,
		Status:    1, // 1-正常状态
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, user)
	if err != nil {
		l.Errorf("Failed to insert user: %v", err)
		return nil, model.ErrInternalServer
	}

	// 6. 获取插入的用户ID
	userID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("Failed to get last insert id: %v", err)
		return nil, model.ErrInternalServer
	}
	user.ID = uint64(userID)

	// 7. 返回用户信息（不包含密码）
	resp = &types.User{
		ID:       user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
		Status:   user.Status,
	}

	l.Infof("User registered successfully: %s", req.Email)
	return resp, nil
}

// validateEmail 验证邮箱格式
func (l *RegisterLogic) validateEmail(email string) error {
	if email == "" {
		return model.ErrInvalidEmail
	}

	// 使用正则表达式验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return model.ErrInvalidEmail
	}

	return nil
}

// validatePassword 验证密码强度
func (l *RegisterLogic) validatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	// 检查密码是否包含至少一个字母和一个数字
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasNumber {
		return errors.New("password must contain at least one letter and one number")
	}

	return nil
}

// hashPassword 使用bcrypt加密密码
func (l *RegisterLogic) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
