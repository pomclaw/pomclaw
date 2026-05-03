// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pomclaw/pomclaw/internal/store"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"golang.org/x/crypto/bcrypt"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// User registration
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.AuthResp, err error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("username, email, and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := store.CreateUser(l.svcCtx.Conn.DB(), req.Username, req.Email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("username or email already exists")
	}

	// Generate JWT token using go-zero's approach
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessToken, err := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, accessExpire, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &types.AuthResp{
		AccessToken:  accessToken,
		RefreshToken: "", // TODO: implement refresh token if needed
		ExpiresIn:    accessExpire,
		TokenType:    "Bearer",
	}, nil
}

func (l *RegisterLogic) getJwtToken(secretKey string, seconds int64, userId string) (string, error) {
	now := time.Now().Unix()
	claims := make(jwt.MapClaims)
	claims["exp"] = now + seconds
	claims["iat"] = now
	claims["userId"] = userId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
